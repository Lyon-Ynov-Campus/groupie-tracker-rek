package main

import (
	"fmt"
	"net/http"
	"rek/spotify"
)

// handler pour la connexion spotify
func spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	//generer l'URL d'autorisation
	authURL := spotify.GetAuthURL()
	//rediriger vers spotify
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handler pour le callback spotify
func spotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	//recuperer le code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Pas de code", http.StatusBadRequest)
		return
	}

	//echanger le code contre un token
	err := spotify.ExchangeCode(code)
	if err != nil {
		http.Error(w, "Erreur token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//charger les musiques
	InitSpotify()

	//rediriger vers la page d'accueil
	fmt.Fprintf(w, "<h1>Connexion Spotify réussie!</h1><p>%d musiques chargées.</p><a href='/'>Retour</a>", len(tracks))
}
