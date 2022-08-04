package main

import (
	"database/sql"
	"time"
)

// Get the name of the current game the player is in, or "" if none
func inGame(username string, db *sql.DB) string {
	rows, err := db.Query(`SELECT games.name FROM games
	                      INNER JOIN players  ON games.id    = players.game_id
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      WHERE accounts.username = ?`, username)
	hdl(err)

	if rows.Next() {
		var gameName string
		err = rows.Scan(&gameName)
		hdl(err)

		return gameName
	}

	return ""
}

// Delete games scheduled for deletion (either finished or AFK timeout)
func gameDeleteTimer(db *sql.DB) {
	for range time.Tick(time.Second * 10) {
		current := time.Now().Unix()

		_, err := db.Exec(`DELETE games, players
		                  FROM games
		                  INNER JOIN players ON games.id = players.game_id
		                  WHERE games.delete_at < ?`, current)
		hdl(err)
	}
}

// Add a player to a game
func addPlayer(gameName, username string, db *sql.DB) {
	// Find the game ID and user ID
	var game_id, user_id int

	rows, err := db.Query("SELECT id FROM games WHERE name = ?", gameName)
	hdl(err)

	if rows.Next() {
		err = rows.Scan(&game_id)
		hdl(err)
	}
	rows.Close()

	rows, err = db.Query("SELECT id FROM accounts WHERE username = ?", username)
	hdl(err)

	if rows.Next() {
		err = rows.Scan(&user_id)
		hdl(err)
	}
	rows.Close()

	_, err = db.Exec("UPDATE players SET progress = -1 WHERE user_id = ?", user_id)
	hdl(err)

	_, err = db.Exec("INSERT INTO players (game_id, user_id) VALUES (?, ?)",
		game_id, user_id)
	hdl(err)
}
