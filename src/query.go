package server

// TablesSQL contient toutes les requêtes de création de tables
var TablesSQL = map[string]string{
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
}

//variable IndexesSQL contient les requêtes de création d'index LIANT les tables entre elles

var IndexesSQL = map[string]string{
	"idx_rooms_owner":       "CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(creator_id);",
	"idx_room_players_room": "CREATE INDEX IF NOT EXISTS idx_room_players_room ON room_players(room_id);",
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
)
