package server

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type WSRoomSnapshot struct {
	Room    *Room        `json:"room"`
	Players []RoomPlayer `json:"players"`
}
