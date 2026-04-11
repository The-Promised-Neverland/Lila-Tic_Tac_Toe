package game

import "strings"

func NewMatchState(params map[string]interface{}) *MatchState {
	roomName, _ := params["room_name"].(string)
	if strings.TrimSpace(roomName) == "" {
		roomName = DefaultRoomName
	}

	privateRoom, _ := params["private"].(bool)
	timed, _ := params["timed"].(bool)
	inviteCode, _ := params["invite_code"].(string)

	turnTimeSeconds := TimedModeTurnSeconds
	if seconds, ok := AsInt(params["turn_time_seconds"]); ok && seconds > 0 {
		turnTimeSeconds = seconds
	}

	return &MatchState{
		RoomName:        roomName,
		Private:         privateRoom,
		InviteCode:      NormalizeInviteCode(inviteCode),
		Timed:           timed,
		TurnTimeSeconds: turnTimeSeconds,
		Players:         make(map[string]*MatchPlayer, MaxPlayers),
		PlayerOrder:     make([]string, 0, MaxPlayers),
		Status:          "waiting",
	}
}
