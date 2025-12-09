package main

import (
	"fmt"
	"os"
	"rek/spotify"
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

var currentGame *Game
var tracks []spotify.Track
var currentTrackIndex int

// recuperer les musiques spotify
func InitSpotify() {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	spotify.SetCredentials(clientID, clientSecret)

	//verifier si on a un token
	if !spotify.IsAuthenticated() {
		fmt.Println("Pas de token Spotify. L'utilisateur doit se connecter.")
		return
	}

	//recuperer playlist rock par defaut
	tracks, err := spotify.GetPlaylistTracksByGenre("Rock")
	if err != nil {
		fmt.Println("Erreur tracks:", err)
		return
	}

	fmt.Printf("Chargé %d musiques Spotify\n", len(tracks))
	currentTrackIndex = 0
}

func GetSpotifySong() (string, string, string) {
	if len(tracks) == 0 {
		return "song1.mp3", "Test", ""
	}

	//prendre la musique suivante
	track := tracks[currentTrackIndex]
	currentTrackIndex++
	if currentTrackIndex >= len(tracks) {
		currentTrackIndex = 0
	}

	return track.Name + ".mp3", track.Name, track.PreviewURL
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
