package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Game struct {
	Players       map[string]*Player
	CurrentSong   string
	StartTime     time.Time
	RoundDuration int
	IsActive      bool
	CorrectAnswer string
	RoundNumber   int
	SpotifyURL    string
}

type Player struct {
	Name       string
	Score      int
	LastSubmit time.Time
}

type SpotifyTrack struct {
	Name string `json:"name"`
	URL  string `json:"preview_url"`
}

type SpotifyResponse struct {
	Tracks struct {
		Items []struct {
			Track SpotifyTrack `json:"track"`
		} `json:"items"`
	} `json:"tracks"`
}

var currentGame *Game
var token = "BQC8..."

func GetSpotifySong() (string, string, string) {
	var url = "https://api.spotify.com/v1/playlists/37i9dQZF1DXcBWIGoYBM5M/tracks?limit=50"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "song1.mp3", "Bohemian Rhapsody", ""
	}
	req.Header.Add("Authorization", "Bearer "+token)

	var client = &http.Client{}
	resp, err2 := client.Do(req)
	if err2 != nil {
		fmt.Println("erreur")
		return "song1.mp3", "Bohemian Rhapsody", ""
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var data SpotifyResponse
	err3 := json.Unmarshal(body, &data)
	if err3 != nil {
		return "song1.mp3", "Test", ""
	}

	var nbItems = len(data.Tracks.Items)
	if nbItems > 0 {
		var idx = time.Now().Second()
		var index = idx % nbItems
		var track = data.Tracks.Items[index].Track
		var songname = track.Name + ".mp3"
		return songname, track.Name, track.URL
	}

	return "song1.mp3", "Test", ""
}

func NewGame() *Game {
	var song, ans, url = GetSpotifySong()

	var game = &Game{
		Players:       make(map[string]*Player),
		CurrentSong:   song,
		StartTime:     time.Now(),
		RoundDuration: 30,
		IsActive:      true,
		CorrectAnswer: ans,
		RoundNumber:   1,
		SpotifyURL:    url,
	}
	return game
}
