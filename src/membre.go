package server

import (
	"context"
	"database/sql"
)

func IsUserInRoom(ctx context.Context, roomID, userID int) (bool, error) {
	if Rekdb == nil {
		return false, ErrDatabaseNotInitialised
	}
	var one int
	err := Rekdb.QueryRowContext(ctx, SQLRoomPlayerExists, roomID, userID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
