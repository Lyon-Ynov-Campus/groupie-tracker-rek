package server

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Rekdb *sql.DB

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Printf("Erreur lors de l'ouverture de la base de données (%s) : %v\n", filepath, err)
		return nil, err
	}

	// TestONS la connexion  avec  (Ping)

	if err = db.Ping(); err != nil {
		log.Printf("Erreur lors du test de connexion (Ping): %v\n", err)
		db.Close()
		log.Printf("Erreur ouverture DB (%s) : %v\n", filepath, err)

		return nil, err
	}

	// ensuuite on execute  la commande SQL pour créer la table

	log.Println("Connexion SQLite établie et vérifiée.")

	// Activer les clés étrangères pour pouvoir utiliser les contraintes de clés étrangères
	if _, err = db.Exec(SQLPragmaForeignKeysOn); err != nil {
		log.Printf("Erreur activation clés étrangères: %v\n", err)
		db.Close()
		return nil, err
	}

	// Créer les tables
	for name, query := range TablesSQL {
		log.Printf("Création table '%s'...", name)
		if _, err = db.Exec(query); err != nil {
			log.Printf("Erreur création table '%s' : %v\n", name, err)
			db.Close()
			return nil, err
		}
	}

	// Créer les indexes

	for name, query := range IndexesSQL {
		log.Printf("Création index '%s'...", name)
		if _, err = db.Exec(query); err != nil {
			log.Printf("Erreur création index '%s' : %v\n", name, err)
			db.Close()
			return nil, err
		}
	}

	Rekdb = db
	log.Println("Base de données initialisée avec succès.")
	return db, nil
}
