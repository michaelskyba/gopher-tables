package main

import (
	"database/sql"
	"net/http"

	"fmt"
)

// Basically, a user creates a game and that user's progress is set to -1. /play/
// renders the page and tells the client that the current question is "".
// play.js knows that "" means that we're waiting for a player to join, so it says that.
// Once a new player joins, we can start playing, so answer.js sends a request here, to
// /init_question/. initQuestionHandler needs to set the player's progress to 0
// and return the first question, which the user can now start to solve. The rest of
// the questions are provided by /answer/ on correct answer submissions.
//
// This is the best way I can think of for fetching the first question in /play/...
// We can't give the question to the client from the start or else
// the host will have an advantage using a userscript.
// You can argue that it doesn't matter because it's already possible to have
// a userscript that solves multiplication for you, but that's different, since it
// hijacks the core mechanic of the game instead of a specific implementation detail.
func initQuestionHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	// TODO: Make sure the player is logged in
	// TODO: Make sure the player has joined a game
	// TODO: Make sure the game has two players

	// TODO: Ensure valid URL (/init_question/<game id>/)
	// Now that I think about it, is providing the game ID useless if we're already
	// going to be confirming it using the username cookie? At that point, we might
	// as well just figure it out on the server side, right?
	// Yeah, it /is/ useless. Get rid of game_id as an argument (same with /progress/).

	username := getCookie(request, "username")

	// Set progress to first real value "0" instead of -1
	_, err := db.Exec(`UPDATE players
	                    INNER JOIN accounts ON accounts.id = players.user_id
	                    SET players.progress = 0
	                    WHERE accounts.username = ?`, username)
	hdl(err)

	// Display first question
	rows, err := db.Query(`SELECT text FROM questions
	                      INNER JOIN games    ON games.id    = questions.game_id
	                      INNER JOIN players  ON games.id    = players.game_id
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      WHERE questions.progress = 0
	                      AND   accounts.username  = ?`, username)
	hdl(err)

	if rows.Next() {
		var question string
		err = rows.Scan(&question)
		hdl(err)

		fmt.Fprintln(writer, question)
	}
}
