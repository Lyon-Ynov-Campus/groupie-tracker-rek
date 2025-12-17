package server

import (
	
	"net/http"
	"time"
)

// le LogoutHandler gère la déconnexion de l'utilisateur

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie("session_id"); err == nil {
		DeleteSession(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	http.Redirect(w, r, "/connexion", http.StatusSeeOther)
}