package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var tmpl = template.Must(template.ParseFiles("template/index.html", "template/salle.html", "template/jeu.html"))

var mu sync.RWMutex
var salles = make(map[string]bool)
var joueurs = make(map[string][]string)
var max = make(map[string]int)
var messages = make(map[string][]Message)
var logs = make(map[string][]string)
var counter int
var prets = make(map[string][]string)
var scoresPartie = make(map[string]map[string]int)
var scoresGeneral = make(map[string]int)

type Message struct {
	Author string
	Text   string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/entrer", entrerHandler)
	http.HandleFunc("/rejoindre", rejoindreHandler)
	http.HandleFunc("/salle", salleHandler)
	http.HandleFunc("/salle/pret", pretHandler)
	http.HandleFunc("/salle/message", messageHandler)
	http.HandleFunc("/salle/leave", leaveHandler)
	http.HandleFunc("/jeu", jeuHandler)
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/style.css")
	})

	log.Println("Serveur démarré sur :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func generateCode() string {
	letters := "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	var code string
	code = ""
	i := 0
	for i < 6 {
		n := rand.Intn(len(letters))
		code = code + string(letters[n])
		i = i + 1
	}
	return code
}

func entrerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var creator string
	creator = r.FormValue("creator_name")
	creator = strings.TrimSpace(creator)
	var maxPlayersStr string
	maxPlayersStr = r.FormValue("max_players")
	var maxPlayersInt int
	var err error
	maxPlayersInt, err = strconv.Atoi(maxPlayersStr)
	if err != nil {
		maxPlayersInt = 4
	}
	if maxPlayersInt < 2 {
		maxPlayersInt = 4
	}
	var code string
	code = generateCode()
	mu.Lock()
	salles[code] = true
	var tableauJoueurs []string
	tableauJoueurs = []string{}
	tableauJoueurs = append(tableauJoueurs, creator)
	joueurs[code] = tableauJoueurs
	max[code] = maxPlayersInt
	var tableauLogs []string
	tableauLogs = []string{}
	tableauLogs = append(tableauLogs, "Salle créée par: "+creator)
	logs[code] = tableauLogs
	var tableauPrets []string
	tableauPrets = []string{}
	prets[code] = tableauPrets
	var mapScores map[string]int
	mapScores = make(map[string]int)
	scoresPartie[code] = mapScores
	mu.Unlock()
	var url string
	url = "/salle?code=" + code + "&player=" + creator
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func rejoindreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var code string
	code = r.FormValue("room_code")
	code = strings.TrimSpace(code)
	var player string
	player = r.FormValue("player_name")
	player = strings.TrimSpace(player)
	mu.Lock()
	var salleExiste bool
	salleExiste = salles[code]
	if salleExiste == false {
		mu.Unlock()
		http.Error(w, "Salle introuvable", http.StatusNotFound)
		return
	}
	var ancienneListeJoueurs []string
	ancienneListeJoueurs = joueurs[code]
	ancienneListeJoueurs = append(ancienneListeJoueurs, player)
	joueurs[code] = ancienneListeJoueurs
	var ancienLogs []string
	ancienLogs = logs[code]
	var nouveauLog string
	nouveauLog = player + " a rejoint la salle"
	ancienLogs = append(ancienLogs, nouveauLog)
	logs[code] = ancienLogs
	mu.Unlock()
	var urlRedirect string
	urlRedirect = "/salle?code=" + code + "&player=" + player
	http.Redirect(w, r, urlRedirect, http.StatusSeeOther)
}

func salleHandler(w http.ResponseWriter, r *http.Request) {
	var code string
	code = r.URL.Query().Get("code")
	code = strings.TrimSpace(code)
	var player string
	player = r.URL.Query().Get("player")
	player = strings.TrimSpace(player)
	if code == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	mu.RLock()
	var listeJoueurs []string
	listeJoueurs = []string{}
	var i int
	i = 0
	for i < len(joueurs[code]) {
		var joueur string
		joueur = joueurs[code][i]
		listeJoueurs = append(listeJoueurs, joueur)
		i = i + 1
	}
	var maxJoueurs int
	maxJoueurs = max[code]
	var listeMessages []Message
	listeMessages = []Message{}
	var j int
	j = 0
	for j < len(messages[code]) {
		var msg Message
		msg = messages[code][j]
		listeMessages = append(listeMessages, msg)
		j = j + 1
	}
	var listeLogs []string
	listeLogs = []string{}
	var k int
	k = 0
	for k < len(logs[code]) {
		var logItem string
		logItem = logs[code][k]
		listeLogs = append(listeLogs, logItem)
		k = k + 1
	}
	var listePrets []string
	listePrets = []string{}
	var m int
	m = 0
	for m < len(prets[code]) {
		var pretItem string
		pretItem = prets[code][m]
		listePrets = append(listePrets, pretItem)
		m = m + 1
	}
	var scoresJoueurs map[string]int
	scoresJoueurs = scoresPartie[code]
	mu.RUnlock()
	type DataSalle struct {
		Code          string
		Players       []string
		MaxPlayers    int
		Player        string
		Messages      []Message
		Logs          []string
		Admin         string
		Prets         []string
		Scores        map[string]int
		ScoresGeneral map[string]int
	}
	var data DataSalle
	data.Code = code
	data.Players = listeJoueurs
	data.MaxPlayers = maxJoueurs
	data.Player = player
	data.Messages = listeMessages
	data.Logs = listeLogs
	data.Admin = ""
	data.Prets = listePrets
	data.Scores = scoresJoueurs
	data.ScoresGeneral = scoresGeneral
	var nombreJoueurs int
	nombreJoueurs = len(listeJoueurs)
	if nombreJoueurs > 0 {
		var premierJoueur string
		premierJoueur = listeJoueurs[0]
		data.Admin = premierJoueur
	}
	tmpl.ExecuteTemplate(w, "salle.html", data)
}

func pretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var code string
	code = r.FormValue("code")
	var player string
	player = r.FormValue("player")
	mu.Lock()
	var listePrets []string
	listePrets = prets[code]
	var dejaPret bool
	dejaPret = false
	var i int
	i = 0
	for i < len(listePrets) {
		if listePrets[i] == player {
			dejaPret = true
			break
		}
		i = i + 1
	}
	if dejaPret == false {
		listePrets = append(listePrets, player)
		prets[code] = listePrets
		var message string
		message = player + " est prêt(e)"
		logs[code] = append(logs[code], message)
	}
	var listeJoueurs []string
	listeJoueurs = joueurs[code]
	var nombreJoueurs int
	nombreJoueurs = len(listeJoueurs)
	var nombrePrets int
	nombrePrets = len(prets[code])
	var toutLeMondePret bool
	toutLeMondePret = false
	if nombrePrets == nombreJoueurs {
		if nombreJoueurs > 0 {
			toutLeMondePret = true
		}
	}
	mu.Unlock()
	if toutLeMondePret == true {
		var urlJeu string
		urlJeu = "/jeu?code=" + code + "&player=" + player
		http.Redirect(w, r, urlJeu, http.StatusSeeOther)
		return
	}
	var urlSalle string
	urlSalle = "/salle?code=" + code + "&player=" + player
	http.Redirect(w, r, urlSalle, http.StatusSeeOther)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var code string
	code = r.FormValue("code")
	var player string
	player = r.FormValue("player")
	var msg string
	msg = r.FormValue("message")
	msg = strings.TrimSpace(msg)
	var longueurMsg int
	longueurMsg = len(msg)
	if longueurMsg == 0 {
		var url string
		url = "/salle?code=" + code + "&player=" + player
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}
	mu.Lock()
	var nouveauMessage Message
	nouveauMessage = Message{}
	nouveauMessage.Author = player
	nouveauMessage.Text = msg
	var ancienMessages []Message
	ancienMessages = messages[code]
	ancienMessages = append(ancienMessages, nouveauMessage)
	messages[code] = ancienMessages
	var logMessage string
	logMessage = player + ": " + msg
	var ancienLogs []string
	ancienLogs = logs[code]
	ancienLogs = append(ancienLogs, logMessage)
	logs[code] = ancienLogs
	mu.Unlock()
	var urlRetour string
	urlRetour = "/salle?code=" + code + "&player=" + player
	http.Redirect(w, r, urlRetour, http.StatusSeeOther)
}

func leaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var code string
	code = r.FormValue("code")
	var player string
	player = r.FormValue("player")
	mu.Lock()
	var listeJoueurs []string
	listeJoueurs = joueurs[code]
	var indexASupprimer int
	indexASupprimer = -1
	var i int
	i = 0
	for i < len(listeJoueurs) {
		var joueurActuel string
		joueurActuel = listeJoueurs[i]
		if joueurActuel == player {
			indexASupprimer = i
			break
		}
		i = i + 1
	}
	var joueurTrouve bool
	joueurTrouve = false
	if indexASupprimer != -1 {
		joueurTrouve = true
	}
	if joueurTrouve == true {
		var nouvelleListe []string
		nouvelleListe = []string{}
		var j int
		j = 0
		for j < len(listeJoueurs) {
			if j != indexASupprimer {
				var joueurAGarder string
				joueurAGarder = listeJoueurs[j]
				nouvelleListe = append(nouvelleListe, joueurAGarder)
			}
			j = j + 1
		}
		joueurs[code] = nouvelleListe
		var messageLog string
		messageLog = player + " a quitté la salle"
		logs[code] = append(logs[code], messageLog)
	}
	var nombreJoueurs int
	nombreJoueurs = len(joueurs[code])
	var salleVide bool
	salleVide = false
	if nombreJoueurs == 0 {
		salleVide = true
	}
	if salleVide == true {
		delete(salles, code)
		delete(joueurs, code)
		delete(max, code)
		delete(messages, code)
		delete(logs, code)
	}
	mu.Unlock()
	var urlAccueil string
	urlAccueil = "/"
	http.Redirect(w, r, urlAccueil, http.StatusSeeOther)
}

func jeuHandler(w http.ResponseWriter, r *http.Request) {
	var code string
	code = r.URL.Query().Get("code")
	var player string
	player = r.URL.Query().Get("player")
	if code == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	mu.RLock()
	var listeJoueurs []string
	listeJoueurs = joueurs[code]
	var scoresJoueurs map[string]int
	scoresJoueurs = scoresPartie[code]
	mu.RUnlock()
	type DataJeu struct {
		Code    string
		Player  string
		Players []string
		Scores  map[string]int
	}
	var data DataJeu
	data.Code = code
	data.Player = player
	data.Players = listeJoueurs
	data.Scores = scoresJoueurs
	tmpl.ExecuteTemplate(w, "jeu.html", data)
}
