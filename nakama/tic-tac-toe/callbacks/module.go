package callbacks

import (
	"context"
	"database/sql"
	"fmt"

	"lila/nakama/tic-tac-toe/game"

	"github.com/heroiclabs/nakama-common/runtime"
)

func RegisterTicTacToeModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := game.EnsureGlobalLeaderboardExists(ctx, logger, nk); err != nil {
		return err
	}

	registrations := []struct {
		name string
		fn   func() error
	}{
		{"match handler", func() error {
			return initializer.RegisterMatch(game.MatchModuleName, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
				return &TicTacToeMatchHandler{}, nil
			})
		}},
		{"matchmaker matched", func() error { return initializer.RegisterMatchmakerMatched(handleMatchmakerMatched) }},
		{"rpc create_room", func() error { return initializer.RegisterRpc("create_room", handleCreateRoomRPC) }},
		{"rpc list_rooms", func() error { return initializer.RegisterRpc("list_rooms", handleListRoomsRPC) }},
		{"rpc join_private_room", func() error { return initializer.RegisterRpc("join_private_room", handleJoinPrivateRoomRPC) }},
		{"rpc get_player_profile", func() error { return initializer.RegisterRpc("get_player_profile", handleGetPlayerProfileRPC) }},
		{"rpc get_online_player_count", func() error { return initializer.RegisterRpc("get_online_player_count", handleGetOnlinePlayerCountRPC) }},
		{"session start", func() error { return initializer.RegisterEventSessionStart(handleSessionStart) }},
		{"session end", func() error { return initializer.RegisterEventSessionEnd(handleSessionEnd) }},
	}

	for _, registration := range registrations {
		if err := registration.fn(); err != nil {
			return fmt.Errorf("register %s: %w", registration.name, err)
		}
	}

	logger.Info("LILA Nakama module loaded.")
	return nil
}
