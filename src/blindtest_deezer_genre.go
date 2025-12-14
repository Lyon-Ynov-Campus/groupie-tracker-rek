package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type BlindtestTrack struct {
	TrackID    int64
	PreviewURL string
	Title      string
	Artist     string
}

type deezerGenre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type deezerGenresResp struct {
	Data []deezerGenre `json:"data"`
}

type deezerRadio struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
type deezerRadiosResp struct {
	Data []deezerRadio `json:"data"`
}

type deezerRadioTracksResp struct {
	Data []struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Preview string `json:"preview"`
		Artist  struct {
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"data"`
}

func deezerDoJSON(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("deezer status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func deezerGenreNameForType(playlistType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(playlistType)) {
	case "rock":
		return "Rock", nil
	case "rap":
		// Deezer utilise souvent ce label
		return "Rap/Hip Hop", nil
	case "pop":
		return "Pop", nil
	default:
		return "", errors.New("playlist invalide (Rock/Rap/Pop)")
	}
}

func resolveDeezerGenreID(ctx context.Context, wantedName string) (int64, error) {
	var resp deezerGenresResp
	if err := deezerDoJSON(ctx, "https://api.deezer.com/genre", &resp); err != nil {
		return 0, err
	}
	for _, g := range resp.Data {
		if strings.EqualFold(strings.TrimSpace(g.Name), strings.TrimSpace(wantedName)) {
			return g.ID, nil
		}
	}
	return 0, fmt.Errorf("genre Deezer introuvable: %s", wantedName)
}

func isFrenchRadio(title string) bool {
	lower := strings.ToLower(title)
	keywords := []string{"france", "français", "francais", "french", "chanson"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func selectRadios(radios []deezerRadio) []deezerRadio {
	var french []deezerRadio
	for _, r := range radios {
		if isFrenchRadio(r.Title) {
			french = append(french, r)
		}
	}

	// Si radios françaises existent, en prendre une + une aléatoire
	if len(french) > 0 {
		selected := []deezerRadio{french[rand.Intn(len(french))]}
		// Ajouter une radio aléatoire supplémentaire pour varier
		if len(radios) > 1 {
			selected = append(selected, radios[rand.Intn(len(radios))])
		}
		return selected
	}
	// Sinon, retourner une radio aléatoire
	return []deezerRadio{radios[rand.Intn(len(radios))]}
}

func FetchDeezerGenreTracks(ctx context.Context, playlistType string) ([]BlindtestTrack, error) {
	genreName, err := deezerGenreNameForType(playlistType)
	if err != nil {
		return nil, err
	}

	genreID, err := resolveDeezerGenreID(ctx, genreName)
	if err != nil {
		return nil, err
	}

	var radios deezerRadiosResp
	if err := deezerDoJSON(ctx, fmt.Sprintf("https://api.deezer.com/genre/%d/radios", genreID), &radios); err != nil {
		return nil, err
	}
	if len(radios.Data) == 0 {
		return nil, fmt.Errorf("aucune radio Deezer pour le genre %s", genreName)
	}

	// Sélectionner radios avec biais français
	selectedRadios := selectRadios(radios.Data)

	// Récupérer et fusionner les tracks
	trackMap := make(map[int64]BlindtestTrack)

	for _, radio := range selectedRadios {
		var tracksResp deezerRadioTracksResp
		if err := deezerDoJSON(ctx, fmt.Sprintf("https://api.deezer.com/radio/%d/tracks?limit=200", radio.ID), &tracksResp); err != nil {
			continue // Ignorer si une radio échoue
		}

		for _, t := range tracksResp.Data {
			if strings.TrimSpace(t.Preview) == "" {
				continue
			}
			// Ajouter si pas déjà vu (déduplication)
			if _, exists := trackMap[t.ID]; !exists {
				trackMap[t.ID] = BlindtestTrack{
					TrackID:    t.ID,
					PreviewURL: t.Preview,
					Title:      t.Title,
					Artist:     t.Artist.Name,
				}
			}
		}
	}

	// Convertir map en slice
	out := make([]BlindtestTrack, 0, len(trackMap))
	for _, track := range trackMap {
		out = append(out, track)
	}

	if len(out) == 0 {
		return nil, errors.New("aucun track avec preview trouvé")
	}
	return out, nil
}
