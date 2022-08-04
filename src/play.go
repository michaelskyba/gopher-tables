package main

import (
	"database/sql"
	"net/http"
	"strings"
)

func playHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	// TODO: Make an "abort game" button in play.html so that a host can abort
	// if nobody is joining and thus wants to join a different game

	username := getCookie(request, "username")
	if username == "" {
		setCookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
	}

	// play.html will be sending requests to /progress/ to see progress.
	// It will need to send the game_id to /progress/, so we need to give that ID
	// to the client through a template value.

	// We want the trailing slash, so 3, not 2
	path := strings.Split(request.URL.Path, "/")
	if len(path) < 3 {
		setCookie(writer, "message", "Error: Visit the lobby to join a game.")
		redirect(writer, request, "/")
	}

	gameName := path[2]

	rows, err := db.Query("SELECT id FROM games WHERE name = ?", gameName)
	hdl(err)

	var game_id int
	if rows.Next() {
		err = rows.Scan(&game_id)
		hdl(err)

	} else {
		setCookie(writer, "message", "Error: Game doesn't exist.")
		redirect(writer, request, "/lobby/")
		return
	}

	// User hasn't joined but is going to /play/x/ manually
	if inGame(username, db) != gameName {
		setCookie(writer, "message", "Error: You are not in this game.")
		redirect(writer, request, "/lobby/")
		return
	}

	// Don't send an empty string if the player has won
	rows, err = db.Query(`SELECT players.progress FROM players
	                     INNER JOIN accounts ON accounts.id = players.user_id
	                     WHERE players.progress > 9
	                     AND accounts.username = ?`, username)
	hdl(err)

	if rows.Next() {
		templateInput := struct {
			Name     string
			ID       int
			Question string
		}{
			gameName,
			game_id,
			"winner",
		}

		err = templates.ExecuteTemplate(writer, "play.html", templateInput)
		hdl(err)
		return
	}

	// If there's no question (i.e. progress == -1), this question variable
	// will be "", which is accounted for by the client
	var question string
	rows, err = db.Query(`SELECT text FROM questions
	                      INNER JOIN games    ON games.id    = questions.game_id
	                      INNER JOIN players  ON games.id    = players.game_id
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      WHERE questions.progress = players.progress
	                      AND accounts.username = ?`, username)
	if rows.Next() {
		err = rows.Scan(&question)
		hdl(err)
	}

	templateInput := struct {
		Name     string
		ID       int
		Question string
	}{
		gameName,
		game_id,
		question,
	}

	err = templates.ExecuteTemplate(writer, "play.html", templateInput)
	hdl(err)
}
