package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
)

var sessions = make(map[string]string) // sessionID -> userID


// le CreateSession crée une nouvelle session pour un utilisateur donné et retourne l'ID de session qui peut être stocké dans un cookie

func CreateSession(userID int) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	sessionID := hex.EncodeToString(token)
	sessions[sessionID] = strconv.Itoa(userID)
	return sessionID, nil
}

// le RequireAuth est un focntion qui agit comme un middleware pour protéger les routes qui nécessitent une authentification avant d'y accéder

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || sessions[cookie.Value] == "" {
			http.Redirect(w, r, "/connexion", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
