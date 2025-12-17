package server

import (
	
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// le RejoindreSalleHandler gère la logique pour qu'un utilisateur rejoigne une salle existante


func RejoindreSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide.", http.StatusBadRequest)
		return
	}

	roomCode := strings.ToUpper(strings.TrimSpace(r.FormValue("room_code")))
	if roomCode == "" {
		http.Error(w, "Code de salle requis.", http.StatusBadRequest)
		return
	}

	room, err := GetRoomByCode(r.Context(), roomCode)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.Error(w, "Salle introuvable.", http.StatusNotFound)
			return
		}
		log.Printf("Recherche salle %s : %v", roomCode, err)
		http.Error(w, "Erreur lors de la récupération de la salle.", http.StatusInternalServerError)
		return
	}

	if _, err := AddRoomPlayer(r.Context(), room.ID, userID, false); err != nil {
		switch {
		case errors.Is(err, ErrRoomCapacityReached):
			http.Error(w, "La salle est complète.", http.StatusForbidden)
		case errors.Is(err, ErrPlayerAlreadyInRoom):
			http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, "Utilisateur inconnu.", http.StatusBadRequest)
		default:
			log.Printf("Rejoindre salle %s (user %d) : %v", room.Code, userID, err)
			http.Error(w, "Impossible de rejoindre la salle.", http.StatusInternalServerError)
		}
		return
	}

	BroadcastRoomUpdated(room.ID)
	http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
}