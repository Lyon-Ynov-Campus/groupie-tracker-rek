package main

import (
	"log"
	"net/http"
	"text/template"
)

func FinalScoreboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("pages/finalscoreboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", FinalScoreboard)
	log.Println("Starting server on :8080")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
