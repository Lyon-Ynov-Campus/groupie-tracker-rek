package server

import "context"

func AddScore(ctx context.Context, roomID, userID, delta int) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	_, err := Rekdb.ExecContext(ctx, SQLAddScoreToRoomPlayer, delta, roomID, userID)
	return err
}
