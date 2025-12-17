package server

import (
	
	"net/http"
	
)

// le LandingPageHandler gère l'affichage de la page d'atterrissage après la connexion	

func LandingPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "landingpage.html", nil)
}