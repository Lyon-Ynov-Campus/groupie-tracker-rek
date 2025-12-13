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
	http.Handle("/salle-initialisation", server.RequireAuth(http.HandlerFunc(server.AfficherCreationSalleHandler)))
	http.Handle("/creer-salle", server.RequireAuth(http.HandlerFunc(server.CreerSalleHandler)))
	http.Handle("/rejoindre-salle", server.RequireAuth(http.HandlerFunc(server.RejoindreSalleHandler)))
	http.Handle("/salle/", server.RequireAuth(http.HandlerFunc(server.AfficherSalleHandler)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", nil)
}
