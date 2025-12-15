package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

func APISalleHandler(w http.ResponseWriter, r *http.Request) {
	// /api/salle/{code}/blindtest/state
	path := strings.TrimPrefix(r.URL.Path, "/api/salle/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	code := parts[0]

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erreur room.", http.StatusInternalServerError)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Error(w, "Non authentifié.", http.StatusUnauthorized)
		return
	}
	if ok, _ := IsUserInRoom(r.Context(), room.ID, userID); !ok {
		http.Error(w, "Accès refusé.", http.StatusForbidden)
		return
	}

	if len(parts) == 2 && parts[1] == "players" {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
			return
		}
		players, err := ListRoomPlayers(r.Context(), room.ID)
		if err != nil {
			http.Error(w, "Erreur room.", http.StatusInternalServerError)
			return
		}
		sort.Slice(players, func(i, j int) bool {
			if players[i].Score == players[j].Score {
				return players[i].Pseudo < players[j].Pseudo
			}
			return players[i].Score > players[j].Score
		})
		writeJSON(w, players)
		return
	}

	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	if parts[1] == "blindtest" {
		switch parts[2] {
		case "state":
			if r.Method != http.MethodGet {
				http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
				return
			}
			game, ok := GetBlindtestGame(room.ID)
			if !ok {
				writeJSON(w, map[string]any{"phase": "idle"})
				return
			}
			writeJSON(w, game.StateForUser(userID))
			return

		case "guess":
			if r.Method != http.MethodPost {
				http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
				return
			}
			game, ok := GetBlindtestGame(room.ID)
			if !ok {
				writeJSON(w, map[string]any{"locked": true})
				return
			}

			var body struct {
				Guess string `json:"guess"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)

			res, _ := game.SubmitGuess(r.Context(), room.ID, userID, body.Guess)
			writeJSON(w, res)
			return

		default:
			http.NotFound(w, r)
			return
		}
	}

	if parts[1] == "petitbac" {
		switch parts[2] {
		case "state":
			if r.Method != http.MethodGet {
				http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
				return
			}
			game, ok := petitBacGames[room.ID]
			if !ok {
				http.Error(w, "Aucune partie en cours.", http.StatusNotFound)
				return
			}
			writeJSON(w, game.StateForUser(userID))
			return

		case "answers":
			if r.Method != http.MethodPost {
				http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
				return
			}
			var req map[int]string
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Requête invalide.", http.StatusBadRequest)
				return
			}
			game, ok := petitBacGames[room.ID]
			if !ok {
				http.Error(w, "Aucune partie en cours.", http.StatusNotFound)
				return
			}
			if err := game.SubmitAnswers(userID, req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			BroadcastRoomUpdated(room.ID)
			writeJSON(w, map[string]string{"status": "ok"})
			return

		case "votes":
			if r.Method != http.MethodPost {
				http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
				return
			}
			var req map[int]map[int]bool
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Requête invalide.", http.StatusBadRequest)
				return
			}
			game, ok := petitBacGames[room.ID]
			if !ok {
				http.Error(w, "Aucune partie en cours.", http.StatusNotFound)
				return
			}
			if err := game.SubmitVotes(userID, req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			BroadcastRoomUpdated(room.ID)
			writeJSON(w, map[string]string{"status": "ok"})
			return
		}
	}

	http.NotFound(w, r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
