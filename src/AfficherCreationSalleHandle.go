package server

import (
	
	"net/http"
	"strings"
	
)

// le AfficherCreationSalleHandler gère l'affichage de la page de création de salle

func AfficherCreationSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	typeJeu := strings.TrimSpace(r.URL.Query().Get("type_jeu"))
	switch RoomType(typeJeu) {
	case RoomTypeBlindTest, RoomTypePetitBac:
		// ok
	default:
		typeJeu = string(RoomTypeBlindTest)
	}

	renderTemplate(w, "init_room.html", struct {
		TypeJeu string
	}{
		TypeJeu: typeJeu,
	})
}