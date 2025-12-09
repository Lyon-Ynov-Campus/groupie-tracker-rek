package main

import (
	"log"
	"net/http"
)

func main() {
	//charger les credentials spotify
	InitSpotify()

	var handler1 = indexHandler
	http.HandleFunc("/", handler1)

	var handler2 = gameHandler
	http.HandleFunc("/game", handler2)
	http.HandleFunc("/salle-de-jeu", handler2)

	var handler3 = submitHandler
	http.HandleFunc("/submit", handler3)

	var handler4 = scoresHandler
	http.HandleFunc("/scores", handler4)

	var handler5 = configHandler
	http.HandleFunc("/config", handler5)

	var handler6 = classementHandler
	http.HandleFunc("/classement", handler6)

	//routes spotify
	var handler7 = spotifyLoginHandler
	http.HandleFunc("/spotify/login", handler7)

	var handler8 = spotifyCallbackHandler
	http.HandleFunc("/callback", handler8)

	var staticPath = "/static/"
	var staticDir = "static"
	var staticHandler = http.FileServer(http.Dir(staticDir))
	var stripped1 = http.StripPrefix(staticPath, staticHandler)
	http.Handle(staticPath, stripped1)

	var musicPath = "/music/"
	var musicDir = "static/music"
	var musicHandler = http.FileServer(http.Dir(musicDir))
	var stripped2 = http.StripPrefix(musicPath, musicHandler)
	http.Handle(musicPath, stripped2)

	var msg = "Serveur demarré sur http://localhost:8080"
	log.Println(msg)

	var port = ":8080"
	var err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
