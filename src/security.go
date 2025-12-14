package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var (
	sessions   = make(map[string]string) // sessionID -> userID
	sessionsMu sync.RWMutex
)

// le CreateSession crée une nouvelle session pour un utilisateur donné et retourne l'ID de session qui peut être stocké dans un cookie

func CreateSession(userID int) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	sessionID := hex.EncodeToString(token)

	sessionsMu.Lock()
	sessions[sessionID] = strconv.Itoa(userID)
	sessionsMu.Unlock()

	return sessionID, nil
}

// le RequireAuth est un focntion qui agit comme un middleware pour protéger les routes qui nécessitent une authentification avant d'y accéder

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/connexion", http.StatusSeeOther)
			return
		}

		sessionsMu.RLock()
		userID := sessions[cookie.Value]
		sessionsMu.RUnlock()

		if userID == "" {
			http.Redirect(w, r, "/connexion", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Cette fonction permet d’identifier l’utilisateur connecté à partir du cookie de session et de sécuriser l’accès aux fonctionnalités réservées aux utilisateurs authentifiés.
//ETANT DONNé que les sessions sont stockées en mémoire pour simplifier la gestion et éviter la complexité d’un système externe.

func GetSessionUserID(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, fmt.Errorf("session manquante : %w", err)
	}

	sessionsMu.RLock()
	userIDStr, ok := sessions[cookie.Value]
	sessionsMu.RUnlock()

	if !ok || userIDStr == "" {
		return 0, fmt.Errorf("session invalide ou expirée")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("identifiant de session invalide : %w", err)
	}
	return userID, nil
}
