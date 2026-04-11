package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func EnsureGlobalLeaderboardExists(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	err := nk.LeaderboardCreate(ctx, LeaderboardID, true, "desc", "best", "", map[string]interface{}{
		"game": "tic_tac_toe",
	}, true)
	if err == nil {
		return nil
	}

	logger.Info("leaderboard create skipped or already exists: %v", err)
	return nil
}

func GetContextString(ctx context.Context, key interface{}) (string, error) {
	value := ctx.Value(key)
	if value == nil {
		return "", errors.New("missing runtime context value")
	}

	str, ok := value.(string)
	if !ok || str == "" {
		return "", errors.New("invalid runtime context value")
	}

	return str, nil
}

func MustJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal response: %w", err)
	}
	return string(data), nil
}

func (s *MatchState) HandleMove(logger runtime.Logger, dispatcher runtime.MatchDispatcher, message runtime.MatchData, tick int64) error {
	if s.Status != "playing" {
		return errors.New("match is not active")
	}

	player, ok := s.Players[message.GetUserId()]
	if !ok {
		return errors.New("player not part of match")
	}
	if !player.Connected {
		return errors.New("player is disconnected")
	}
	if s.CurrentTurnUserID != player.UserID {
		return errors.New("not your turn")
	}

	var move MoveRequest
	if err := json.Unmarshal(message.GetData(), &move); err != nil {
		return errors.New("invalid move payload")
	}
	if move.Index < 0 || move.Index >= BoardSize {
		return errors.New("move out of range")
	}
	if s.Board[move.Index] != "" {
		return errors.New("cell already occupied")
	}

	s.Board[move.Index] = player.Mark
	s.MoveCount++

	if winningLine, won := DetectWinner(s.Board, player.Mark); won {
		s.Status = "finished"
		s.WinnerUserID = player.UserID
		s.WinnerMark = player.Mark
		s.WinningLine = winningLine
		s.TurnDeadlineTick = 0
		s.ResultRecorded = true
		s.BroadcastState(logger, dispatcher, nil)
		s.SyncLabel(logger, dispatcher)
		return nil
	}

	if s.MoveCount == BoardSize {
		s.Status = "draw"
		s.TurnDeadlineTick = 0
		s.ResultRecorded = true
		s.BroadcastState(logger, dispatcher, nil)
		s.SyncLabel(logger, dispatcher)
		return nil
	}

	s.CurrentTurnUserID = s.OtherPlayerID(player.UserID)
	s.SetTurnDeadline(tick)
	s.BroadcastState(logger, dispatcher, nil)
	s.SyncLabel(logger, dispatcher)
	return nil
}

func (s *MatchState) HandleDisconnects(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64) {
	for _, userID := range s.PlayerOrder {
		player := s.Players[userID]
		if player == nil || player.Connected || player.DisconnectAt == 0 || tick < player.DisconnectAt {
			continue
		}

		if s.MoveCount < NoContestMoveThreshold {
			s.FinishNoContest(ctx, logger, nk, dispatcher)
			return
		}

		s.FinishForfeit(ctx, logger, nk, dispatcher, s.OtherPlayerID(player.UserID))
		return
	}
}

func (s *MatchState) HandleTurnTimeout(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64) {
	if !s.Timed || s.TurnDeadlineTick == 0 || tick < s.TurnDeadlineTick {
		return
	}

	s.FinishForfeit(ctx, logger, nk, dispatcher, s.OtherPlayerID(s.CurrentTurnUserID))
}

func (s *MatchState) FinishNoContest(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher) {
	if s.ResultRecorded {
		return
	}

	s.Status = "no_contest"
	s.CurrentTurnUserID = ""
	s.TurnDeadlineTick = 0
	s.ResultRecorded = true
	s.BroadcastState(logger, dispatcher, nil)
	s.SyncLabel(logger, dispatcher)

	for _, userID := range s.PlayerOrder {
		player := s.Players[userID]
		if player == nil {
			continue
		}

		stats, err := ReadPlayerStats(ctx, nk, player.UserID, player.Username)
		if err == nil {
			stats.GamesPlayed++
			stats.NoContests++
			stats.WinStreak = 0
			_ = WritePlayerStats(ctx, nk, stats)
		}

		_ = AppendMatchHistory(ctx, nk, player.UserID, MatchHistoryEntry{
			MatchID:   s.MatchID,
			PlayedAt:  time.Now().UTC(),
			Result:    "no_contest",
			NoContest: true,
			Timed:     s.Timed,
			MoveCount: s.MoveCount,
		})
	}
}

func (s *MatchState) FinishForfeit(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, winnerID string) {
	if s.ResultRecorded {
		return
	}

	s.Status = "forfeit"
	s.WinnerUserID = winnerID
	if winner := s.Players[winnerID]; winner != nil {
		s.WinnerMark = winner.Mark
	}
	s.CurrentTurnUserID = ""
	s.TurnDeadlineTick = 0
	s.ResultRecorded = true
	s.BroadcastState(logger, dispatcher, nil)
	s.SyncLabel(logger, dispatcher)
	_ = s.PersistMatchResult(ctx, nk)
}

func (s *MatchState) PersistMatchResult(ctx context.Context, nk runtime.NakamaModule) error {
	if len(s.PlayerOrder) != 2 {
		return nil
	}

	a := s.Players[s.PlayerOrder[0]]
	b := s.Players[s.PlayerOrder[1]]
	if a == nil || b == nil {
		return nil
	}

	now := time.Now().UTC()
	for _, current := range []*MatchPlayer{a, b} {
		opponent := a
		if current.UserID == a.UserID {
			opponent = b
		}

		stats, err := ReadPlayerStats(ctx, nk, current.UserID, current.Username)
		if err != nil {
			return err
		}
		stats.GamesPlayed++

		entry := MatchHistoryEntry{
			MatchID:      s.MatchID,
			PlayedAt:     now,
			OpponentID:   opponent.UserID,
			OpponentName: opponent.Username,
			Forfeit:      s.Status == "forfeit",
			Timed:        s.Timed,
			MoveCount:    s.MoveCount,
		}

		switch {
		case s.Status == "draw":
			stats.Draws++
			stats.WinStreak = 0
			entry.Result = "draw"
		case current.UserID == s.WinnerUserID:
			stats.Wins++
			stats.WinStreak++
			if stats.WinStreak > stats.BestStreak {
				stats.BestStreak = stats.WinStreak
			}
			entry.Result = "win"
		default:
			stats.Losses++
			stats.WinStreak = 0
			if s.Status == "forfeit" {
				entry.Result = "forfeit_loss"
			} else {
				entry.Result = "loss"
			}
		}

		if err := WritePlayerStats(ctx, nk, stats); err != nil {
			return err
		}
		if err := AppendMatchHistory(ctx, nk, current.UserID, entry); err != nil {
			return err
		}

		score := int64(stats.Wins*3 + stats.Draws)
		subscore := int64(stats.BestStreak)
		if _, err := nk.LeaderboardRecordWrite(ctx, LeaderboardID, current.UserID, current.Username, score, subscore, map[string]interface{}{
			"wins":        stats.Wins,
			"losses":      stats.Losses,
			"draws":       stats.Draws,
			"win_streak":  stats.WinStreak,
			"best_streak": stats.BestStreak,
		}, nil); err != nil {
			return err
		}
	}

	if s.Private {
		_ = DeleteInviteCode(ctx, nk, s.InviteCode)
	}
	return nil
}

func (s *MatchState) BroadcastState(logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.Presence) {
	data, err := MustJSON(s.Snapshot())
	if err != nil {
		logger.Error("marshal state failed: %v", err)
		return
	}

	if err := dispatcher.BroadcastMessage(OpCodeState, []byte(data), nil, sender, true); err != nil {
		logger.Error("broadcast state failed: %v", err)
	}
}

func (s *MatchState) Snapshot() GameStatePayload {
	payload := GameStatePayload{
		MatchID:            s.MatchID,
		RoomName:           s.RoomName,
		Private:            s.Private,
		InviteCode:         s.InviteCode,
		Board:              s.Board,
		Status:             s.Status,
		WinnerUserID:       s.WinnerUserID,
		WinnerMark:         s.WinnerMark,
		CurrentTurnUserID:  s.CurrentTurnUserID,
		WinningLine:        s.WinningLine,
		MoveCount:          s.MoveCount,
		Timed:              s.Timed,
		TurnTimeSeconds:    s.TurnTimeSeconds,
		ReconnectWindowSec: DisconnectGraceSeconds,
	}

	if s.TurnDeadlineTick > 0 {
		payload.TurnExpiresAt = time.Now().UTC().Add(time.Duration(SecondsUntilDeadline(s.LastTick, s.TurnDeadlineTick)) * time.Second).Unix()
	}

	for _, userID := range s.PlayerOrder {
		player := s.Players[userID]
		if player == nil {
			continue
		}
		payload.Players = append(payload.Players, PlayerState{
			UserID:        player.UserID,
			Username:      player.Username,
			Mark:          player.Mark,
			Connected:     player.Connected,
			IsCurrentTurn: player.UserID == s.CurrentTurnUserID,
		})
	}

	return payload
}

func (s *MatchState) Label() string {
	label, _ := MustJSON(MatchLabel{
		Version:         MatchLabelVersion,
		RoomName:        s.RoomName,
		Private:         s.Private,
		Timed:           s.Timed,
		TurnTimeSeconds: s.TurnTimeSeconds,
		InviteCode:      s.InviteCode,
		PlayerCount:     len(s.PlayerOrder),
		MaxPlayers:      MaxPlayers,
		Open:            len(s.PlayerOrder) < MaxPlayers && s.Status == "waiting",
		Status:          s.Status,
	})
	return label
}

func (s *MatchState) SyncLabel(logger runtime.Logger, dispatcher runtime.MatchDispatcher) {
	if err := dispatcher.MatchLabelUpdate(s.Label()); err != nil {
		logger.Warn("match label update failed: %v", err)
	}
}

func (s *MatchState) SetTurnDeadline(currentTick int64) {
	if !s.Timed {
		s.TurnDeadlineTick = 0
		return
	}
	s.TurnDeadlineTick = currentTick + int64(s.TurnTimeSeconds*DefaultTickRate)
}

func (s *MatchState) OtherPlayerID(userID string) string {
	for _, id := range s.PlayerOrder {
		if id != userID {
			return id
		}
	}
	return ""
}

func (s *MatchState) NoConnectedPlayers() bool {
	for _, player := range s.Players {
		if player != nil && player.Connected {
			return false
		}
	}
	return true
}

func NextMark(index int) string {
	if index == 0 {
		return "X"
	}
	return "O"
}

func DetectWinner(board [BoardSize]string, mark string) ([]int, bool) {
	for _, line := range winningLines {
		if board[line[0]] == mark && board[line[1]] == mark && board[line[2]] == mark {
			return []int{line[0], line[1], line[2]}, true
		}
	}
	return nil, false
}

func AsInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func SecondsUntilDeadline(currentTick, deadlineTick int64) int64 {
	if deadlineTick <= currentTick {
		return 0
	}
	ticksRemaining := deadlineTick - currentTick
	seconds := ticksRemaining / DefaultTickRate
	if ticksRemaining%DefaultTickRate != 0 {
		seconds++
	}
	return seconds
}
