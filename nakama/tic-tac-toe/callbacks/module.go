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
		{RegMatchHandler, func() error {
			return initializer.RegisterMatch(game.MatchModuleName, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
				return &TicTacToeMatchHandler{}, nil
			})
		}},
		{RegMatchmakerMatched, func() error { return initializer.RegisterMatchmakerMatched(handleMatchmakerMatched) }},
		{RegRPCCreateRoom, func() error { return initializer.RegisterRpc(RPCCreateRoom, handleCreateRoomRPC) }},
		{RegRPCListRooms, func() error { return initializer.RegisterRpc(RPCListRooms, handleListRoomsRPC) }},
		{RegRPCJoinPrivateRoom, func() error { return initializer.RegisterRpc(RPCJoinPrivateRoom, handleJoinPrivateRoomRPC) }},
		{RegRPCGetPlayerProfile, func() error { return initializer.RegisterRpc(RPCGetPlayerProfile, handleGetPlayerProfileRPC) }},
		{RegRPCGetOnlineCount, func() error { return initializer.RegisterRpc(RPCGetOnlinePlayerCount, handleGetOnlinePlayerCountRPC) }},
		{RegSessionStart, func() error { return initializer.RegisterEventSessionStart(handleSessionStart) }},
		{RegSessionEnd, func() error { return initializer.RegisterEventSessionEnd(handleSessionEnd) }},
	}

	for _, registration := range registrations {
		if err := registration.fn(); err != nil {
			return fmt.Errorf("register %s: %w", registration.name, err)
		}
	}

	logger.Info("LILA Nakama module loaded.")
	return nil
}
