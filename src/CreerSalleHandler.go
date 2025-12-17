package server

import (
	
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// le CreerSalleHandler gère la création de salle selon le jeu choisi


func CreerSalleHandler(w http.ResponseWriter, r *http.Request) {
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

	roomType := RoomType(strings.TrimSpace(r.FormValue("type_jeu")))
	if roomType == "" {
		roomType = RoomTypeBlindTest
	}

	maxPlayers, err := strconv.Atoi(strings.TrimSpace(r.FormValue("max_players")))
	if err != nil {
		http.Error(w, "Nombre de participants invalide.", http.StatusBadRequest)
		return
	}
	timePerRound, err := strconv.Atoi(strings.TrimSpace(r.FormValue("temps")))
	if err != nil {
		http.Error(w, "Temps par manche invalide.", http.StatusBadRequest)
		return
	}
	rounds, err := strconv.Atoi(strings.TrimSpace(r.FormValue("manches")))
	if err != nil {
		http.Error(w, "Nombre de manches invalide.", http.StatusBadRequest)
		return
	}

	room, err := CreateRoom(r.Context(), CreateRoomOptions{
		Type:         roomType,
		CreatorID:    userID,
		MaxPlayers:   maxPlayers,
		TimePerRound: timePerRound,
		Rounds:       rounds,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "Erreur lors de la création de la salle."
		switch {
		case errors.Is(err, ErrInvalidRoomParameters), errors.Is(err, ErrInvalidRoomType):
			status = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, ErrUserNotFound):
			status = http.StatusBadRequest
			message = "Utilisateur inconnu."
		}
		log.Printf("Création salle échouée : %v", err)
		http.Error(w, message, status)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
}