package server

import (

	"net/http"
	
)



// le ConnexionHandler gère l'affichage de la page de connexion

func ConnexionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	// Vérifier si un compte vient d'être créé SI oui affiche un message de succès et dans le cas contraire affiche la page de connexion normale

	data := LoginPageData{}
	if r.URL.Query().Get("created") == "1" {
		data.Success = "Compte créé avec succès. Vous pouvez vous connecter."
	}

	// Afficher la page de connexion avec les données appropriées

	renderLogin(w, data)
}