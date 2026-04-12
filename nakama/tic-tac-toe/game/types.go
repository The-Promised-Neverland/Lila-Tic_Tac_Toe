package game

import "time"

const (
	OpCodeMove  int64 = 1
	OpCodeState int64 = 2
)

const (
	DefaultTickRate        = 5
	BoardSize              = 9
	MaxPlayers             = 2
	DisconnectGraceSeconds = 15
	TimedModeTurnSeconds   = 30
	NoContestMoveThreshold = 2
	MatchLabelVersion      = 1
	DefaultRoomName        = "TicTacToe Room"
	DefaultMatchmadeRoomName = "Matchmade Room"
	MatchModuleName        = "lila_tictactoe"
	LeaderboardID          = "lila_global_rankings"
	StatsCollection        = "lila_player_stats"
	StatsKey               = "summary"
	MatchHistoryCollection = "lila_match_history"
	MatchHistoryKey        = "recent"
	RoomCodeCollection     = "lila_room_codes"
	PublicRoomLimit        = 20
)

var winningLines = [][3]int{
	{0, 1, 2},
	{3, 4, 5},
	{6, 7, 8},
	{0, 3, 6},
	{1, 4, 7},
	{2, 5, 8},
	{0, 4, 8},
	{2, 4, 6},
}

type RoomCreateRequest struct {
	Private         bool   `json:"private"`
	RoomName        string `json:"room_name"`
	Timed           bool   `json:"timed"`
	TurnTimeSeconds int    `json:"turn_time_seconds"`
}

type RoomCreateResponse struct {
	MatchID         string `json:"match_id"`
	Private         bool   `json:"private"`
	InviteCode      string `json:"invite_code,omitempty"`
	Timed           bool   `json:"timed"`
	TurnTimeSeconds int    `json:"turn_time_seconds"`
}

type JoinPrivateRoomRequest struct {
	InviteCode string `json:"invite_code"`
}

type JoinPrivateRoomResponse struct {
	MatchID    string `json:"match_id"`
	InviteCode string `json:"invite_code"`
}

type RoomListResponse struct {
	Rooms []RoomSummary `json:"rooms"`
}

type RoomSummary struct {
	MatchID         string `json:"match_id"`
	RoomName        string `json:"room_name"`
	Private         bool   `json:"private"`
	Timed           bool   `json:"timed"`
	TurnTimeSeconds int    `json:"turn_time_seconds"`
	PlayerCount     int    `json:"player_count"`
	MaxPlayers      int    `json:"max_players"`
	Open            bool   `json:"open"`
	Status          string `json:"status"`
}

type MoveRequest struct {
	Index int `json:"index"`
}

type PlayerState struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	Mark          string `json:"mark"`
	Connected     bool   `json:"connected"`
	IsCurrentTurn bool   `json:"is_current_turn"`
}

type GameStatePayload struct {
	MatchID            string          `json:"match_id"`
	RoomName           string          `json:"room_name"`
	Private            bool            `json:"private"`
	InviteCode         string          `json:"invite_code,omitempty"`
	Board              [BoardSize]string `json:"board"`
	Status             string          `json:"status"`
	WinnerUserID       string          `json:"winner_user_id,omitempty"`
	WinnerMark         string          `json:"winner_mark,omitempty"`
	CurrentTurnUserID  string          `json:"current_turn_user_id,omitempty"`
	WinningLine        []int           `json:"winning_line,omitempty"`
	Players            []PlayerState   `json:"players"`
	MoveCount          int             `json:"move_count"`
	Timed              bool            `json:"timed"`
	TurnTimeSeconds    int             `json:"turn_time_seconds"`
	TurnExpiresAt      int64           `json:"turn_expires_at,omitempty"`
	ReconnectWindowSec int             `json:"reconnect_window_seconds"`
}

type MatchLabel struct {
	Version         int    `json:"version"`
	RoomName        string `json:"room_name"`
	Private         bool   `json:"private"`
	Timed           bool   `json:"timed"`
	TurnTimeSeconds int    `json:"turn_time_seconds"`
	InviteCode      string `json:"invite_code,omitempty"`
	PlayerCount     int    `json:"player_count"`
	MaxPlayers      int    `json:"max_players"`
	Open            bool   `json:"open"`
	Status          string `json:"status"`
}

type PlayerStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
	Draws       int    `json:"draws"`
	NoContests  int    `json:"no_contests"`
	WinStreak   int    `json:"win_streak"`
	BestStreak  int    `json:"best_streak"`
	GamesPlayed int    `json:"games_played"`
}

type MatchHistoryEntry struct {
	MatchID      string    `json:"match_id"`
	PlayedAt     time.Time `json:"played_at"`
	OpponentID   string    `json:"opponent_id,omitempty"`
	OpponentName string    `json:"opponent_name,omitempty"`
	Result       string    `json:"result"`
	Forfeit      bool      `json:"forfeit"`
	NoContest    bool      `json:"no_contest"`
	Timed        bool      `json:"timed"`
	MoveCount    int       `json:"move_count"`
}

type PlayerProfileResponse struct {
	Stats   PlayerStats         `json:"stats"`
	History []MatchHistoryEntry `json:"history"`
}

type MatchPlayer struct {
	UserID       string
	Username     string
	Mark         string
	Connected    bool
	DisconnectAt int64
}

type MatchState struct {
	MatchID           string
	RoomName          string
	Private           bool
	InviteCode        string
	Timed             bool
	TurnTimeSeconds   int
	Board             [BoardSize]string
	Players           map[string]*MatchPlayer
	PlayerOrder       []string
	CurrentTurnUserID string
	Status            string
	WinnerUserID      string
	WinnerMark        string
	WinningLine       []int
	MoveCount         int
	EmptyTicks        int
	LastTick          int64
	TurnDeadlineTick  int64
	ResultRecorded    bool
	ResultPersisted   bool
}
