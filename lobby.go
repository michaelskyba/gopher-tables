package main

import (
	"database/sql"
	"net/http"
)

func lobbyHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	// TODO:
	// Display more information in /lobby/ (difficult)
	// - the number of players in each game
	// - if it has a password or not
	// - if you've joined this game

	if getCookie(request, "username") == "" {
		setCookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
		return
	}

	var current struct {
		Message string
		Games   []string
	}

	current.Message = getCookie(request, "message")
	setCookie(writer, "message", "")

	// Get list of games from database
	rows, err := db.Query("SELECT name FROM games")
	hdl(err)
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		hdl(err)

		current.Games = append(current.Games, name)
	}

	err = templates.ExecuteTemplate(writer, "lobby.html", current)
	hdl(err)
}
