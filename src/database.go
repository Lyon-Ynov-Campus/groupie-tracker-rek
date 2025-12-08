package server

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Rekdb est la variable globale qui contiendra la connexion active de notre base de données.
var Rekdb *sql.DB

type User struct {
	ID           int
	Pseudo       string
	Email        string
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

	// Tester la connexion (Ping)
	if err = database.Ping(); err != nil {
		log.Printf("Erreur lors du test de connexion (Ping): %v\n", err)
		database.Close()
		return nil, err
	}
	log.Println("Connexion à SQLite établie et vérifiée.")

	// Activer les clés étrangères
	if _, err = database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Printf("Activation des clés étrangères impossible : %v\n", err)
		database.Close()
		return nil, err
	}

	tableStatements := []struct {
		name  string
		query string
	}{
		{"users", `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pseudo TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		);`},
		{"rooms", `CREATE TABLE IF NOT EXISTS rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('blindtest','petit_bac')),
			creator_id INTEGER NOT NULL,
			max_players INTEGER NOT NULL,
			time_per_round INTEGER NOT NULL,
			rounds INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'lobby',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
		);`},
		{"room_players", `CREATE TABLE IF NOT EXISTS room_players (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			pseudo TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0,
			is_ready INTEGER NOT NULL DEFAULT 0,
			score INTEGER NOT NULL DEFAULT 0,
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(room_id, user_id),
			FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`},
		{"room_blindtest", `CREATE TABLE IF NOT EXISTS room_blindtest (
			room_id INTEGER PRIMARY KEY,
			playlist_id TEXT NOT NULL,
			playlist_name TEXT,
			FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
		);`},
		{"room_petitbac_categories", `CREATE TABLE IF NOT EXISTS room_petitbac_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id INTEGER NOT NULL,
			label TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(room_id, label),
			FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
		);`},
	}

	for _, stmt := range tableStatements {
		log.Printf("Vérification/Création de la table '%s'...", stmt.name)
		if _, err = database.Exec(stmt.query); err != nil {
			log.Printf("Erreur lors de la création de la table '%s' : %v\n", stmt.name, err)
			return nil, err
		}
	}

	indexStatements := []struct {
		name  string
		query string
	}{
		{"idx_rooms_creator", "CREATE INDEX IF NOT EXISTS idx_rooms_creator ON rooms(creator_id);"},
		{"idx_room_players_room", "CREATE INDEX IF NOT EXISTS idx_room_players_room ON room_players(room_id);"},
		{"idx_room_petitbac_room", "CREATE INDEX IF NOT EXISTS idx_room_petitbac_room ON room_petitbac_categories(room_id);"},
	}

	for _, idx := range indexStatements {
		log.Printf("Vérification/Création de l'index '%s'...", idx.name)
		if _, err = database.Exec(idx.query); err != nil {
			log.Printf("Erreur lors de la création de l'index '%s' : %v\n", idx.name, err)
			return nil, err
		}
	}

	Rekdb = database

	return database, nil
}
