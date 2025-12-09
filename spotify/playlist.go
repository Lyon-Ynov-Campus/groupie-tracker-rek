package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Track c'est une musique
type Track struct {
	Name       string
	PreviewURL string
}

// reponse de spotify pour recommendations
type recommendationsResponse struct {
	Tracks []struct {
		Name       string `json:"name"`
		PreviewURL string `json:"preview_url"`
	} `json:"tracks"`
}

// seeds pour les genres
var genreSeeds = map[string]string{
	"rock": "rock",
	"rap":  "hip-hop",
	"pop":  "pop",
	"Rock": "rock",
	"Rap":  "hip-hop",
	"Pop":  "pop",
}

// GetTracksByGenre recupere des recommendations par genre
func GetTracksByGenre(genre string) ([]Track, error) {
	//obtenir le token
	token := GetAccessToken()
	if token == "" {
		return nil, fmt.Errorf("pas de token")
	}

	//obtenir le seed du genre
	seed := genreSeeds[genre]
	if seed == "" {
		seed = "pop" //par defaut
	}

	//utiliser l'endpoint recommendations
	url := fmt.Sprintf("https://api.spotify.com/v1/recommendations?seed_genres=%s&limit=50", seed)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erreur %d", resp.StatusCode)
	}

	var data recommendationsResponse
	json.NewDecoder(resp.Body).Decode(&data)

	var tracks []Track
	//on filtre ceux qui ont pas de preview
	for _, item := range data.Tracks {
		if item.PreviewURL != "" {
			tracks = append(tracks, Track{
				Name:       item.Name,
				PreviewURL: item.PreviewURL,
			})
		}
	}

	return tracks, nil
}

// Pour compatibilité avec l'ancien code
func GetPlaylistTracksByGenre(genre string) ([]Track, error) {
	return GetTracksByGenre(genre)
}
