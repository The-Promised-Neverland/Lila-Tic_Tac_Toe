package callbacks

import (
	"context"
	"database/sql"
	"sync"

	"lila/nakama/tic-tac-toe/game"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type onlinePresenceTracker struct {
	mu             sync.Mutex
	sessionsByUser map[string]map[string]struct{}
}

var onlinePlayers = &onlinePresenceTracker{
	sessionsByUser: make(map[string]map[string]struct{}),
}

func (t *onlinePresenceTracker) add(userID, sessionID string) {
	if userID == "" || sessionID == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	sessions := t.sessionsByUser[userID]
	if sessions == nil {
		sessions = make(map[string]struct{})
		t.sessionsByUser[userID] = sessions
	}
	sessions[sessionID] = struct{}{}
}

func (t *onlinePresenceTracker) remove(userID, sessionID string) {
	if userID == "" || sessionID == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	sessions := t.sessionsByUser[userID]
	if sessions == nil {
		return
	}

	delete(sessions, sessionID)
	if len(sessions) == 0 {
		delete(t.sessionsByUser, userID)
	}
}

func (t *onlinePresenceTracker) count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.sessionsByUser)
}

func handleSessionStart(ctx context.Context, logger runtime.Logger, evt *api.Event) {
	onlinePlayers.add(
		getCtxString(ctx, runtime.RUNTIME_CTX_USER_ID),
		getCtxString(ctx, runtime.RUNTIME_CTX_SESSION_ID),
	)
}

func handleSessionEnd(ctx context.Context, logger runtime.Logger, evt *api.Event) {
	onlinePlayers.remove(
		getCtxString(ctx, runtime.RUNTIME_CTX_USER_ID),
		getCtxString(ctx, runtime.RUNTIME_CTX_SESSION_ID),
	)
}

func getCtxString(ctx context.Context, key interface{}) string {
	value, _ := ctx.Value(key).(string)
	return value
}

func handleGetOnlinePlayerCountRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return game.MustJSON(game.OnlinePlayerCountResponse{
		Count: onlinePlayers.count(),
	})
}
