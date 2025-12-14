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

	// Choix radio: random pour varier
	radio := radios.Data[rand.Intn(len(radios.Data))]

	var tracksResp deezerRadioTracksResp
	if err := deezerDoJSON(ctx, fmt.Sprintf("https://api.deezer.com/radio/%d/tracks?limit=200", radio.ID), &tracksResp); err != nil {
		return nil, err
	}

	out := make([]BlindtestTrack, 0, len(tracksResp.Data))
	for _, t := range tracksResp.Data {
		if strings.TrimSpace(t.Preview) == "" {
			continue
		}
		out = append(out, BlindtestTrack{
			TrackID:    t.ID,
			PreviewURL: t.Preview,
			Title:      t.Title,
			Artist:     t.Artist.Name,
		})
	}
	if len(out) == 0 {
		return nil, errors.New("aucun track avec preview dans la radio")
	}
	return out, nil
}
