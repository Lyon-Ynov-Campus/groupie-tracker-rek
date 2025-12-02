package main

import (
	"log"
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
}