package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsWriteWait      = 10 * time.Second
	wsPongWait       = 60 * time.Second
	wsPingPeriod     = (wsPongWait * 9) / 10
	wsMaxMessageSize = 1024
)

type WSClient struct {
	conn   *websocket.Conn
	send   chan []byte
	roomID int
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// même origin attendu; si besoin tu peux renforcer
		return true
	},
}

func WSRoomHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /ws/salle/{code}
	code := strings.TrimPrefix(r.URL.Path, "/ws/salle/")
	code = strings.Trim(code, "/")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Error(w, "Non authentifié.", http.StatusUnauthorized)
		return
	}

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erreur room.", http.StatusInternalServerError)
		return
	}

	// (option sécurité) vérifier que l'utilisateur est dans la salle
	if ok, err := IsUserInRoom(r.Context(), room.ID, userID); err != nil || !ok {
		http.Error(w, "Accès refusé.", http.StatusForbidden)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &WSClient{
		conn:   conn,
		send:   make(chan []byte, 32),
		roomID: room.ID,
	}

	hub := getRoomHub(room.ID)
	hub.register <- client

	players, _ := ListRoomPlayers(r.Context(), room.ID)
	client.send <- mustJSON(WSMessage{
		Type: "room_snapshot",
		Payload: WSRoomSnapshot{
			Room:    room,
			Players: players,
		},
	})

	go func() { client.writePump() }()
	client.readPump(hub)
}

func (c *WSClient) readPump(hub *RoomHub) {
	defer func() {
		hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(wsMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// On ignore les messages clients (WS = diffusion serveur)
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
