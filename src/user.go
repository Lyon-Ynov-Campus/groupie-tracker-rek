package server

import (
	"database/sql"
	"log"
	"unicode"
	"golang.org/x/crypto/bcrypt"
)


// Vérifie si un pseudo existe déjà dans la base
func IsPseudoTaken(pseudo string) bool {
	var id int

	row := Rekdb.QueryRow("SELECT id FROM users WHERE pseudo = ?", pseudo)

	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Println("Erreur SQL vérification pseudo :", err)
		return false // on évite de bloquer l'inscription
	}

	return true
}

// Vérifie si un email existe déjà dans la base
func IsEmailTaken(email string) bool {
	var id int

	row := Rekdb.QueryRow("SELECT id FROM users WHERE email = ?", email)

	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		log.Println("Erreur SQL vérification email :", err)
		return false
	}

	return true
}

// Vérifie si le mot de passe respecte les recommandations CNIL
func IsPasswordValid(password string) bool {
	if len(password) >= 12 {
		return true
	}
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erreur hash mot de passe : %v", err)
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		log.Printf("Erreur comparaison hash mot de passe : %v", err)
	}
	return err == nil
}

// Crée un nouvel utilisateur dans la base de données
func CreateUser(pseudo, email, password string) error {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return err
	}
	
	err = InsertValuesUser(pseudo, email, passwordHash)
	if err != nil {
		return err
	}
	return nil
}

