package server

import (
	"errors"
	"net/http"
	"strings"
)

func StartBlindtestHandler(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
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
	if room.Type != RoomTypeBlindTest {
		http.Error(w, "Réservé au Blindtest.", http.StatusBadRequest)
		return
	}

	isAdmin, _ := IsUserAdminInRoom(r.Context(), room.ID, userID)
	if !isAdmin {
		http.Error(w, "Réservé à l'administrateur.", http.StatusForbidden)
		return
	}

	playlist, ok, err := GetBlindtestPlaylist(r.Context(), room.ID)
	if err != nil || !ok || strings.TrimSpace(playlist) == "" {
		http.Error(w, "Playlist non configurée.", http.StatusBadRequest)
		return
	}

	if _, err := StartOrResetBlindtest(r.Context(), room, playlist); err != nil {
		http.Error(w, "Erreur Deezer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/game/"+room.Code, http.StatusSeeOther)
}

func GameHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/game/")
	code = strings.Trim(code, "/")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if ok, _ := IsUserInRoom(r.Context(), room.ID, userID); !ok {
		http.Error(w, "Accès refusé.", http.StatusForbidden)
		return
	}

	renderTemplate(w, "game.html", struct {
		Code string
	}{
		Code: room.Code,
	})
}
