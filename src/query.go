package server



// Fonction pour inserer les données d'un nouvel utilisateur

func InsertValuesUser(pseudo, email, passwordHash string) error {
	_, err := Rekdb.Exec("INSERT INTO users (pseudo, email, password_hash) VALUES (?, ?, ?)", pseudo, email, passwordHash)
	return err
}	

