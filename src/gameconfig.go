package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type PetitBacCategory struct {
	ID       int
	Name     string
	Position int
}

func GetBlindtestPlaylist(ctx context.Context, roomID int) (string, bool, error) {
	if Rekdb == nil {
		return "", false, ErrDatabaseNotInitialised
	}
	var playlist string
	err := Rekdb.QueryRowContext(ctx, SQLSelectBlindtestPlaylistByRoomID, roomID).Scan(&playlist)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return playlist, true, nil
}

func SetBlindtestPlaylist(ctx context.Context, roomID int, playlist string) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	playlist = strings.TrimSpace(playlist)
	if playlist == "" {
		return errors.New("playlist requise")
	}
	_, err := Rekdb.ExecContext(ctx, SQLUpsertBlindtestPlaylist, roomID, playlist)
	return err
}

func ListPetitBacCategories(ctx context.Context, roomID int) ([]PetitBacCategory, error) {
	if Rekdb == nil {
		return nil, ErrDatabaseNotInitialised
	}
	rows, err := Rekdb.QueryContext(ctx, SQLListPetitBacCategoriesByRoomID, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PetitBacCategory
	for rows.Next() {
		var c PetitBacCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Position); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func EnsureDefaultPetitBacCategories(ctx context.Context, roomID int) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	var count int
	if err := Rekdb.QueryRowContext(ctx, SQLCountPetitBacCategoriesByRoomID, roomID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaults := []string{
		"Artiste",
		"Album",
		"Groupe de musique",
		"Instrument de musique",
		"Featuring",
	}

	for i, name := range defaults {
		if _, err := Rekdb.ExecContext(ctx, SQLInsertPetitBacCategory, roomID, name, i+1); err != nil {
			return err
		}
	}
	return nil
}

func AddPetitBacCategory(ctx context.Context, roomID int, name string) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nom de catégorie requis")
	}

	var count int
	if err := Rekdb.QueryRowContext(ctx, SQLCountPetitBacCategoriesByRoomID, roomID).Scan(&count); err != nil {
		return err
	}
	_, err := Rekdb.ExecContext(ctx, SQLInsertPetitBacCategory, roomID, name, count+1)
	return err
}

func UpdatePetitBacCategoryName(ctx context.Context, roomID, id int, name string) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nom de catégorie requis")
	}
	_, err := Rekdb.ExecContext(ctx, SQLUpdatePetitBacCategory, name, id, roomID)
	return err
}

func DeletePetitBacCategory(ctx context.Context, roomID, id int) error {
	if Rekdb == nil {
		return ErrDatabaseNotInitialised
	}
	_, err := Rekdb.ExecContext(ctx, SQLDeletePetitBacCategory, id, roomID)
	return err
}
