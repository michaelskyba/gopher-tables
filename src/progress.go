package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// API for the /play/ client to send requests to with AJAX
// This will return the progress of the players so that the client can render them
// Valid: /progress/game_id/
func progressHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	path := strings.Split(request.URL.Path, "/")
	game_id := path[2]

	// TODO: Return error if the URL is invalid (e.g. localhost:8000/progress/)
	// TODO: Return error if the user isn't signed in
	// TODO: Return error if the user hasn't joined

	// TODO: Use game associated with username cookie instead of a URL argument
	// Why? Well, we have to query the database either way to see if they
	// have joined, so it makes no sense to pass around a game ID on top of that

	progressSet := map[string]int{}

	rows, err := db.Query(`SELECT players.progress, accounts.username
	                      FROM players
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      INNER JOIN games    ON games.id    = players.game_id
	                      WHERE players.game_id = ?`, game_id)
	hdl(err)

	for rows.Next() {
		var progress int
		var username string
		err = rows.Scan(&progress, &username)

		progressSet[username] = progress
	}
	rows.Close()

	encoder := json.NewEncoder(writer)
	encoder.Encode(progressSet)
}
