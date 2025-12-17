package server

import (
	"strings"
	"log"
	"net/http"
)

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
