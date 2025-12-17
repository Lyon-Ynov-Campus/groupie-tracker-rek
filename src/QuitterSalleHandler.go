package server

import (	
	"errors"
	"log"
	"net/http"	
)

// le QuitterSalleHandler gère la déconnexion d'un utilisateur d'une salle de jeu

func QuitterSalleHandler(w http.ResponseWriter, r *http.Request, code string) {
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

	// Vérifier que l'utilisateur est bien membre de la salle
	ok, err := IsUserInRoom(r.Context(), room.ID, userID)
	if err != nil {
		log.Printf("Check membre salle %s : %v", room.Code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Redirect(w, r, "/salle-initialisation", http.StatusSeeOther)
		return
	}

	var pseudo string
	if err := Rekdb.QueryRowContext(r.Context(), SQLSelectUserPseudoByID, userID).Scan(&pseudo); err != nil {
		pseudo = ""
	}

	if err := RemoveRoomPlayer(r.Context(), room.ID, userID); err != nil {
		http.Error(w, "Impossible de quitter la salle.", http.StatusInternalServerError)
		return
	}

	BroadcastRoomUpdated(room.ID)
	if pseudo != "" {
		BroadcastPlayerLeft(room.ID, pseudo)
	}

	http.Redirect(w, r, "/salle-initialisation", http.StatusSeeOther)
}
