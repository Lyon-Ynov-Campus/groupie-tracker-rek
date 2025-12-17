package server

import (
	
	"net/http"
	
)

// le Homehandler(qui est la premiere page de notre application) affiche la page d'inscription avec un bouton qui dirige vers la page de connexion
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderRegister(w, RegisterPageData{})
}