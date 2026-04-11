# Local Run

This repo can now be run locally with Docker using:

- Nakama `3.38.0`
- PostgreSQL `15`
- Go runtime plugin built into the Nakama image

## Why this version

Heroic Labs release notes currently state that:

- Nakama `3.38.0` requires `nakama-common v1.45.0`

The project already uses `github.com/heroiclabs/nakama-common v1.45.0`, so the Docker setup is aligned to that version.

## Files

- [docker-compose.yml](/c:/Users/sovajit/Desktop/LILA/docker-compose.yml:1)
- [docker/nakama/Dockerfile](/c:/Users/sovajit/Desktop/LILA/docker/nakama/Dockerfile:1)
- [docker/nakama/local.yml](/c:/Users/sovajit/Desktop/LILA/docker/nakama/local.yml:1)

## Start

From the repo root:

```powershell
docker compose up --build -d
```

## Check logs

```powershell
docker compose logs nakama --tail=200
```

## Stop

```powershell
docker compose down
```

## Reset local database

```powershell
docker compose down -v
```

## Endpoints

- API: `http://127.0.0.1:7350`
- Console: `http://127.0.0.1:7351`
- Client server key: `lila-socket-server-key`
- Console login:
  - username: `admin`
  - password: `password`

## Current verified checks

- Nakama starts and loads the Go plugin
- Postgres migrations run successfully
- custom authentication works with the configured server key
- `get_player_profile` returns live data
- `create_room` returns a live authoritative match id

## Current caveat

- `list_rooms` will not show a freshly created room until the room has an active presence joined to it
- this is because `create_room` creates the authoritative match, but does not join the creator into that match
- frontend flow should call `create_room` and then join the returned match immediately

## What to test next

1. Confirm Nakama starts and loads the Go plugin.
2. Confirm the leaderboard is created during startup.
3. Authenticate a test user from a client.
4. Call:
   - `create_room`
   - `list_rooms`
   - `join_private_room`
   - `get_player_profile`
5. Join an authoritative match and test move flow.
