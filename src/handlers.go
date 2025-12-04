package server

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

type RegisterPageData struct {
	Error  string
	Values map[string]string
}

type LoginPageData struct {
	Error   string
	Success string
	User    string
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t, err := template.ParseFiles("./templates/" + name)
	if err != nil {
		log.Printf("Erreur chargement template %s : %v", name, err)
		http.Error(w, "Erreur serveur.", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Printf("Erreur rendu template %s : %v", name, err)
		http.Error(w, "Erreur serveur.", http.StatusInternalServerError)
	}
}

func renderRegister(w http.ResponseWriter, data RegisterPageData) {
	renderTemplate(w, "accueil.html", data)
}

func renderLogin(w http.ResponseWriter, data LoginPageData) {
	renderTemplate(w, "authentification.html", data)
}

// le Homehandler(qui est la premiere page de notre application) affiche la page d'inscription avec un bouton qui dirige vers la page de connexion
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderRegister(w, RegisterPageData{})
}

// le RegisterHandler gère la logique d'inscription des utilisateurs

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		renderRegister(w, RegisterPageData{Error: "Formulaire invalide."})
		return
	}

	pseudo := strings.TrimSpace(r.FormValue("pseudo"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	data := RegisterPageData{
		Values: map[string]string{
			"pseudo": pseudo,
			"email":  email,
		},
	}

	// s'assurer que tous les champs de l'inscription sont remplis correctement

	if pseudo == "" || email == "" || password == "" || confirm == "" {
		data.Error = "Merci de remplir tous les champs."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mail n'est pas encore présent dans notre base de données

	if IsEmailTaken(email) {
		data.Error = "Cet email est déjà utilisé."
		renderRegister(w, data)
		return
	}

	// s'assurer que le pseudo n'est pas encore présent dans notre base de données

	if IsPseudoTaken(pseudo) {
		data.Error = "Ce pseudo est déjà pris."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mot de passe respecte les règles CNIL

	if !IsPasswordValid(password) {
		data.Error = "Mot de passe non conforme aux règles CNIL."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mot de passe et sa confirmation correspondent

	if password != confirm {
		data.Error = "Les mots de passe ne correspondent pas."
		renderRegister(w, data)
		return
	}

	// maintenant on peut créer l'utilisateur en utilisant la fonction CreateUser de user.go

	if err := CreateUser(pseudo, email, password); err != nil {
		log.Printf("Erreur création utilisateur : %v", err)
		data.Error = "Erreur lors de la création du compte."
		renderRegister(w, data)
		return
	}

	// une fois l'inscription réussie, on redirige l'utilisateur vers la page de connexion avec un message de succès

	http.Redirect(w, r, "/connexion?created=1", http.StatusSeeOther)
}

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

func LandingPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "landingpage.html", nil)
}

func DeleteSession(sessionID string) {
	delete(sessions, sessionID)
}

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
