package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type RegisterPageData struct {
	Error  string
	Values map[string]string
}

type LoginPageData struct {
	Error   string
	Success string
	User    string
}

func AfficherCreationSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "init_room.html", nil)
}

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

func AfficherSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/salle/")
	code = strings.SplitN(code, "/", 2)[0]
	code = strings.TrimSpace(code)
	if code == "" {
		http.NotFound(w, r)
		return
	}

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("Chargement salle échoué (%s) : %v", code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}

	players, err := ListRoomPlayers(r.Context(), room.ID)
	if err != nil {
		log.Printf("Listing joueurs salle %s : %v", room.Code, err)
		http.Error(w, "Impossible de récupérer les joueurs.", http.StatusInternalServerError)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	isAdmin, err := IsUserAdminInRoom(r.Context(), room.ID, userID)
	if err != nil {
		log.Printf("Admin check salle %s : %v", room.Code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}

	label := "Salle"
	switch room.Type {
	case RoomTypeBlindTest:
		label = "Blind Test"
	case RoomTypePetitBac:
		label = "Petit Bac"
	}

	renderTemplate(w, "salle.html", SallePageData{
		Room:      room,
		Players:   players,
		GameLabel: label,
		IsAdmin:   isAdmin,
	})
}

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

	http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
}
