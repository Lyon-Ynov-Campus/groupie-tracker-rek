package server


import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" 
)

// Rekdb est la variable globale qui contiendra la connexion active de notre base de données.
var Rekdb *sql.DB 


type User struct {
	ID            int
	Pseudo        string
	Email         string
	PasswordHash string 
}

// InitDB initialise la connexion à la base de données SQLite et crée la table 'users'.

func InitDB(filepath string) (*sql.DB, error) {

	var err error
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Printf("Erreur lors de l'ouverture de la base de données (%s) : %v\n", filepath, err)
		return nil, err 
	}
    
	// TestONS la connexion  avec  (Ping)
	
	if err = database.Ping(); err != nil {
		log.Printf("Erreur lors du test de connexion (Ping): %v\n", err)
		database.Close() 
		return nil, err
	}
	log.Println("Connexion à SQLite établie et vérifiée.")

	// ensuuite on execute  la commande SQL pour créer la table

	log.Println("Vérification/Création de la table 'users'...")
	_, err = database.Exec(createTableQuery)
	if err != nil {
		log.Printf("Erreur lors de la création de la table 'users' : %v\n", err)
		return nil, err
	} 
	
	log.Println("Table 'users' prête avec les contraintes d'unicité.")
	
    Rekdb = database 

	return database, nil 
}



 