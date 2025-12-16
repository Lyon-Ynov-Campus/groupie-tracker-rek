package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// le Homehandler(qui est la premiere page de notre application) affiche la page d'inscription avec un bouton qui dirige vers la page de connexion
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderRegister(w, RegisterPageData{})
}

// le RegisterHandler gère la logique d'inscription des utilisateurs

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		renderRegister(w, RegisterPageData{Error: "Formulaire invalide."})
		return
	}

	pseudo := strings.TrimSpace(r.FormValue("pseudo"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	data := RegisterPageData{
		Values: map[string]string{
			"pseudo": pseudo,
			"email":  email,
		},
	}

	// s'assurer que tous les champs de l'inscription sont remplis correctement

	if pseudo == "" || email == "" || password == "" || confirm == "" {
		data.Error = "Merci de remplir tous les champs."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mail n'est pas encore présent dans notre base de données

	if IsEmailTaken(email) {
		data.Error = "Cet email est déjà utilisé."
		renderRegister(w, data)
		return
	}

	// s'assurer que le pseudo n'est pas encore présent dans notre base de données

	if IsPseudoTaken(pseudo) {
		data.Error = "Ce pseudo est déjà pris."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mot de passe respecte les règles CNIL

	if !IsPasswordValid(password) {
		data.Error = "Mot de passe non conforme aux règles CNIL."
		renderRegister(w, data)
		return
	}

	// s'assurer que le mot de passe et sa confirmation correspondent

	if password != confirm {
		data.Error = "Les mots de passe ne correspondent pas."
		renderRegister(w, data)
		return
	}

	// maintenant on peut créer l'utilisateur en utilisant la fonction CreateUser de user.go

	if err := CreateUser(pseudo, email, password); err != nil {
		log.Printf("Erreur création utilisateur : %v", err)
		data.Error = "Erreur lors de la création du compte."
		renderRegister(w, data)
		return
	}

	// une fois l'inscription réussie, on redirige l'utilisateur vers la page de connexion avec un message de succès

	http.Redirect(w, r, "/connexion?created=1", http.StatusSeeOther)
}

// le ConnexionHandler gère l'affichage de la page de connexion

func ConnexionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	// Vérifier si un compte vient d'être créé SI oui affiche un message de succès et dans le cas contraire affiche la page de connexion normale

	data := LoginPageData{}
	if r.URL.Query().Get("created") == "1" {
		data.Success = "Compte créé avec succès. Vous pouvez vous connecter."
	}

	// Afficher la page de connexion avec les données appropriées

	renderLogin(w, data)
}

// le LoginHandler gère la logique de connexion des utilisateurs une fois qu'ils ont un compte

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// On s'assure que l'utilisateur utilise la méthode POST pour envoyer le formulaire de connexion sinon on le redirige vers la page de connexion
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	// On parse le formulaire pour récupérer les données envoyées

	if err := r.ParseForm(); err != nil {
		renderLogin(w, LoginPageData{Error: "Formulaire invalide."})
		return
	}

	// Récupérer les valeurs du formulaire en les nettoyant des espaces inutileset stocker dans une structure de données

	user := strings.TrimSpace(r.FormValue("user"))
	password := r.FormValue("password")

	data := LoginPageData{User: user}

	// S'assurer que tous les champs sont remplis

	if user == "" || password == "" {
		data.Error = "Merci de remplir tous les champs."
		renderLogin(w, data)
		return
	}

	// Récupérer l'utilisateur dans la base de données (par pseudo ou email) et vérifier le mot de passe
	// Si ici notre requete sur la base de données rek.db renvoie une erreur cela signifie que l'utilisateur n'existe pas ou qu'il y a un problème avec la base de données

	var (
		storedHash string
		userID     int
	)

	row := Rekdb.QueryRow("SELECT id, password_hash FROM users WHERE pseudo = ? OR email = ?", user, user)
	if err := row.Scan(&userID, &storedHash); err != nil {
		if err == sql.ErrNoRows {
			data.Error = "Utilisateur non trouvé."
			renderLogin(w, data)
			return
		}
		log.Printf("Erreur récupération utilisateur : %v", err)
		data.Error = "Erreur lors de la récupération de l'utilisateur."
		renderLogin(w, data)
		return
	}

	// ici nous verifions que le mot de passe fourni correspond bien au hash stocké dans la base de données

	if !CheckPasswordHash(password, storedHash) {
		data.Error = "Mot de passe incorrect."
		renderLogin(w, data)
		return
	}

	// Création de la session utilisateur  avec le userID et redirection vers le tableau de bord
	sessionID, err := CreateSession(userID)
	if err != nil {
		log.Printf("Erreur création session : %v", err)
		data.Error = "Erreur interne. Merci de réessayer."
		renderLogin(w, data)
		return
	}

	// Définir le cookie de session pour l'utilisateur

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Affiche le langipage apres une connexion réussie

func LandingPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "landingpage.html", nil)
}

func DeleteSession(sessionID string) {
	sessionsMu.Lock()
	delete(sessions, sessionID)
	sessionsMu.Unlock()
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie("session_id"); err == nil {
		DeleteSession(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	http.Redirect(w, r, "/connexion", http.StatusSeeOther)
}

func AfficherCreationSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	typeJeu := strings.TrimSpace(r.URL.Query().Get("type_jeu"))
	switch RoomType(typeJeu) {
	case RoomTypeBlindTest, RoomTypePetitBac:
		// ok
	default:
		typeJeu = string(RoomTypeBlindTest)
	}

	renderTemplate(w, "init_room.html", struct {
		TypeJeu string
	}{
		TypeJeu: typeJeu,
	})
}

func CreerSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide.", http.StatusBadRequest)
		return
	}

	roomType := RoomType(strings.TrimSpace(r.FormValue("type_jeu")))
	if roomType == "" {
		roomType = RoomTypeBlindTest
	}

	maxPlayers, err := strconv.Atoi(strings.TrimSpace(r.FormValue("max_players")))
	if err != nil {
		http.Error(w, "Nombre de participants invalide.", http.StatusBadRequest)
		return
	}
	timePerRound, err := strconv.Atoi(strings.TrimSpace(r.FormValue("temps")))
	if err != nil {
		http.Error(w, "Temps par manche invalide.", http.StatusBadRequest)
		return
	}
	rounds, err := strconv.Atoi(strings.TrimSpace(r.FormValue("manches")))
	if err != nil {
		http.Error(w, "Nombre de manches invalide.", http.StatusBadRequest)
		return
	}

	room, err := CreateRoom(r.Context(), CreateRoomOptions{
		Type:         roomType,
		CreatorID:    userID,
		MaxPlayers:   maxPlayers,
		TimePerRound: timePerRound,
		Rounds:       rounds,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "Erreur lors de la création de la salle."
		switch {
		case errors.Is(err, ErrInvalidRoomParameters), errors.Is(err, ErrInvalidRoomType):
			status = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, ErrUserNotFound):
			status = http.StatusBadRequest
			message = "Utilisateur inconnu."
		}
		log.Printf("Création salle échouée : %v", err)
		http.Error(w, message, status)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
}

func AfficherSalleHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/salle/")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}
	code := strings.TrimSpace(parts[0])

	// /salle/{code}/config
	if len(parts) >= 2 && parts[1] == "config" {
		ConfigurerSalleHandler(w, r, code)
		return
	}

	// /salle/{code}/start
	if len(parts) >= 2 && parts[1] == "start" {
		StartGameHandler(w, r, code)
		return
	}

	// /salle/{code}/leave
	if len(parts) >= 2 && parts[1] == "leave" {
		QuitterSalleHandler(w, r, code)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("Chargement salle échoué (%s) : %v", code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}

	players, err := ListRoomPlayers(r.Context(), room.ID)
	if err != nil {
		log.Printf("Listing joueurs salle %s : %v", room.Code, err)
		http.Error(w, "Impossible de récupérer les joueurs.", http.StatusInternalServerError)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	isAdmin, err := IsUserAdminInRoom(r.Context(), room.ID, userID)
	if err != nil {
		log.Printf("Admin check salle %s : %v", room.Code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}

	label := "Salle"
	switch room.Type {
	case RoomTypeBlindTest:
		label = "Blind Test"
	case RoomTypePetitBac:
		label = "Petit Bac"
	}

	// Charger config spécifique
	var playlist string
	var categories []PetitBacCategory

	switch room.Type {
	case RoomTypeBlindTest:
		if p, ok, err := GetBlindtestPlaylist(r.Context(), room.ID); err != nil {
			http.Error(w, "Erreur lors du chargement de la configuration.", http.StatusInternalServerError)
			return
		} else if ok {
			playlist = p
		}
	case RoomTypePetitBac:
		cats, err := ListPetitBacCategories(r.Context(), room.ID)
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la configuration.", http.StatusInternalServerError)
			return
		}
		categories = cats
	}

	renderTemplate(w, "salle.html", SallePageData{
		Room:      room,
		Players:   players,
		GameLabel: label,
		IsAdmin:   isAdmin,

		BlindtestPlaylist:  playlist,
		PetitBacCategories: categories,
	})
}

func RejoindreSalleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulaire invalide.", http.StatusBadRequest)
		return
	}

	roomCode := strings.ToUpper(strings.TrimSpace(r.FormValue("room_code")))
	if roomCode == "" {
		http.Error(w, "Code de salle requis.", http.StatusBadRequest)
		return
	}

	room, err := GetRoomByCode(r.Context(), roomCode)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.Error(w, "Salle introuvable.", http.StatusNotFound)
			return
		}
		log.Printf("Recherche salle %s : %v", roomCode, err)
		http.Error(w, "Erreur lors de la récupération de la salle.", http.StatusInternalServerError)
		return
	}

	if _, err := AddRoomPlayer(r.Context(), room.ID, userID, false); err != nil {
		switch {
		case errors.Is(err, ErrRoomCapacityReached):
			http.Error(w, "La salle est complète.", http.StatusForbidden)
		case errors.Is(err, ErrPlayerAlreadyInRoom):
			http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, "Utilisateur inconnu.", http.StatusBadRequest)
		default:
			log.Printf("Rejoindre salle %s (user %d) : %v", room.Code, userID, err)
			http.Error(w, "Impossible de rejoindre la salle.", http.StatusInternalServerError)
		}
		return
	}

	BroadcastRoomUpdated(room.ID)
	http.Redirect(w, r, fmt.Sprintf("/salle/%s", room.Code), http.StatusSeeOther)
}

type SalleConfigPageData struct {
	Room               *Room
	GameLabel          string
	Error              string
	BlindtestPlaylist  string
	PetitBacCategories []PetitBacCategory
}

func ConfigurerSalleHandler(w http.ResponseWriter, r *http.Request, code string) {
	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}
	isAdmin, err := IsUserAdminInRoom(r.Context(), room.ID, userID)
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Réservé à l'administrateur.", http.StatusForbidden)
		return
	}

	label := "Salle"
	switch room.Type {
	case RoomTypeBlindTest:
		label = "Blind Test"
	case RoomTypePetitBac:
		label = "Petit Bac"
	}

	switch room.Type {
	case RoomTypeBlindTest:
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Formulaire invalide.", http.StatusBadRequest)
				return
			}
			playlist := strings.TrimSpace(r.FormValue("playlist"))
			switch strings.ToLower(playlist) {
			case "rock":
				playlist = "Rock"
			case "rap":
				playlist = "Rap"
			case "pop":
				playlist = "Pop"
			default:
				renderTemplate(w, "config_salle.html", SalleConfigPageData{
					Room:              room,
					GameLabel:         label,
					Error:             "Playlist invalide (Rock, Rap, Pop).",
					BlindtestPlaylist: playlist,
				})
				return
			}
			if err := SetBlindtestPlaylist(r.Context(), room.ID, playlist); err != nil {
				http.Error(w, "Erreur lors de l'enregistrement.", http.StatusInternalServerError)
				return
			}
			BroadcastRoomUpdated(room.ID)
			http.Redirect(w, r, "/salle/"+room.Code, http.StatusSeeOther)
			return
		}

		playlist, _, _ := GetBlindtestPlaylist(r.Context(), room.ID)
		renderTemplate(w, "config_salle.html", SalleConfigPageData{
			Room:              room,
			GameLabel:         label,
			BlindtestPlaylist: playlist,
		})
		return

	case RoomTypePetitBac:
		if err := EnsureDefaultPetitBacCategories(r.Context(), room.ID); err != nil {
			http.Error(w, "Erreur lors de l'initialisation des catégories.", http.StatusInternalServerError)
			return
		}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Formulaire invalide.", http.StatusBadRequest)
				return
			}

			cats, err := ListPetitBacCategories(r.Context(), room.ID)
			if err != nil {
				http.Error(w, "Erreur lors du chargement.", http.StatusInternalServerError)
				return
			}

			for _, c := range cats {
				if r.FormValue("delete_"+strconv.Itoa(c.ID)) == "on" {
					_ = DeletePetitBacCategory(r.Context(), room.ID, c.ID)
					continue
				}
				newName := strings.TrimSpace(r.FormValue("name_" + strconv.Itoa(c.ID)))
				if newName != "" && newName != c.Name {
					_ = UpdatePetitBacCategoryName(r.Context(), room.ID, c.ID, newName)
				}
			}

			newCat := strings.TrimSpace(r.FormValue("new_category"))
			if newCat != "" {
				_ = AddPetitBacCategory(r.Context(), room.ID, newCat)
			}

			BroadcastRoomUpdated(room.ID)
			http.Redirect(w, r, "/salle/"+room.Code+"/config", http.StatusSeeOther)
			return
		}

		cats, err := ListPetitBacCategories(r.Context(), room.ID)
		if err != nil {
			http.Error(w, "Erreur lors du chargement.", http.StatusInternalServerError)
			return
		}
		renderTemplate(w, "config_salle.html", SalleConfigPageData{
			Room:               room,
			GameLabel:          label,
			PetitBacCategories: cats,
		})
		return
	default:
		http.Error(w, "Type de salle invalide.", http.StatusBadRequest)
		return
	}
}

func QuitterSalleHandler(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée.", http.StatusMethodNotAllowed)
		return
	}

	userID, err := GetSessionUserID(r)
	if err != nil {
		http.Redirect(w, r, "/connexion", http.StatusSeeOther)
		return
	}

	room, err := GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erreur room.", http.StatusInternalServerError)
		return
	}

	// Vérifier que l'utilisateur est bien membre de la salle
	ok, err := IsUserInRoom(r.Context(), room.ID, userID)
	if err != nil {
		log.Printf("Check membre salle %s : %v", room.Code, err)
		http.Error(w, "Erreur lors du chargement de la salle.", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Redirect(w, r, "/salle-initialisation", http.StatusSeeOther)
		return
	}

	var pseudo string
	if err := Rekdb.QueryRowContext(r.Context(), SQLSelectUserPseudoByID, userID).Scan(&pseudo); err != nil {
		pseudo = ""
	}

	if err := RemoveRoomPlayer(r.Context(), room.ID, userID); err != nil {
		http.Error(w, "Impossible de quitter la salle.", http.StatusInternalServerError)
		return
	}

	BroadcastRoomUpdated(room.ID)
	if pseudo != "" {
		BroadcastPlayerLeft(room.ID, pseudo)
	}

	http.Redirect(w, r, "/salle-initialisation", http.StatusSeeOther)
}
