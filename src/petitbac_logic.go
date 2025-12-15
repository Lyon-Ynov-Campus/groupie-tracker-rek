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
	phase   string // "idle" | "playing" | "validation" | "finished"
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
		if g.timer != nil {
			g.timer.Stop()
			g.timer = nil
		}
		go g.onRoundEnd()
	}
	return nil
}

func (g *PetitBacGame) onRoundEnd() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != "playing" {
		return
	}

	// S'assurer que chaque joueur a une entrée dans g.answers (même vide)
	players, _ := ListRoomPlayers(context.Background(), g.roomID)
	categories, _ := ListPetitBacCategories(context.Background(), g.roomID)
	for _, p := range players {
		if g.answers[p.UserID] == nil {
			g.answers[p.UserID] = map[int]string{}
		}
		for _, cat := range categories {
			if _, ok := g.answers[p.UserID][cat.ID]; !ok {
				g.answers[p.UserID][cat.ID] = ""
			}
		}
	}

	g.phase = "validation"
	g.endsAt = time.Now().Add(30 * time.Second)
	g.votes = make(map[int]map[int]map[int]bool)
	BroadcastRoomUpdated(g.roomID)
	g.timer = time.AfterFunc(30*time.Second, func() {
		g.onValidationEnd()
	})
}

func (g *PetitBacGame) SubmitVotes(userID int, votes map[int]map[int]bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.phase != "validation" {
		return errors.New("not in validation phase")
	}
	for catID, userVotes := range votes {
		if g.votes[catID] == nil {
			g.votes[catID] = map[int]map[int]bool{}
		}
		for targetUserID, valid := range userVotes {
			if g.votes[catID][targetUserID] == nil {
				g.votes[catID][targetUserID] = map[int]bool{}
			}
			g.votes[catID][targetUserID][userID] = valid
		}
	}
	return nil
}

func (g *PetitBacGame) onValidationEnd() {
	g.mu.Lock()
	defer g.mu.Unlock()
	ctx := context.Background()
	nbPlayers := countPlayersInRoom(g.roomID)
	categories, _ := ListPetitBacCategories(ctx, g.roomID)

	for _, cat := range categories {
		catID := cat.ID
		for userID, userAnswers := range g.answers {
			answer := userAnswers[catID]
			// Compter les votes valides
			nbVotes := 0
			if g.votes[catID] != nil && g.votes[catID][userID] != nil {
				for _, valid := range g.votes[catID][userID] {
					if valid {
						nbVotes++
					}
				}
			}
			seuil := (2*nbPlayers + 2) / 3
			isValid := nbVotes >= seuil && answer != "" && strings.HasPrefix(strings.ToUpper(answer), g.letter)
			// Compter combien de joueurs ont donné cette réponse
			count := 0
			for _, ansMap := range g.answers {
				if ansMap[catID] == answer && answer != "" {
					count++
				}
			}
			points := 0
			if isValid {
				if count == 1 {
					points = 2
				} else {
					points = 1
				}
			}
			_ = AddScore(ctx, g.roomID, userID, points)
		}
	}

	if g.round >= g.totalRounds {
		g.phase = "finished"
		BroadcastRoomUpdated(g.roomID)
		return
	}

	// Manche suivante
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

func (g *PetitBacGame) StateForUser(userID int) map[string]any {
	g.mu.Lock()
	defer g.mu.Unlock()

	categories, _ := ListPetitBacCategories(context.Background(), g.roomID)
	players, _ := ListRoomPlayers(context.Background(), g.roomID)
	scores := map[int]int{}
	for _, p := range players {
		scores[p.UserID] = p.Score
	}

	var answers any
	switch g.phase {
	case "validation", "finished":
		answers = g.answers
	default:
		if g.answers[userID] != nil {
			answers = map[int]map[int]string{userID: g.answers[userID]}
		} else {
			answers = map[int]map[int]string{}
		}
	}

	return map[string]any{
		"phase":       g.phase,
		"round":       g.round,
		"totalRounds": g.totalRounds,
		"endsAt":      g.endsAt.Unix(),
		"letter":      g.letter,
		"categories":  categories,
		"answers":     answers,
		"votes":       g.votes,
		"scores":      scores,
		"players":     players,
	}
}

func countPlayersInRoom(roomID int) int {
	players, _ := ListRoomPlayers(context.Background(), roomID)
	return len(players)
}
