package server


//  Définir la commande SQL pour créer la table 'users'
var createTableQuery = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pseudo TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`

	
// Fonction pour inserer les données d'un nouvel utilisateur dans la base de données 

func InsertValuesUser(pseudo, email, passwordHash string) error {
	_, err := Rekdb.Exec("INSERT INTO users (pseudo, email, password_hash) VALUES (?, ?, ?)", pseudo, email, passwordHash)
	return err
}	



