package server

import (
	
	"errors"
	"log"
	"net/http"
	"strings"
)


func AfficherSalleHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/salle/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}
	code := strings.TrimSpace(parts[0])

	// /salle/{code}/config
	if len(parts) >= 2 && parts[1] == "config" {
		ConfigurerSalleHandler(w, r, code)
		return
	}

	// /salle/{code}/start
	if len(parts) >= 2 && parts[1] == "start" {
		StartGameHandler(w, r, code)
		return
	}

	// /salle/{code}/leave
	if len(parts) >= 2 && parts[1] == "leave" {
		QuitterSalleHandler(w, r, code)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
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

	// Charger config spécifique
	var playlist string
	var categories []PetitBacCategory

	switch room.Type {
	case RoomTypeBlindTest:
		if p, ok, err := GetBlindtestPlaylist(r.Context(), room.ID); err != nil {
			http.Error(w, "Erreur lors du chargement de la configuration.", http.StatusInternalServerError)
			return
		} else if ok {
			playlist = p
		}
	case RoomTypePetitBac:
		cats, err := ListPetitBacCategories(r.Context(), room.ID)
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la configuration.", http.StatusInternalServerError)
			return
		}
		categories = cats
	}

	renderTemplate(w, "salle.html", SallePageData{
		Room:      room,
		Players:   players,
		GameLabel: label,
		IsAdmin:   isAdmin,

		BlindtestPlaylist:  playlist,
		PetitBacCategories: categories,
	})
}