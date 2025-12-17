package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)


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

