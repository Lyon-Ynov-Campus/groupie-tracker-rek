package main

import (
	"log"
	"net/http"
	server "rek/src"
)

func main() {
	// Initialiser la base de données
	db, err := server.InitDB("./rek.db")
	if err != nil {
		log.Fatalf("Échec de l'initialisation de la base de données : %v", err)
	}
	defer db.Close()
	log.Println("Base de données initialisée avec succès.")

	http.HandleFunc("/", server.HomeHandler)
	http.HandleFunc("/register", server.RegisterHandler)
	http.HandleFunc("/connexion", server.ConnexionHandler)
	http.HandleFunc("/login", server.LoginHandler)
	http.Handle("/dashboard", server.RequireAuth(http.HandlerFunc(server.LandingPageHandler)))
	http.Handle("/logout", server.RequireAuth(http.HandlerFunc(server.LogoutHandler)))
	http.Handle("/blindtest-initialisation", server.RequireAuth(http.HandlerFunc(server.AfficherCreationBlindTestHandler)))
	http.Handle("/petitbac-initialisation", server.RequireAuth(http.HandlerFunc(server.AfficherCreationPetitBacHandler)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", nil)
}
