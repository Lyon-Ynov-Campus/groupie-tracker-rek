package server

import (
	"html/template"
	"log"
	"net/http"
)


//Nous avons mis en place une séparation claire entre la logique métier en Go et l’affichage HTML. Les erreurs sont gérées côté Go puis transmises aux templates afin d’améliorer la lisibilité du code et l’expérience utilisateur.

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