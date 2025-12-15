package server

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type PetitBacGame struct {
	roomID       int
	totalRounds  int
	timePerRound time.Duration

	mu      sync.Mutex
	phase   string // "idle"|"playing"|"validation"|"finished"
	round   int
	endsAt  time.Time
	letter  string
	answers map[int]map[int]string       // userID -> catID -> answer
	votes   map[int]map[int]map[int]bool // catID -> userID -> voterID -> bool
	timer   *time.Timer
}

var (
	petitBacGamesMu sync.Mutex
	petitBacGames   = map[int]*PetitBacGame{} // roomID -> game
)

func StartOrResetPetitBac(ctx context.Context, room *Room) (*PetitBacGame, error) {
	petitBacGamesMu.Lock()
	defer petitBacGamesMu.Unlock()

	if g, ok := petitBacGames[room.ID]; ok && g.timer != nil {
		g.timer.Stop()
	}

	game := &PetitBacGame{
		roomID:       room.ID,
		totalRounds:  room.Rounds,
		timePerRound: time.Duration(room.TimePerRound) * time.Second,
		phase:        "playing",
		round:        1,
		letter:       randomLetter(),
		answers:      map[int]map[int]string{},
		votes:        map[int]map[int]map[int]bool{},
	}
	game.endsAt = time.Now().Add(game.timePerRound)
	game.timer = time.AfterFunc(game.timePerRound, func() {
		game.onRoundEnd()
	})
	petitBacGames[room.ID] = game

	// Broadcast event spécifique pour permettre côté client une redirection vers /game/{code}
	getRoomHub(room.ID).broadcast <- mustJSON(WSMessage{
		Type: "petitbac_round_started",
		Payload: map[string]any{
			"room_id":      game.roomID,
			"round":        game.round,
			"total_rounds": game.totalRounds,
			"ends_at_unix": game.endsAt.Unix(),
			"letter":       game.letter,
		},
	})

	return game, nil
}

func randomLetter() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(letters[rand.Intn(len(letters))])
}

func (g *PetitBacGame) SubmitAnswers(userID int, answers map[int]string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != "playing" {
		return errors.New("not in playing phase")
	}
	g.answers[userID] = answers

	// Vérifier si ce joueur a rempli toutes les catégories
	allFilled := true
	for _, ans := range answers {
		if strings.TrimSpace(ans) == "" {
			allFilled = false
			break
		}
	}
	if allFilled {
		// Fin du tour pour tout le monde
		if g.timer != nil {
			g.timer.Stop()
			g.timer = nil
		}
		go g.onRoundEnd()
		return nil
	}

	// Sinon, attendre la fin du timer
	return nil
}

func (g *PetitBacGame) onRoundEnd() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != "playing" {
		return
	}
	g.phase = "validation"
	g.endsAt = time.Now().Add(30 * time.Second)
	g.votes = make(map[int]map[int]map[int]bool)
	BroadcastRoomUpdated(g.roomID)
	g.timer = time.AfterFunc(30*time.Second, func() {
		g.onValidationEnd()
	})
}

func (g *PetitBacGame) SubmitVotes(voterID int, votes map[int]map[int]bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != "validation" {
		return errors.New("not in validation phase")
	}
	for catID, userVotes := range votes {
		if g.votes[catID] == nil {
			g.votes[catID] = map[int]map[int]bool{}
		}
		for userID, valid := range userVotes {
			if g.votes[catID][userID] == nil {
				g.votes[catID][userID] = map[int]bool{}
			}
			g.votes[catID][userID][voterID] = valid
		}
	}
	return nil
}

func (g *PetitBacGame) onValidationEnd() {
	ctx := context.Background()
	nbPlayers := countPlayersInRoom(g.roomID)
	if nbPlayers == 0 {
		g.phase = "finished"
		return
	}

	// Pour chaque catégorie et chaque joueur, compter les votes
	for catID, userVotes := range g.votes {
		// Pour chaque joueur ayant donné une réponse dans cette catégorie
		for userID, voters := range userVotes {
			// Calcul du nombre de validations reçues
			nbVotes := 0
			for _, valid := range voters {
				if valid {
					nbVotes++
				}
			}
			// Seuil de validation : 2/3 arrondi supérieur
			seuil := (2*nbPlayers + 2) / 3
			isValid := nbVotes >= seuil

			// Récupérer la réponse du joueur
			answer := ""
			if g.answers[userID] != nil {
				answer = g.answers[userID][catID]
			}

			// Compter combien de joueurs ont donné cette réponse (pour l’unicité)
			count := 0
			for _, ansMap := range g.answers {
				if ansMap != nil && ansMap[catID] == answer && answer != "" {
					count++
				}
			}

			// Attribution des points
			points := 0
			if isValid && answer != "" {
				if count == 1 {
					points = 2
				} else {
					points = 1
				}
			}
			// Sinon points = 0

			// Ajoute le score
			_ = AddScore(ctx, g.roomID, userID, points)
		}
	}

	// Passage à la manche suivante ou fin de partie
	if g.round >= g.totalRounds {
		g.phase = "finished"
		// Broadcast fin de partie, etc.
	} else {
		// Préparer la prochaine manche
		g.round++
		g.phase = "playing"
		g.letter = randomLetter()
		g.answers = map[int]map[int]string{}
		g.votes = map[int]map[int]map[int]bool{}
		g.endsAt = time.Now().Add(g.timePerRound)
		g.timer = time.AfterFunc(g.timePerRound, func() {
			g.onRoundEnd()
		})
		BroadcastRoomUpdated(g.roomID)
	}
}

func countPlayersInRoom(roomID int) int {
	players, _ := ListRoomPlayers(context.Background(), roomID)
	return len(players)
}

func (g *PetitBacGame) StateForUser(userID int) map[string]any {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Récupérer les catégories dynamiquement
	categories, _ := ListPetitBacCategories(context.Background(), g.roomID)

	// Récupérer les scores (à partir des joueurs de la room)
	players, _ := ListRoomPlayers(context.Background(), g.roomID)
	scores := map[int]int{}
	for _, p := range players {
		scores[p.UserID] = p.Score
	}

	return map[string]any{
		"phase":       g.phase,
		"round":       g.round,
		"totalRounds": g.totalRounds,
		"endsAt":      g.endsAt.Unix(),
		"letter":      g.letter,
		"categories":  categories,
		"answers":     g.answers,
		"votes":       g.votes,
		"scores":      scores,
	}
}
