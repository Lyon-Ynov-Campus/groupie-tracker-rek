package server

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type BlindtestGame struct {
	roomID       int
	totalRounds  int
	timePerRound time.Duration

	mu       sync.Mutex
	phase    string // "idle"|"playing"|"reveal"|"finished"
	round    int
	endsAt   time.Time
	current  BlindtestTrack
	attempts map[int]bool // userID -> attempted
	tracks   []BlindtestTrack
	used     map[int64]bool
	timer    *time.Timer
}

var (
	blindtestGamesMu sync.Mutex
	blindtestGames   = map[int]*BlindtestGame{} // roomID -> game
)

func GetBlindtestGame(roomID int) (*BlindtestGame, bool) {
	blindtestGamesMu.Lock()
	defer blindtestGamesMu.Unlock()
	g, ok := blindtestGames[roomID]
	return g, ok
}

func StartOrResetBlindtest(ctx context.Context, room *Room, playlistType string) (*BlindtestGame, error) {
	tracks, err := FetchDeezerGenreTracks(ctx, playlistType)
	if err != nil {
		return nil, err
	}

	// stop propre d'une ancienne partie si elle existe
	blindtestGamesMu.Lock()
	if old, ok := blindtestGames[room.ID]; ok {
		old.mu.Lock()
		if old.timer != nil {
			old.timer.Stop()
			old.timer = nil
		}
		old.mu.Unlock()
	}
	blindtestGamesMu.Unlock()

	g := &BlindtestGame{
		roomID:       room.ID,
		totalRounds:  room.Rounds,
		timePerRound: time.Duration(room.TimePerRound) * time.Second,
		phase:        "playing",
		round:        0,
		attempts:     map[int]bool{},
		tracks:       tracks,
		used:         map[int64]bool{},
	}

	blindtestGamesMu.Lock()
	blindtestGames[room.ID] = g
	blindtestGamesMu.Unlock()

	g.mu.Lock()
	defer g.mu.Unlock()
	g.startNextRoundLocked()
	return g, nil
}

func (g *BlindtestGame) startNextRoundLocked() {
	if g.timer != nil {
		g.timer.Stop()
		g.timer = nil
	}

	g.round++
	if g.round > g.totalRounds {
		g.phase = "finished"
		getRoomHub(g.roomID).broadcast <- mustJSON(WSMessage{Type: "blindtest_finished"})
		return
	}

	g.phase = "playing"
	g.attempts = map[int]bool{}

	candidates := make([]BlindtestTrack, 0, len(g.tracks))
	for _, t := range g.tracks {
		if !g.used[t.TrackID] {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) == 0 {
		g.used = map[int64]bool{}
		candidates = g.tracks
	}

	pick := candidates[rand.Intn(len(candidates))]
	g.current = pick
	g.used[pick.TrackID] = true

	g.endsAt = time.Now().Add(g.timePerRound)

	// On n'envoie JAMAIS title/artist ici
	getRoomHub(g.roomID).broadcast <- mustJSON(WSMessage{
		Type: "blindtest_round_started",
		Payload: map[string]any{
			"room_id":      g.roomID,
			"round":        g.round,
			"total_rounds": g.totalRounds,
			"ends_at_unix": g.endsAt.Unix(),
			"preview_url":  g.current.PreviewURL,
		},
	})

	g.timer = time.AfterFunc(g.timePerRound, func() {
		g.onRoundEnd()
	})
}

func (g *BlindtestGame) onRoundEnd() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != "playing" {
		return
	}
	g.phase = "reveal"

	// Reveal seulement fin de timer
	getRoomHub(g.roomID).broadcast <- mustJSON(WSMessage{
		Type: "blindtest_round_reveal",
		Payload: map[string]any{
			"title":  g.current.Title,
			"artist": g.current.Artist,
		},
	})

	// next round après une mini pause
	g.timer = time.AfterFunc(3*time.Second, func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.startNextRoundLocked()
	})
}

func (g *BlindtestGame) StateForUser(userID int) map[string]any {
	g.mu.Lock()
	defer g.mu.Unlock()

	st := map[string]any{
		"phase":         g.phase,
		"round":         g.round,
		"total_rounds":  g.totalRounds,
		"ends_at_unix":  g.endsAt.Unix(),
		"preview_url":   "",
		"already_tried": g.attempts[userID],
	}

	if g.phase == "playing" {
		st["preview_url"] = g.current.PreviewURL
	}
	if g.phase == "reveal" || g.phase == "finished" {
		st["title"] = g.current.Title
		st["artist"] = g.current.Artist
	}
	return st
}

func (g *BlindtestGame) SubmitGuess(ctx context.Context, roomID, userID int, guess string) (map[string]any, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.phase != "playing" || time.Now().After(g.endsAt) {
		return map[string]any{"locked": true}, nil
	}
	if g.attempts[userID] {
		return map[string]any{"already_tried": true}, nil
	}
	g.attempts[userID] = true

	correct := isCorrectGuess(guess, g.current.Title, g.current.Artist)
	points := 0

	if correct {
		remaining := int(time.Until(g.endsAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		points = remaining

		if points > 0 {
			_ = AddScore(ctx, roomID, userID, points)
			BroadcastRoomUpdated(roomID) // refresh scoreboard
		}
	}

	return map[string]any{
		"correct":        correct,
		"points_awarded": points,
		"locked":         false,
		"already_tried":  true,
	}, nil
}
