package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BlindtestTrack représente une piste jouable
type BlindtestTrack struct {
	TrackID    int64
	PreviewURL string
	Title      string
	Artist     string
}

// Structure simplifiée pour lire la réponse JSON de Deezer
type deezerResp struct {
	Data []struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Preview string `json:"preview"`
		Artist  struct {
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"data"`
}

// Fonction utilitaire pour télécharger le JSON
func deezerGet(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// Choix intelligent de la playlist (Mélange FR/Inter inclus dans ces playlists)
func getSimpleQuery(genre string) string {
	switch strings.ToLower(strings.TrimSpace(genre)) {
	case "rock":
		return "Best of Rock"
	case "rap":
		return "Rap Hits"
	case "pop":
		return "Pop Hits"
	default:
		return "Top " + genre
	}
}

// FetchDeezerGenreTracks récupère 500 titres, mélange, et renvoie les 100 meilleurs

func FetchDeezerGenreTracks(ctx context.Context, playlistType string) ([]BlindtestTrack, error) {
	// la proicédure est la suivante : commence tout dabord
	// 1. Trouver la meilleure playlist pour le genre

	query := getSimpleQuery(playlistType)

	// On cherche la playlist la mieux notée (RATING_DESC) genre ce qui est dejà confirmé par les utilisateurs

	searchURL := fmt.Sprintf("https://api.deezer.com/search/playlist?q=%s&order=RATING_DESC&limit=1", url.QueryEscape(query))

	var searchResp deezerResp
	if err := deezerGet(ctx, searchURL, &searchResp); err != nil || len(searchResp.Data) == 0 {
		return nil, errors.New("aucune playlist trouvée")
	}

	bestPlaylistID := searchResp.Data[0].ID

	// 2. LE SECRET DE LA VARIÉTÉ : On récupère 500 chansons (le max) pour eviter que un joeurs capte les musique a force d'y jouer.

	tracksURL := fmt.Sprintf("https://api.deezer.com/playlist/%d/tracks?limit=500", bestPlaylistID)
	var tracksResp deezerResp
	if err := deezerGet(ctx, tracksURL, &tracksResp); err != nil {
		return nil, err
	}

	// 3. Filtrage : On garde uniquement celles avec un extrait audio pour pouvoir jouer en fonction du temps imparti configurer par l'administrateur de la salle

	var cleanTracks []BlindtestTrack
	seen := make(map[int64]bool)

	for _, t := range tracksResp.Data {
		if t.Preview == "" || seen[t.ID] {
			continue
		}
		seen[t.ID] = true

		cleanTracks = append(cleanTracks, BlindtestTrack{
			TrackID:    t.ID,
			PreviewURL: t.Preview,
			Title:      t.Title,
			Artist:     t.Artist.Name,
		})
	}

	if len(cleanTracks) == 0 {
		return nil, errors.New("aucune chanson jouable trouvée")
	}

	// 4. LE SHUFFLE : On mélange tout le paquet genre remier pour rendre la sélection aléatoire

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cleanTracks), func(i, j int) {
		cleanTracks[i], cleanTracks[j] = cleanTracks[j], cleanTracks[i]
	})

	// 5. LA COUPE : On ne garde que les 100 premières du mélange pour la partie pour éviter les répétitions
	if len(cleanTracks) > 100 {
		cleanTracks = cleanTracks[:100]
	}

	return cleanTracks, nil
}
