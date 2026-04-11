package main

import (
	"context"
	"database/sql"

	"lila/nakama/tic-tac-toe/callbacks"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	return callbacks.RegisterTicTacToeModule(ctx, logger, db, nk, initializer)
}

func main() {}
