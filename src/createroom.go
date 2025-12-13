package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	roomCodeLength   = 6
	roomCodeAttempts = 20

	minRoomPlayers  = 2
	maxRoomPlayers  = 10
	minTimePerRound = 20
	minRounds       = 1
)

var (
	ErrDatabaseNotInitialised = errors.New("database not initialised")
	ErrRoomNotFound           = errors.New("room not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrRoomCapacityReached    = errors.New("room capacity reached")
	ErrPlayerAlreadyInRoom    = errors.New("player already in room")
	ErrInvalidRoomType        = errors.New("invalid room type")
	ErrInvalidRoomParameters  = errors.New("invalid room parameters")
)

type RoomType string

const (
	RoomTypeBlindTest RoomType = "blindtest"
	RoomTypePetitBac  RoomType = "petit_bac"
)

type Room struct {
	ID           int
	Code         string
	Type         RoomType
	CreatorID    int
	MaxPlayers   int
	TimePerRound int
	Rounds       int
	Status       string
}

type RoomPlayer struct {
	UserID  int
	Pseudo  string
	IsAdmin bool
	IsReady bool
	Score   int
}

type CreateRoomOptions struct {
	Type         RoomType
	CreatorID    int
	MaxPlayers   int
	TimePerRound int
	Rounds       int
}

type SallePageData struct {
	Room      *Room
	Players   []RoomPlayer
	GameLabel string
	IsAdmin   bool
}

func CreateRoom(ctx context.Context, opts CreateRoomOptions) (*Room, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}
	if err := validateCreateRoomOptions(opts); err != nil {
		return nil, err
	}

	tx, err := Rekdb.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Vérifie que l'utilisateur existe
	var creatorExists int
	if err := tx.QueryRowContext(ctx, SQLUserExistsByID, opts.CreatorID).Scan(&creatorExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var roomID int
	var code string

	for attempt := 0; attempt < roomCodeAttempts; attempt++ {
		code, err = randomRoomCode()
		if err != nil {
			return nil, err
		}

		res, err := tx.ExecContext(
			ctx,
			SQLInsertRoomLobby,
			code,
			string(opts.Type),
			opts.CreatorID,
			opts.MaxPlayers,
			opts.TimePerRound,
			opts.Rounds,
		)
		if err != nil {
			// collision UNIQUE(code) -> retry
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "unique") && strings.Contains(msg, "rooms.code") {
				continue
			}
			return nil, err
		}

		id64, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		roomID = int(id64)
		break
	}

	if roomID == 0 || code == "" {
		return nil, fmt.Errorf("unable to generate unique room code")
	}

	// Créateur ajouté comme admin
	if _, err := tx.ExecContext(ctx, SQLInsertRoomPlayerAdmin, roomID, opts.CreatorID); err != nil {
		return nil, err
	}

	room, err := getRoomByIDTx(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return room, nil
}

func GetRoomByCode(ctx context.Context, code string) (*Room, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, ErrRoomNotFound
	}

	var r Room
	var typ string
	err := Rekdb.QueryRowContext(ctx, SQLSelectRoomByCode, code).
		Scan(&r.ID, &r.Code, &typ, &r.CreatorID, &r.MaxPlayers, &r.TimePerRound, &r.Rounds, &r.Status)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	r.Type = RoomType(typ)
	return &r, nil
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

	room, err := getRoomByIDTx(ctx, tx, roomID)
	if err != nil {
		return nil, err
	}

	// Capacité
	var count int
	if err := tx.QueryRowContext(ctx, SQLCountRoomPlayersByRoomID, roomID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= room.MaxPlayers {
		return nil, ErrRoomCapacityReached
	}

	// déjà dedans ?
	var exists int
	err = tx.QueryRowContext(ctx, SQLRoomPlayerExists, roomID, userID).Scan(&exists)
	if err == nil {
		return nil, ErrPlayerAlreadyInRoom
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// user existe + pseudo
	var pseudo string
	if err := tx.QueryRowContext(ctx, SQLSelectUserPseudoByID, userID).Scan(&pseudo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	adminInt := 0
	if isAdmin {
		adminInt = 1
	}

	if _, err := tx.ExecContext(ctx, SQLInsertRoomPlayerMember, roomID, userID, adminInt); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrPlayerAlreadyInRoom
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &RoomPlayer{
		UserID:  userID,
		Pseudo:  pseudo,
		IsAdmin: isAdmin,
		IsReady: false,
		Score:   0,
	}, nil
}

func ListRoomPlayers(ctx context.Context, roomID int) ([]RoomPlayer, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}

	rows, err := Rekdb.QueryContext(ctx, SQLListRoomPlayersByRoomID, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []RoomPlayer
	for rows.Next() {
		var p RoomPlayer
		var adminInt, readyInt int
		if err := rows.Scan(&p.UserID, &p.Pseudo, &adminInt, &readyInt, &p.Score); err != nil {
			return nil, err
		}
		p.IsAdmin = adminInt == 1
		p.IsReady = readyInt == 1
		players = append(players, p)
	}
	return players, rows.Err()
}

func IsUserAdminInRoom(ctx context.Context, roomID, userID int) (bool, error) {
	if Rekdb == nil {
		return false, ErrDatabaseNotInitialised
	}
	var adminInt int
	if err := Rekdb.QueryRowContext(ctx, SQLSelectIsAdminInRoom, roomID, userID).Scan(&adminInt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return adminInt == 1, nil
}

func getRoomByIDTx(ctx context.Context, tx *sql.Tx, id int) (*Room, error) {
	var r Room
	var typ string
	err := tx.QueryRowContext(ctx, SQLSelectRoomByID, id).
		Scan(&r.ID, &r.Code, &typ, &r.CreatorID, &r.MaxPlayers, &r.TimePerRound, &r.Rounds, &r.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	r.Type = RoomType(typ)
	return &r, nil
}

func validateCreateRoomOptions(opts CreateRoomOptions) error {
	switch opts.Type {
	case RoomTypeBlindTest, RoomTypePetitBac:
	default:
		return ErrInvalidRoomType
	}
	if opts.CreatorID <= 0 {
		return ErrInvalidRoomParameters
	}
	if opts.MaxPlayers < minRoomPlayers || opts.MaxPlayers > maxRoomPlayers {
		return fmt.Errorf("%w: max_players must be between %d and %d", ErrInvalidRoomParameters, minRoomPlayers, maxRoomPlayers)
	}
	if opts.TimePerRound < minTimePerRound {
		return fmt.Errorf("%w: time_per_round too small", ErrInvalidRoomParameters)
	}
	if opts.Rounds < minRounds {
		return fmt.Errorf("%w: rounds too small", ErrInvalidRoomParameters)
	}
	return nil
}

var roomCodeAlphabet = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

func randomRoomCode() (string, error) {
	var b strings.Builder
	n := big.NewInt(int64(len(roomCodeAlphabet)))
	for i := 0; i < roomCodeLength; i++ {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		b.WriteByte(roomCodeAlphabet[idx.Int64()])
	}
	return b.String(), nil
}
