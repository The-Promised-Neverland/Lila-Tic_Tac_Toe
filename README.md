# LILA Backend

This repo is structured so Nakama-specific code lives under `nakama/`, with a clear split between Nakama callback entrypoints and our internal game logic.

## Current structure

- `nakama/main.go`: Nakama plugin entrypoint
- `nakama/tic-tac-toe/callbacks/`: functions Nakama calls
- `nakama/tic-tac-toe/game/`: functions and types we use internally
- `docker-compose.yml`: production-style container wiring driven by `.env`

## Included in the current scaffold

- Server-authoritative Tic-Tac-Toe match loop
- Server-side move validation
- Public room creation and discovery
- Private room creation with invite codes
- Matchmaker callback for authoritative match creation
- 15-second reconnect grace window
- Timed-mode forfeits
- Player stats persistence
- Last-50 match history persistence
- Leaderboard writes

## Next likely additions

- `nakama/build/` for plugin build output
- `deploy/` or `infra/` for Docker/cloud setup
- `clients/` or `frontend/` when the UI lands
- secret rotation and environment-specific `.env` files
