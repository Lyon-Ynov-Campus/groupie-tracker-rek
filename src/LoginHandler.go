package server

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
)



// le LoginHandler gère la logique de connexion des utilisateurs une fois qu'ils ont un compte

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// On s'assure que l'utilisateur utilise la méthode POST pour envoyer le formulaire de connexion sinon on le redirige vers la page de connexion
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	// On parse le formulaire pour récupérer les données envoyées

	if err := r.ParseForm(); err != nil {
		renderLogin(w, LoginPageData{Error: "Formulaire invalide."})
		return
	}

	// Récupérer les valeurs du formulaire en les nettoyant des espaces inutileset stocker dans une structure de données

	user := strings.TrimSpace(r.FormValue("user"))
	password := r.FormValue("password")

	data := LoginPageData{User: user}

	// S'assurer que tous les champs sont remplis

	if user == "" || password == "" {
		data.Error = "Merci de remplir tous les champs."
		renderLogin(w, data)
		return
	}

	// Récupérer l'utilisateur dans la base de données (par pseudo ou email) et vérifier le mot de passe
	// Si ici notre requete sur la base de données rek.db renvoie une erreur cela signifie que l'utilisateur n'existe pas ou qu'il y a un problème avec la base de données

	var (
		storedHash string
		userID     int
	)

	row := Rekdb.QueryRow("SELECT id, password_hash FROM users WHERE pseudo = ? OR email = ?", user, user)
	if err := row.Scan(&userID, &storedHash); err != nil {
		if err == sql.ErrNoRows {
			data.Error = "Utilisateur non trouvé."
			renderLogin(w, data)
			return
		}
		log.Printf("Erreur récupération utilisateur : %v", err)
		data.Error = "Erreur lors de la récupération de l'utilisateur."
		renderLogin(w, data)
		return
	}

	// ici nous verifions que le mot de passe fourni correspond bien au hash stocké dans la base de données

	if !CheckPasswordHash(password, storedHash) {
		data.Error = "Mot de passe incorrect."
		renderLogin(w, data)
		return
	}

	// Création de la session utilisateur  avec le userID et redirection vers le tableau de bord
	sessionID, err := CreateSession(userID)
	if err != nil {
		log.Printf("Erreur création session : %v", err)
		data.Error = "Erreur interne. Merci de réessayer."
		renderLogin(w, data)
		return
	}

	// Définir le cookie de session pour l'utilisateur

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Affiche le langipage apres une connexion réussie

