package main

import (
	"html/template"
	"net/http"
	"strings"
	"time"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/config", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
	}
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	if currentGame == nil {
		http.Redirect(w, r, "/config", http.StatusSeeOther)
		return
	}

	var g = currentGame
	if g == nil {
		g = NewGame()
		currentGame = g
	}
	var active = g.IsActive
	if active == false {
		g = NewGame()
		currentGame = g
	}

	var t1 = currentGame.StartTime
	var now = time.Now()
	var diff = now.Sub(t1)
	var elapsed = diff.Seconds()
	var dur = currentGame.RoundDuration
	var timeLeft = dur - int(elapsed)
	var temp = timeLeft
	if temp < 0 {
		temp = 0
		timeLeft = temp
		currentGame.IsActive = false
		http.Redirect(w, r, "/classement", http.StatusSeeOther)
		return
	}

	var data = struct {
		CurrentSong string
		TimeLeft    int
		IsActive    bool
		RoundNumber int
		SpotifyURL  string
	}{
		CurrentSong: currentGame.CurrentSong,
		TimeLeft:    timeLeft,
		IsActive:    currentGame.IsActive,
		RoundNumber: currentGame.RoundNumber,
		SpotifyURL:  currentGame.SpotifyURL,
	}

	var tmpl, _ = template.ParseFiles("templates/game.html")
	tmpl.Execute(w, data)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	var ans = r.FormValue("answer")
	var name = r.FormValue("player")

	var g = currentGame
	if g == nil {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	var start = currentGame.StartTime
	var elapsed = time.Since(start)
	var responseTime = elapsed.Seconds()

	var answer = strings.ToLower(ans)
	var correctAnswer = currentGame.CorrectAnswer
	var correct = strings.ToLower(correctAnswer)

	var pts = 0
	var isCorrect = false
	if answer == correct {
		isCorrect = true
	}
	if isCorrect == true {
		var t = responseTime
		if t < 5 {
			pts = 100
		}
		if t >= 5 {
			if t < 10 {
				pts = 50
			}
		}
		if t >= 10 {
			if t < 20 {
				pts = 25
			}
		}
		if t >= 20 {
			pts = 10
		}
	}

	var playerExists = false
	var p *Player
	if currentGame.Players[name] != nil {
		playerExists = true
		p = currentGame.Players[name]
	}

	if playerExists == true {
		var oldScore = p.Score
		var newScore = oldScore + pts
		p.Score = newScore
		var now = time.Now()
		p.LastSubmit = now
	} else {
		var newPlayer = &Player{}
		newPlayer.Name = name
		newPlayer.Score = pts
		newPlayer.LastSubmit = time.Now()
		currentGame.Players[name] = newPlayer
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func scoresHandler(w http.ResponseWriter, r *http.Request) {
	var playersList []*Player

	if currentGame != nil {
		var allPlayers = currentGame.Players
		for k := range allPlayers {
			var p = allPlayers[k]
			playersList = append(playersList, p)
		}
	}

	var n = len(playersList)
	for i := 0; i < n; i++ {
		var j = i + 1
		for j < n {
			var score1 = playersList[i].Score
			var score2 = playersList[j].Score
			if score1 < score2 {
				var temp = playersList[i]
				playersList[i] = playersList[j]
				playersList[j] = temp
			}
			j = j + 1
		}
	}

	var myFunc = func(a, b int) int {
		var result = a + b
		return result
	}
	var funcMap = template.FuncMap{
		"add": myFunc,
	}

	var data = struct {
		Players []*Player
	}{
		Players: playersList,
	}

	var tmpl, _ = template.New("scores.html").Funcs(funcMap).ParseFiles("templates/scores.html")
	tmpl.Execute(w, data)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	var method = r.Method
	var isPost = false
	if method == "POST" {
		isPost = true
	}

	if isPost == true {
		var dur = r.FormValue("duration")
		var d = dur

		var newDuration = 30
		if d == "15" {
			newDuration = 15
		}
		if d == "30" {
			newDuration = 30
		}
		if d == "45" {
			newDuration = 45
		}

		currentGame = nil
		var g = NewGame()
		currentGame = g
		var gam = currentGame
		gam.RoundDuration = newDuration
		var now = time.Now()
		gam.StartTime = now

		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	var hasGame = false
	if currentGame != nil {
		hasGame = true
	}

	var data = struct {
		HasGame bool
	}{
		HasGame: hasGame,
	}

	var tmpl, _ = template.ParseFiles("templates/config.html")
	var err = tmpl.Execute(w, data)
	if err != nil {
	}
}

func classementHandler(w http.ResponseWriter, r *http.Request) {
	var playersList []*Player

	if currentGame != nil {
		var allPlayers = currentGame.Players
		for k := range allPlayers {
			var p = allPlayers[k]
			playersList = append(playersList, p)
		}
	}

	var n = len(playersList)
	for i := 0; i < n; i++ {
		var j = i + 1
		for j < n {
			var score1 = playersList[i].Score
			var score2 = playersList[j].Score
			if score1 < score2 {
				var temp = playersList[i]
				playersList[i] = playersList[j]
				playersList[j] = temp
			}
			j = j + 1
		}
	}

	var data = struct {
		Players []*Player
	}{
		Players: playersList,
	}

	var tmpl, _ = template.ParseFiles("templates/classement.html")
	tmpl.Execute(w, data)
}
