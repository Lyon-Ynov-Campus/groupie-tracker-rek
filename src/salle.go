package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	minRoomPlayers   = 2
	maxRoomPlayers   = 10
	minTimePerRound  = 10
	maxTimePerRound  = 120
	minRounds        = 1
	maxRounds        = 20
	roomCodeLength   = 6
	roomCodeAttempts = 10
)

var (
	ErrDatabaseNotInitialised = errors.New("database not initialised")
	ErrRoomNotFound           = errors.New("room not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrRoomCapacityReached    = errors.New("room capacity reached")
	ErrPlayerAlreadyInRoom    = errors.New("player already in room")
	ErrInvalidRoomType        = errors.New("invalid room type")
	ErrInvalidRoomParameters  = errors.New("invalid room parameters")
	ErrCategoryAlreadyExists  = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrPlayerNotInRoom        = errors.New("player not in room")
)

var defaultPetitBacCategories = []string{
	"Artistes",
	"Albums",
	"Groupes",
	"Instruments",
	"Featuring",
}

type RoomType string

const (
	RoomTypeBlindTest RoomType = "blindtest"
	RoomTypePetitBac  RoomType = "petit_bac"
)

type RoomStatus string

const (
	RoomStatusLobby    RoomStatus = "lobby"
	RoomStatusInGame   RoomStatus = "in_game"
	RoomStatusFinished RoomStatus = "finished"
)

type Room struct {
	ID           int
	Code         string
	Type         RoomType
	CreatorID    int
	MaxPlayers   int
	TimePerRound int
	Rounds       int
	Status       RoomStatus
	CreatedAt    time.Time
}

type RoomPlayer struct {
	ID       int
	RoomID   int
	UserID   int
	Pseudo   string
	IsAdmin  bool
	IsReady  bool
	Score    int
	JoinedAt time.Time
}

type CreateRoomOptions struct {
	Type         RoomType
	CreatorID    int
	MaxPlayers   int
	TimePerRound int
	Rounds       int
	PlaylistID   string
	PlaylistName string
	Categories   []string
}

func CreateRoom(ctx context.Context, opts CreateRoomOptions) (*Room, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	sanitisedCategories, err := validateCreateRoomOptions(&opts)
	if err != nil {
		return nil, err
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	roomCode, err := generateUniqueRoomCode(ctx, tx)
	if err != nil {
		return nil, err
	}

	var creatorPseudo string
	err = tx.QueryRowContext(ctx, "SELECT pseudo FROM users WHERE id = ?", opts.CreatorID).Scan(&creatorPseudo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO rooms (code, type, creator_id, max_players, time_per_round, rounds, status) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		roomCode,
		string(opts.Type),
		opts.CreatorID,
		opts.MaxPlayers,
		opts.TimePerRound,
		opts.Rounds,
		string(RoomStatusLobby),
	)
	if err != nil {
		return nil, err
	}

	roomID64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	roomID := int(roomID64)

	switch opts.Type {
	case RoomTypeBlindTest:
		playlistName := sql.NullString{String: opts.PlaylistName, Valid: strings.TrimSpace(opts.PlaylistName) != ""}
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO room_blindtest (room_id, playlist_id, playlist_name) VALUES (?, ?, ?)`,
			roomID,
			opts.PlaylistID,
			playlistName,
		); err != nil {
			return nil, err
		}

	case RoomTypePetitBac:
		for _, label := range sanitisedCategories {
			if _, err = tx.ExecContext(
				ctx,
				`INSERT INTO room_petitbac_categories (room_id, label, is_active) VALUES (?, ?, 1)`,
				roomID,
				label,
			); err != nil {
				return nil, err
			}
		}
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO room_players (room_id, user_id, pseudo, is_admin, is_ready, score) VALUES (?, ?, ?, 1, 0, 0)`,
		roomID,
		opts.CreatorID,
		creatorPseudo,
	); err != nil {
		return nil, err
	}

	room, err := fetchRoomByID(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return room, nil
}

func AddRoomPlayer(ctx context.Context, roomID, userID int, isAdmin bool) (*RoomPlayer, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	room, err := fetchRoomByID(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != RoomStatusLobby {
		return nil, fmt.Errorf("room %s is not open for joining", room.Status)
	}

	var playerCount int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM room_players WHERE room_id = ?", roomID).Scan(&playerCount); err != nil {
		return nil, err
	}
	if playerCount >= room.MaxPlayers {
		return nil, ErrRoomCapacityReached
	}

	var existing int
	err = tx.QueryRowContext(ctx, "SELECT 1 FROM room_players WHERE room_id = ? AND user_id = ?", roomID, userID).Scan(&existing)
	if err == nil {
		return nil, ErrPlayerAlreadyInRoom
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var pseudo string
	err = tx.QueryRowContext(ctx, "SELECT pseudo FROM users WHERE id = ?", userID).Scan(&pseudo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	adminFlag := 0
	if isAdmin {
		adminFlag = 1
	}

	insertResult, err := tx.ExecContext(
		ctx,
		`INSERT INTO room_players (room_id, user_id, pseudo, is_admin, is_ready, score) VALUES (?, ?, ?, ?, 0, 0)`,
		roomID,
		userID,
		pseudo,
		adminFlag,
	)
	if err != nil {
		return nil, err
	}

	playerID64, err := insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	player, err := fetchRoomPlayerByID(ctx, tx, int(playerID64))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return player, nil
}

func ListRoomPlayers(ctx context.Context, roomID int) ([]RoomPlayer, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	rows, err := Rekdb.QueryContext(
		ctx,
		`SELECT id, room_id, user_id, pseudo, is_admin, is_ready, score, joined_at 
         FROM room_players 
         WHERE room_id = ? 
         ORDER BY joined_at ASC, id ASC`,
		roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []RoomPlayer
	for rows.Next() {
		var (
			player     RoomPlayer
			isAdminInt int
			isReadyInt int
			joinedAt   time.Time
		)

		if err = rows.Scan(
			&player.ID,
			&player.RoomID,
			&player.UserID,
			&player.Pseudo,
			&isAdminInt,
			&isReadyInt,
			&player.Score,
			&joinedAt,
		); err != nil {
			return nil, err
		}

		player.IsAdmin = isAdminInt == 1
		player.IsReady = isReadyInt == 1
		player.JoinedAt = joinedAt

		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}

func GetRoomByCode(ctx context.Context, code string) (*Room, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	trimmed := strings.ToUpper(strings.TrimSpace(code))
	if trimmed == "" {
		return nil, ErrRoomNotFound
	}

	row := Rekdb.QueryRowContext(
		ctx,
		`SELECT id, code, type, creator_id, max_players, time_per_round, rounds, status, created_at 
         FROM rooms WHERE code = ?`,
		trimmed,
	)

	var (
		room      Room
		typeStr   string
		statusStr string
	)

	err := row.Scan(
		&room.ID,
		&room.Code,
		&typeStr,
		&room.CreatorID,
		&room.MaxPlayers,
		&room.TimePerRound,
		&room.Rounds,
		&statusStr,
		&room.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	room.Type = RoomType(typeStr)
	room.Status = RoomStatus(statusStr)

	return &room, nil
}

func GetRoomByID(ctx context.Context, roomID int) (*Room, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	return fetchRoomByID(ctx, Rekdb, roomID)
}

func fetchRoomByID(ctx context.Context, executor queryer, roomID int) (*Room, error) {
	row := executor.QueryRowContext(
		ctx,
		`SELECT id, code, type, creator_id, max_players, time_per_round, rounds, status, created_at 
         FROM rooms WHERE id = ?`,
		roomID,
	)

	var (
		room      Room
		typeStr   string
		statusStr string
	)

	err := row.Scan(
		&room.ID,
		&room.Code,
		&typeStr,
		&room.CreatorID,
		&room.MaxPlayers,
		&room.TimePerRound,
		&room.Rounds,
		&statusStr,
		&room.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	room.Type = RoomType(typeStr)
	room.Status = RoomStatus(statusStr)

	return &room, nil
}

func fetchRoomPlayerByID(ctx context.Context, executor queryer, playerID int) (*RoomPlayer, error) {
	row := executor.QueryRowContext(
		ctx,
		`SELECT id, room_id, user_id, pseudo, is_admin, is_ready, score, joined_at 
         FROM room_players WHERE id = ?`,
		playerID,
	)

	var (
		player     RoomPlayer
		isAdminInt int
		isReadyInt int
	)

	err := row.Scan(
		&player.ID,
		&player.RoomID,
		&player.UserID,
		&player.Pseudo,
		&isAdminInt,
		&isReadyInt,
		&player.Score,
		&player.JoinedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	player.IsAdmin = isAdminInt == 1
	player.IsReady = isReadyInt == 1

	return &player, nil
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func validateCreateRoomOptions(opts *CreateRoomOptions) ([]string, error) {
	if opts == nil {
		return nil, ErrInvalidRoomParameters
	}

	switch opts.Type {
	case RoomTypeBlindTest, RoomTypePetitBac:
	default:
		return nil, ErrInvalidRoomType
	}

	if opts.CreatorID <= 0 {
		return nil, fmt.Errorf("%w: creator id must be positive", ErrInvalidRoomParameters)
	}

	if opts.MaxPlayers < minRoomPlayers || opts.MaxPlayers > maxRoomPlayers {
		return nil, fmt.Errorf("%w: max players must be between %d and %d", ErrInvalidRoomParameters, minRoomPlayers, maxRoomPlayers)
	}

	if opts.TimePerRound < minTimePerRound || opts.TimePerRound > maxTimePerRound {
		return nil, fmt.Errorf("%w: time per round must be between %d and %d", ErrInvalidRoomParameters, minTimePerRound, maxTimePerRound)
	}

	if opts.Rounds < minRounds || opts.Rounds > maxRounds {
		return nil, fmt.Errorf("%w: rounds must be between %d and %d", ErrInvalidRoomParameters, minRounds, maxRounds)
	}

	opts.PlaylistID = strings.TrimSpace(opts.PlaylistID)
	opts.PlaylistName = strings.TrimSpace(opts.PlaylistName)

	switch opts.Type {
	case RoomTypeBlindTest:
		if opts.PlaylistID == "" {
			return nil, fmt.Errorf("%w: playlist id is required for blind test rooms", ErrInvalidRoomParameters)
		}
	case RoomTypePetitBac:
		categories := sanitiseCategories(opts.Categories)
		if len(categories) == 0 {
			categories = sanitiseCategories(defaultPetitBacCategories)
		}
		if len(categories) == 0 {
			return nil, fmt.Errorf("%w: at least one category is required for petit bac rooms", ErrInvalidRoomParameters)
		}
		return categories, nil
	}

	return nil, nil
}

func sanitiseCategories(labels []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, label := range labels {
		trimmed := strings.TrimSpace(label)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}

func generateUniqueRoomCode(ctx context.Context, tx *sql.Tx) (string, error) {
	for attempt := 0; attempt < roomCodeAttempts; attempt++ {
		code, err := randomRoomCode()
		if err != nil {
			return "", err
		}

		var exists int
		err = tx.QueryRowContext(ctx, "SELECT 1 FROM rooms WHERE code = ? LIMIT 1", code).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("unable to generate a unique room code after %d attempts", roomCodeAttempts)
}

var roomCodeAlphabet = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

func randomRoomCode() (string, error) {
	var builder strings.Builder
	alphabetLength := big.NewInt(int64(len(roomCodeAlphabet)))

	for i := 0; i < roomCodeLength; i++ {
		index, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", err
		}
		builder.WriteByte(roomCodeAlphabet[index.Int64()])
	}

	return builder.String(), nil
}

type BlindTestConfig struct {
	PlaylistID   string
	PlaylistName string
}

func GetBlindTestConfig(ctx context.Context, roomID int) (*BlindTestConfig, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}
	var (
		playlistID   string
		playlistName sql.NullString
	)
	err := Rekdb.QueryRowContext(
		ctx,
		`SELECT playlist_id, playlist_name FROM room_blindtest WHERE room_id = ?`,
		roomID,
	).Scan(&playlistID, &playlistName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(playlistName.String)
	if name == "" {
		name = playlistID
	}
	return &BlindTestConfig{
		PlaylistID:   playlistID,
		PlaylistName: name,
	}, nil
}

type PetitBacCategory struct {
	ID       int
	Label    string
	IsActive bool
}

func AddPetitBacCategory(ctx context.Context, roomID int, label string) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}

	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("%w: label cannot be empty", ErrInvalidRoomParameters)
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM room_petitbac_categories WHERE room_id = ? AND LOWER(label) = LOWER(?)`, roomID, label).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryAlreadyExists
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO room_petitbac_categories (room_id, label, is_active) VALUES (?, ?, 1)`, roomID, label)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func TogglePetitBacCategory(ctx context.Context, roomID, categoryID int) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var isActive int
	err = tx.QueryRowContext(ctx, `SELECT is_active FROM room_petitbac_categories WHERE id = ? AND room_id = ?`, categoryID, roomID).Scan(&isActive)
	if err != nil {
		return err
	}

	newStatus := 1 - isActive
	_, err = tx.ExecContext(ctx, `UPDATE room_petitbac_categories SET is_active = ? WHERE id = ? AND room_id = ?`, newStatus, categoryID, roomID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func DeletePetitBacCategory(ctx context.Context, roomID, categoryID int) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM room_petitbac_categories WHERE id = ? AND room_id = ?`, categoryID, roomID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func ListPetitBacCategories(ctx context.Context, roomID int) ([]PetitBacCategory, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}
	rows, err := Rekdb.QueryContext(
		ctx,
		`SELECT id, label, is_active FROM room_petitbac_categories 
         WHERE room_id = ? 
         ORDER BY id ASC`,
		roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []PetitBacCategory
	for rows.Next() {
		var category PetitBacCategory
		var isActiveInt int

		if err := rows.Scan(&category.ID, &category.Label, &isActiveInt); err != nil {
			return nil, err
		}

		category.IsActive = isActiveInt == 1
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func IsUserAdminInRoom(ctx context.Context, roomID, userID int) (bool, error) {
	if Rekdb == nil {
		return false, ErrDatabaseNotInitialised
	}

	var isAdmin int
	err := Rekdb.QueryRowContext(
		ctx,
		`SELECT is_admin FROM room_players WHERE room_id = ? AND user_id = ?`,
		roomID, userID,
	).Scan(&isAdmin)

	if err != nil {
		return false, err
	}

	return isAdmin == 1, nil
}
