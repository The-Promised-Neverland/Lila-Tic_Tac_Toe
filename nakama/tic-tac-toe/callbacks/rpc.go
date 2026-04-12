package callbacks

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"lila/nakama/tic-tac-toe/game"

	"github.com/heroiclabs/nakama-common/runtime"
)

func handleCreateRoomRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, err := game.GetContextString(ctx, runtime.RUNTIME_CTX_USER_ID)
	if err != nil {
		return "", err
	}

	var req game.RoomCreateRequest
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return "", errors.New("invalid create_room payload")
		}
	}

	req.RoomName = strings.TrimSpace(req.RoomName)
	if req.RoomName == "" {
		req.RoomName = game.DefaultRoomName
	}
	if req.TurnTimeSeconds <= 0 {
		req.TurnTimeSeconds = game.TimedModeTurnSeconds
	}

	params := map[string]interface{}{
		"creator_user_id":   userID,
		"room_name":         req.RoomName,
		"private":           req.Private,
		"timed":             req.Timed,
		"turn_time_seconds": req.TurnTimeSeconds,
	}

	if req.Private {
		code, err := game.GenerateInviteCode()
		if err != nil {
			return "", err
		}
		params["invite_code"] = code
	}

	matchID, err := nk.MatchCreate(ctx, game.MatchModuleName, params)
	if err != nil {
		logger.Error("create room failed: %v", err)
		return "", errors.New("could not create room")
	}

	if req.Private {
		if inviteCode, ok := params["invite_code"].(string); ok && inviteCode != "" {
			if err := game.WriteInviteCode(ctx, nk, inviteCode, matchID); err != nil {
				logger.Error("write invite code on create failed: %v", err)
				return "", errors.New("could not create invite code")
			}
		}
	}

	resp := game.RoomCreateResponse{
		MatchID:         matchID,
		Private:         req.Private,
		Timed:           req.Timed,
		TurnTimeSeconds: req.TurnTimeSeconds,
	}
	if inviteCode, ok := params["invite_code"].(string); ok {
		resp.InviteCode = inviteCode
	}

	return game.MustJSON(resp)
}

func handleListRoomsRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	matches, err := nk.MatchList(ctx, game.PublicRoomLimit, true, "", nil, nil, "")
	if err != nil {
		logger.Error("list rooms failed: %v", err)
		return "", errors.New("could not list rooms")
	}

	resp := game.RoomListResponse{
		Rooms: make([]game.RoomSummary, 0, len(matches)),
	}

	for _, match := range matches {
		if match == nil || match.Label == nil {
			continue
		}

		var label game.MatchLabel
		if err := json.Unmarshal([]byte(match.Label.Value), &label); err != nil {
			continue
		}
		if label.Private || label.Status != game.RoomStatusWaiting || !label.Open {
			continue
		}

		resp.Rooms = append(resp.Rooms, game.RoomSummary{
			MatchID:         match.MatchId,
			RoomName:        label.RoomName,
			Private:         label.Private,
			Timed:           label.Timed,
			TurnTimeSeconds: label.TurnTimeSeconds,
			PlayerCount:     label.PlayerCount,
			MaxPlayers:      label.MaxPlayers,
			Open:            label.Open,
			Status:          label.Status,
		})
	}

	return game.MustJSON(resp)
}

func handleJoinPrivateRoomRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var req game.JoinPrivateRoomRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", errors.New("invalid join_private_room payload")
	}

	req.InviteCode = game.NormalizeInviteCode(req.InviteCode)
	if req.InviteCode == "" {
		return "", errors.New("invite_code is required")
	}

	matchID, err := game.LookupMatchIDByInviteCode(ctx, nk, req.InviteCode)
	if err != nil {
		return "", err
	}

	return game.MustJSON(game.JoinPrivateRoomResponse{
		MatchID:    matchID,
		InviteCode: req.InviteCode,
	})
}

func handleGetPlayerProfileRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, err := game.GetContextString(ctx, runtime.RUNTIME_CTX_USER_ID)
	if err != nil {
		return "", err
	}

	username, _ := game.GetContextString(ctx, runtime.RUNTIME_CTX_USERNAME)
	stats, err := game.ReadPlayerStats(ctx, nk, userID, username)
	if err != nil {
		return "", err
	}
	history, err := game.ReadMatchHistory(ctx, nk, userID)
	if err != nil {
		return "", err
	}

	return game.MustJSON(game.PlayerProfileResponse{
		Stats:   stats,
		History: history,
	})
}

func handleMatchmakerMatched(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	params := map[string]interface{}{
		"room_name":         game.DefaultMatchmadeRoomName,
		"private":           false,
		"timed":             false,
		"turn_time_seconds": game.TimedModeTurnSeconds,
		"source":            "matchmaker",
	}

	if len(entries) > 0 {
		properties := entries[0].GetProperties()
		if timed, ok := properties["timed"].(bool); ok {
			params["timed"] = timed
		}
		if roomName, ok := properties["room_name"].(string); ok && strings.TrimSpace(roomName) != "" {
			params["room_name"] = roomName
		}
		if seconds, ok := game.AsInt(properties["turn_time_seconds"]); ok && seconds > 0 {
			params["turn_time_seconds"] = seconds
		}
	}

	matchID, err := nk.MatchCreate(ctx, game.MatchModuleName, params)
	if err != nil {
		logger.Error("matchmaker create failed: %v", err)
		return "", err
	}

	return matchID, nil
}
