# Nakama Runtime Flow

This document explains the backend in the order it actually runs.

It separates:

- functions Nakama calls
- helpers we call ourselves

The goal is to make the runtime easier to understand before adding more code.

## Big Picture

For this project, Nakama is the backend server runtime.

That means:
- Nakama loads our Go plugin
- Nakama calls our registered callbacks
- our code uses Nakama APIs for matches, storage, and leaderboards

So there are two kinds of functions in this codebase:

- Nakama entrypoints
  Nakama invokes these automatically because we register them.
- Internal helpers
  These are normal functions we call from our own code.

## 1. Startup Flow

This happens when Nakama loads the Go runtime plugin.

### `InitModule`

File: [nakama/main.go](/c:/Users/sovajit/Desktop/LILA/nakama/main.go:1)

Who calls it:
- Nakama

Requirement:
- Nakama expects an exported `InitModule` function in the plugin

What it does:
- hands control to our Tic-Tac-Toe package

### `RegisterModule`

File: [module.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/module.go:1)

Who calls it:
- us, from `InitModule`

What it does:
- creates the leaderboard if needed
- registers the match handler
- registers RPC callbacks
- registers the matchmaker callback

### `ensureLeaderboard`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

Who calls it:
- us, from `RegisterModule`

What it does:
- makes sure the global leaderboard exists

## 2. Room And RPC Flow

These functions are called when the client asks Nakama to do something through RPC.

### `rpcCreateRoom`

File: [rpc.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/rpc.go:1)

Who calls it:
- Nakama, after the client invokes the `create_room` RPC

Requirement:
- must be registered with `initializer.RegisterRpc`

What it does:
- reads room creation settings
- creates an authoritative Nakama match
- generates an invite code for private rooms
- returns match info to the client

Helpers it uses:
- `getContextString`
- `generateInviteCode`
- `mustJSON`

### `rpcListRooms`

File: [rpc.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/rpc.go:1)

Who calls it:
- Nakama, after the client invokes the `list_rooms` RPC

Requirement:
- must be registered with `initializer.RegisterRpc`

What it does:
- asks Nakama for authoritative public matches
- converts match labels into room summaries
- returns a room list to the client

Helpers it uses:
- `mustJSON`

### `rpcJoinPrivateRoom`

File: [rpc.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/rpc.go:1)

Who calls it:
- Nakama, after the client invokes the `join_private_room` RPC

Requirement:
- must be registered with `initializer.RegisterRpc`

What it does:
- takes an invite code from the client
- normalizes it
- looks up the mapped match id
- returns the match id to join

Helpers it uses:
- `normalizeInviteCode`
- `lookupMatchIDByInviteCode`
- `mustJSON`

### `rpcGetPlayerProfile`

File: [rpc.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/rpc.go:1)

Who calls it:
- Nakama, after the client invokes the `get_player_profile` RPC

Requirement:
- must be registered with `initializer.RegisterRpc`

What it does:
- loads player stats
- loads recent match history
- returns them in one response

Helpers it uses:
- `getContextString`
- `readPlayerStats`
- `readMatchHistory`
- `mustJSON`

## 3. Matchmaker Flow

This happens when Nakama’s matchmaker has grouped players together.

### `matchmakerMatched`

File: [rpc.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/rpc.go:1)

Who calls it:
- Nakama

Requirement:
- must be registered with `initializer.RegisterMatchmakerMatched`

What it does:
- creates an authoritative Tic-Tac-Toe match for the matched players
- forwards selected mode settings like timed mode

Helpers it uses:
- `asInt`

## 4. Match Lifecycle Flow

These callbacks belong to Nakama’s authoritative match lifecycle.

Nakama calls them automatically after a match is created and players start interacting with it.

### `MatchInit`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

Requirement:
- `MatchHandler` must implement Nakama’s `runtime.Match` interface

What it does:
- creates initial server-side match state
- applies room settings like private room and timed mode
- returns state, tick rate, and label

Helpers it uses:
- `asInt`
- `label`

### `MatchJoinAttempt`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

What it does:
- decides whether a player may join
- allows reconnecting players back in
- rejects extra players when room is full
- enforces invite-code access for private rooms
- captures the match id from runtime context

Helpers it uses:
- `getContextString`
- `writeInviteCode`
- `normalizeInviteCode`

### `MatchJoin`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

What it does:
- adds accepted players to match state
- assigns marks
- starts the game when two players are present
- broadcasts current game state

Helpers it uses:
- `nextMark`
- `setTurnDeadline`
- `syncLabel`
- `broadcastState`

### `MatchLeave`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

What it does:
- marks a player as disconnected
- starts the reconnect grace window
- broadcasts updated state

Helpers it uses:
- `syncLabel`
- `broadcastState`

### `MatchLoop`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama, every tick for the active match

What it does:
- processes move messages
- validates and applies moves
- checks disconnect expiration
- checks timed-turn expiration
- can shut down an idle waiting room

Helpers it uses:
- `handleMove`
- `handleDisconnects`
- `handleTurnTimeout`
- `noConnectedPlayers`
- `deleteInviteCode`

### `MatchTerminate`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

What it does:
- performs final cleanup
- removes invite-code mapping for private rooms

Helpers it uses:
- `deleteInviteCode`

### `MatchSignal`

File: [match.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/callbacks/match.go:1)

Who calls it:
- Nakama

What it does:
- accepts external signal data for the match
- currently behaves as a pass-through placeholder

## 5. Internal Match Helpers

These functions are not called by Nakama directly.

They are our internal game logic helpers used by the lifecycle callbacks.

### `handleMove`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- validates the incoming move
- enforces turn order
- updates the board
- detects win or draw
- advances the turn if the game continues

### `handleDisconnects`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- checks if a disconnected player missed the grace window
- resolves to no-contest if the game barely started
- otherwise resolves to forfeit

### `handleTurnTimeout`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- checks timed mode deadline expiry
- converts timeout into forfeit

### `finishNoContest`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- ends the game as no-contest
- updates stats and history

### `finishForfeit`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- ends the game as forfeit
- records the winner
- persists final result data

### `persistMatchResult`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- writes player stats
- appends match history
- updates leaderboard records

### `broadcastState`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- pushes validated game state to connected clients

### `snapshot`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- converts internal match state into a client-safe payload

### `label`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- builds the match metadata used for discovery/listing

### `syncLabel`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- updates Nakama’s authoritative match label

### `setTurnDeadline`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- calculates timed-mode move expiry

### `otherPlayerID`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- returns the opposing player in a 2-player match

### `noConnectedPlayers`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- checks whether the room is currently empty

### `nextMark`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- assigns `X` or `O`

### `detectWinner`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- checks all winning line combinations

### `asInt`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- converts generic runtime values into `int`

### `secondsUntilDeadline`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- converts tick difference into seconds

## 6. Internal Storage Helpers

These are our helper functions for storage and invite-code management.

### `generateInviteCode`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- creates a short private-room code

### `normalizeInviteCode`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- standardizes invite-code format for comparison and lookup

### `writeInviteCode`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- stores invite-code to match-id mapping

### `deleteInviteCode`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- removes a private room’s invite-code mapping

### `lookupMatchIDByInviteCode`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- resolves invite code to authoritative match id

### `readPlayerStats`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- fetches saved stats for a player

### `writePlayerStats`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- writes saved stats for a player

### `readMatchHistory`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- loads recent match history

### `appendMatchHistory`

File: [storage.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/storage.go:1)

What it does:
- prepends a new history item
- keeps only the latest 50

## 7. Small Utility Helpers

### `getContextString`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- safely reads string values from Nakama runtime context

### `mustJSON`

File: [helpers.go](/c:/Users/sovajit/Desktop/LILA/nakama/tic-tac-toe/game/helpers.go:1)

What it does:
- marshals a Go value into JSON

## Should Nakama-Only Functions Be In A Separate Folder?

Short answer:
- yes at the top level
- not necessarily inside the game package

What is helpful:
- keeping all Nakama runtime code under `nakama/`
- keeping this game under `nakama/tic-tac-toe/`

What is not especially helpful yet:
- making one folder for Nakama callbacks
- another folder for helpers

Why:
- the callback flow and helper flow are tightly connected
- `MatchLoop` and `handleMove` are easier to follow when they live close together
- splitting by “who calls it” often makes navigation worse

Recommended structure for this project right now:

- `nakama/main.go`
- `nakama/tic-tac-toe/callbacks/module.go`
- `nakama/tic-tac-toe/callbacks/rpc.go`
- `nakama/tic-tac-toe/callbacks/match.go`
- `nakama/tic-tac-toe/game/helpers.go`
- `nakama/tic-tac-toe/game/storage.go`
- `nakama/tic-tac-toe/game/types.go`

If the package grows a lot later, then we can split by concern more deeply.
