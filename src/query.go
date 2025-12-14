package server

// TablesSQL contient toutes les requêtes de création de tables
var TablesSQL = map[string]string{
	"users": `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pseudo TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`,
	"rooms": `CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		creator_id INTEGER NOT NULL,
		max_players INTEGER NOT NULL,
		time_per_round INTEGER NOT NULL,
		rounds INTEGER NOT NULL,
		status TEXT NOT NULL
	);`,
	"room_players": `CREATE TABLE IF NOT EXISTS room_players (
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		is_admin INTEGER DEFAULT 0,
		is_ready INTEGER DEFAULT 0,
		score INTEGER DEFAULT 0,
		PRIMARY KEY (room_id, user_id)
	);`,
	"room_blindtest_settings": `CREATE TABLE IF NOT EXISTS room_blindtest_settings (
    room_id INTEGER PRIMARY KEY,
    playlist TEXT NOT NULL
);`,
	"room_petitbac_categories": `CREATE TABLE IF NOT EXISTS room_petitbac_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    position INTEGER NOT NULL
);`,
}

// Fonction pour inserer les données d'un nouvel utilisateur dans la base de données
//variable IndexesSQL contient les requêtes de création d'index LIANT les tables entre elles

var IndexesSQL = map[string]string{
	"idx_rooms_owner":                  "CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(creator_id);",
	"idx_room_players_room":            "CREATE INDEX IF NOT EXISTS idx_room_players_room ON room_players(room_id);",
	"idx_blindtest_settings_room":      "CREATE INDEX IF NOT EXISTS idx_blindtest_settings_room ON room_blindtest_settings(room_id);",
	"idx_petitbac_categories_room":     "CREATE INDEX IF NOT EXISTS idx_petitbac_categories_room ON room_petitbac_categories(room_id);",
	"idx_petitbac_categories_room_pos": "CREATE UNIQUE INDEX IF NOT EXISTS idx_petitbac_categories_room_pos ON room_petitbac_categories(room_id, position);",
}

func InsertValuesUser(pseudo, email, passwordHash string) error {
	_, err := Rekdb.Exec("INSERT INTO users (pseudo, email, password_hash) VALUES (?, ?, ?)", pseudo, email, passwordHash)
	return err
}

// PRAGMA
const SQLPragmaForeignKeysOn = `PRAGMA foreign_keys = ON`

// Rooms / Players / Users queries
const (
	SQLUserExistsByID = `SELECT 1 FROM users WHERE id = ? LIMIT 1`

	SQLInsertRoomLobby = `
        INSERT INTO rooms (code, type, creator_id, max_players, time_per_round, rounds, status)
        VALUES (?, ?, ?, ?, ?, ?, 'lobby')
    `

	SQLInsertRoomPlayerAdmin = `
        INSERT INTO room_players (room_id, user_id, is_admin, is_ready, score)
        VALUES (?, ?, 1, 0, 0)
    `

	SQLSelectRoomByID = `
        SELECT id, code, type, creator_id, max_players, time_per_round, rounds, status
        FROM rooms
        WHERE id = ?
    `

	SQLSelectRoomByCode = `
        SELECT id, code, type, creator_id, max_players, time_per_round, rounds, status
        FROM rooms
        WHERE code = ?
    `

	SQLCountRoomPlayersByRoomID = `SELECT COUNT(*) FROM room_players WHERE room_id = ?`

	SQLRoomPlayerExists = `SELECT 1 FROM room_players WHERE room_id = ? AND user_id = ?`

	SQLSelectUserPseudoByID = `SELECT pseudo FROM users WHERE id = ?`

	SQLInsertRoomPlayerMember = `
        INSERT INTO room_players (room_id, user_id, is_admin, is_ready, score)
        VALUES (?, ?, ?, 0, 0)
    `

	SQLListRoomPlayersByRoomID = `
        SELECT u.id, u.pseudo, rp.is_admin, rp.is_ready, rp.score
        FROM room_players rp
        JOIN users u ON u.id = rp.user_id
        WHERE rp.room_id = ?
        ORDER BY rp.is_admin DESC, u.pseudo ASC
    `

	SQLSelectIsAdminInRoom = `
        SELECT is_admin FROM room_players WHERE room_id = ? AND user_id = ?
    `

	// Blindtest settings
	SQLSelectBlindtestPlaylistByRoomID = `SELECT playlist FROM room_blindtest_settings WHERE room_id = ?`
	SQLUpsertBlindtestPlaylist         = `
    INSERT INTO room_blindtest_settings (room_id, playlist)
    VALUES (?, ?)
    ON CONFLICT(room_id) DO UPDATE SET playlist = excluded.playlist
`

	// Petit bac categories
	SQLCountPetitBacCategoriesByRoomID = `SELECT COUNT(*) FROM room_petitbac_categories WHERE room_id = ?`
	SQLListPetitBacCategoriesByRoomID  = `
    SELECT id, name, position
    FROM room_petitbac_categories
    WHERE room_id = ?
    ORDER BY position ASC
`
	SQLInsertPetitBacCategory = `INSERT INTO room_petitbac_categories (room_id, name, position) VALUES (?, ?, ?)`
	SQLUpdatePetitBacCategory = `UPDATE room_petitbac_categories SET name = ? WHERE id = ? AND room_id = ?`
	SQLDeletePetitBacCategory = `DELETE FROM room_petitbac_categories WHERE id = ? AND room_id = ?`
)
