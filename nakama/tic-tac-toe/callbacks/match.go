package callbacks

import (
	"context"
	"database/sql"

	"lila/nakama/tic-tac-toe/game"

	"github.com/heroiclabs/nakama-common/runtime"
)

type TicTacToeMatchHandler struct{}

func (m *TicTacToeMatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := game.NewMatchState(params)
	return state, game.DefaultTickRate, state.Label()
}

func (m *TicTacToeMatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	state := stateRaw.(*game.MatchState)
	state.LastTick = tick

	if state.MatchID == "" {
		if matchID, err := game.GetContextString(ctx, runtime.RUNTIME_CTX_MATCH_ID); err == nil {
			state.MatchID = matchID
			if state.Private && state.InviteCode != "" {
				if err := game.WriteInviteCode(ctx, nk, state.InviteCode, state.MatchID); err != nil {
					logger.Error("write invite code failed: %v", err)
				}
			}
		}
	}

	userID := presence.GetUserId()
	if existing, ok := state.Players[userID]; ok {
		existing.Connected = true
		existing.DisconnectAt = 0
		return state, true, ""
	}

	if len(state.PlayerOrder) >= game.MaxPlayers {
		return state, false, "match is full"
	}
	if state.Private && game.NormalizeInviteCode(metadata["invite_code"]) != state.InviteCode {
		return state, false, "invite code required"
	}

	return state, true, ""
}

func (m *TicTacToeMatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, presences []runtime.Presence) interface{} {
	state := stateRaw.(*game.MatchState)
	state.LastTick = tick

	for _, presence := range presences {
		player, exists := state.Players[presence.GetUserId()]
		if !exists {
			player = &game.MatchPlayer{
				UserID:   presence.GetUserId(),
				Username: presence.GetUsername(),
				Mark:     game.NextMark(len(state.PlayerOrder)),
			}
			state.Players[player.UserID] = player
			state.PlayerOrder = append(state.PlayerOrder, player.UserID)
		}

		player.Username = presence.GetUsername()
		player.Connected = true
		player.DisconnectAt = 0
	}

	if len(state.PlayerOrder) == game.MaxPlayers && state.Status == "waiting" {
		state.Status = "playing"
		state.CurrentTurnUserID = state.PlayerOrder[0]
		state.SetTurnDeadline(tick)
	}

	state.SyncLabel(logger, dispatcher)
	state.BroadcastState(logger, dispatcher, nil)
	return state
}

func (m *TicTacToeMatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, presences []runtime.Presence) interface{} {
	state := stateRaw.(*game.MatchState)
	state.LastTick = tick

	for _, presence := range presences {
		if player, ok := state.Players[presence.GetUserId()]; ok {
			player.Connected = false
			player.DisconnectAt = tick + int64(game.DisconnectGraceSeconds*game.DefaultTickRate)
		}
	}

	state.SyncLabel(logger, dispatcher)
	state.BroadcastState(logger, dispatcher, nil)
	return state
}

func (m *TicTacToeMatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, messages []runtime.MatchData) interface{} {
	state := stateRaw.(*game.MatchState)
	state.LastTick = tick

	for _, message := range messages {
		if message.GetOpCode() != game.OpCodeMove {
			logger.Warn("unknown opcode %d", message.GetOpCode())
			continue
		}

		if err := state.HandleMove(ctx, logger, nk, dispatcher, message, tick); err != nil {
			logger.Warn("move rejected for user %s: %v", message.GetUserId(), err)
		}
	}

	if state.Status == "playing" {
		state.HandleDisconnects(ctx, logger, nk, dispatcher, tick)
		state.HandleTurnTimeout(ctx, logger, nk, dispatcher, tick)
	}

	if (state.Status == "finished" || state.Status == "draw" || state.Status == "forfeit") && state.ResultRecorded && !state.ResultPersisted {
		if err := state.PersistMatchResult(ctx, nk); err != nil {
			logger.Error("retry persist match result failed: %v", err)
		} else {
			state.ResultPersisted = true
		}
	}

	if state.Status == "waiting" && state.NoConnectedPlayers() {
		state.EmptyTicks++
		if state.EmptyTicks > 60*game.DefaultTickRate {
			if state.Private {
				_ = game.DeleteInviteCode(ctx, nk, state.InviteCode)
			}
			return nil
		}
	} else {
		state.EmptyTicks = 0
	}

	return state
}

func (m *TicTacToeMatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, graceSeconds int) interface{} {
	state := stateRaw.(*game.MatchState)
	if state.Private {
		_ = game.DeleteInviteCode(ctx, nk, state.InviteCode)
	}
	return state
}

func (m *TicTacToeMatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, stateRaw interface{}, data string) (interface{}, string) {
	state := stateRaw.(*game.MatchState)
	return state, data
}
