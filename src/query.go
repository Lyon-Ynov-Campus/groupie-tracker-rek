package server


// Fonction pour inserer les données d'un nouvel utilisateur dans la base de données 

func InsertValuesUser(pseudo, email, passwordHash string) error {
	_, err := Rekdb.Exec("INSERT INTO users (pseudo, email, password_hash) VALUES (?, ?, ?)", pseudo, email, passwordHash)
	return err
}	

