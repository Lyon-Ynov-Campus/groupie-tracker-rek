package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type GameServer struct {
	Letter    string
	Categorie string
}

func ptitbachandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())

	categories := []string{"Artiste", "Album", "Groupe de musique", "Instrument de musique", "Featuring"}

	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomLetter := string(alphabet[rand.Intn(len(alphabet))])

	var data []GameServer

	for _, cat := range categories {
		data = append(data, GameServer{
			Letter:    randomLetter,
			Categorie: cat,
		})
	}

	t, err := template.ParseFiles("pages/ptitbac.html")
	if err != nil {
		log.Println("ERREUR CRITIQUE :", err)
		http.Error(w, "Erreur détaillée : "+err.Error(), 500)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur interne: Impossible de générer la page", 500)
		return
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", ptitbachandler)

	log.Println("Serveur démarré sur http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
