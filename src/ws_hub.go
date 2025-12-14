package server

import "sync"

var (
	roomHubsMu sync.Mutex
	roomHubs   = map[int]*RoomHub{} // roomID -> hub
)

type RoomHub struct {
	roomID     int
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan []byte

	mu      sync.Mutex
	clients map[*WSClient]struct{}
}

func getRoomHub(roomID int) *RoomHub {
	roomHubsMu.Lock()
	defer roomHubsMu.Unlock()

	if h, ok := roomHubs[roomID]; ok {
		return h
	}
	h := &RoomHub{
		roomID:     roomID,
		register:   make(chan *WSClient, 16),
		unregister: make(chan *WSClient, 16),
		broadcast:  make(chan []byte, 64),
		clients:    make(map[*WSClient]struct{}),
	}
	roomHubs[roomID] = h
	go h.run()
	return h
}

func (h *RoomHub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func BroadcastRoomUpdated(roomID int) {
	h := getRoomHub(roomID)
	h.broadcast <- mustJSON(WSMessage{Type: "room_updated", Payload: map[string]any{"room_id": roomID}})
}

func BroadcastPlayerLeft(roomID int, pseudo string) {
	h := getRoomHub(roomID)
	h.broadcast <- mustJSON(WSMessage{
		Type:    "player_left",
		Payload: map[string]any{"room_id": roomID, "pseudo": pseudo},
	})
}
