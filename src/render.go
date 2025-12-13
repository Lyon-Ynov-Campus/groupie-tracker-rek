package server

import (
	"html/template"
	"log"
	"net/http"
)


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
