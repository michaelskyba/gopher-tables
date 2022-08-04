package main

import (
	"database/sql"
	"net/http"

	"fmt"
	"strconv"
	"strings"
	"time"
)

// API for the /play/ client to send requests to with AJAX
// This is used when answering questions - you submit your answer here to check if it
// was right. If it was, your progress will be updated
func answerHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	// TODO: Return error if the URL is invalid
	// Valid: /answer/your_answer_here/

	// TODO: Return error if the user isn't signed in

	username := getCookie(request, "username")

	path := strings.Split(request.URL.Path, "/")
	answerInput, err := strconv.Atoi(path[2])

	// Bastard sent a string
	if err != nil {
		fmt.Fprintln(writer, "incorrect")
		return
	}

	// TODO: Return error if any player is at > 9 progress (game is over)
	// > 9 instead of = 10 just in case someone figures out how to get a > 10 score,
	// maybe by abusing race conditions

	// Find user ID, progress, and correct answer
	var user_id, progress, answer int
	rows, err := db.Query(`SELECT accounts.id, players.progress, questions.answer
	                     FROM accounts
		                 INNER JOIN players   ON accounts.id = players.user_id
		                 INNER JOIN games     ON games.id    = players.game_id
		                 INNER JOIN questions ON games.id    = questions.game_id
	                     WHERE players.progress = questions.progress
	                     AND accounts.username = ?`, username)
	hdl(err)

	if rows.Next() {
		err = rows.Scan(&user_id, &progress, &answer)
		hdl(err)
	}

	if answerInput == answer {

		_, err = db.Exec("UPDATE players SET progress = ? WHERE user_id = ?",
			progress+1, user_id)
		hdl(err)

		// The player won
		// It's > 8 and not > 9 because we haven't updated the progress variable
		// and instead use progress + 1 when talking to SQL
		if progress > 8 {

			_, err = db.Exec("UPDATE accounts SET wins = wins + 1 WHERE id = ?", user_id)
			hdl(err)

			// TODO: We need some way of deleting games periodically
			//
			// When a game is finished, it should be deleted ~1 minute after.
			// This gives both clients enough time to render the win screen.
			// Games should also be deleted one hour after they are created
			// if no progress is made by either player. The countdown should reset
			// if either player makes progress. This would combat AFK players or
			// players who create a game and then log off without punishing anybody
			// who has to leave for a short time.
			//
			// This could be implemented as a timer and a new field in the games
			// table called "delete_at" or something. This new field would hold the
			// epoch time at which the game should be deleted. Then, every minute,
			// the timer function will delete every game from the games table which
			// has a "delete_at" field of less than the current epoch time.
			//
			// To implement this timer, use the time library
			// https://stackoverflow.com/a/35228972
			//
			// The default value for delete_at should be (current epoch) + 3600 (1h),
			// as I said. answerHandler should add 3600 to the game's delete_at field
			// every time a player gets an answer correct. This would be done after
			// checking to make sure the game isn't over yet (i.e. one of the players
			// is at a score of > 9) so that the game wouldn't be able to run forever.
			// answerHandler would set the time to (current epoch) + 60 (1m) when a
			// player gets > 9 score (in this if statement).

			// Delete the current game in seven seconds
			// It can't be too long because players might want to immediately join
			// or create a new game after they have finished playing this one. If
			// it hasn't been deleted, they won't be able to join yet, which would
			// be annoying. If it's too short, some sort of connection lag on either
			// side could make the client accidentally skip the win message.
			// In reality, the time will be 7-10 seconds since the delete timer
			// runs every ten seconds.
			delete_at := time.Now().Unix() + 7
			_, err = db.Exec(`UPDATE games
			                 INNER JOIN players ON games.id = players.game_id
			                 SET games.delete_at = ?
			                 WHERE players.user_id = ?`, delete_at, user_id)
			hdl(err)

			return
		}

		// Return next question

		rows, err := db.Query(`SELECT text FROM questions
		                      INNER JOIN games    ON games.id    = questions.game_id
		                      INNER JOIN players  ON games.id    = players.game_id
		                      WHERE questions.progress = ?
		                      AND players.user_id = ?`, progress+1, user_id)
		hdl(err)

		var question string
		if rows.Next() {
			err = rows.Scan(&question)
			hdl(err)
		}
		fmt.Fprintln(writer, question)

	} else {
		fmt.Fprintln(writer, "incorrect")
	}
}
