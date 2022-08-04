package main

import (
	"database/sql"
	"net/http"
	"strings"

	"fmt"
)

// /join/<name>/, accessed when pressing "Join" on a game
func joinHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	username := getCookie(request, "username")
	if username == "" {
		setCookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
		return
	}

	path := strings.Split(request.URL.Path, "/")
	if len(path) != 4 {
		setCookie(writer, "message", "Visit the lobby (press 'Play') to join a game.")
		redirect(writer, request, "/")
		return
	}
	var gameName = path[2]

	existingName := inGame(username, db)

	// Player has already joined - don't ask them for the password again
	if existingName == gameName {
		redirect(writer, request, fmt.Sprintf("/play/%v/", gameName))
		return

	} else if existingName != "" {
		message := fmt.Sprintf("Error: You're already in a game ('%v').", existingName)
		setCookie(writer, "message", message)

		redirect(writer, request, "/lobby/")
		return
	}

	// Make sure game exists and get password
	var password string
	rows, err := db.Query("SELECT password FROM games WHERE name = ?", gameName)
	hdl(err)
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&password)
		hdl(err)

	} else {
		setCookie(writer, "message", "Error: That game was not found.")
		redirect(writer, request, "/lobby/")
		return
	}

	if password != "" {
		templateInput := struct {
			Name    string
			Message string
		}{
			gameName,
			getCookie(request, "message"),
		}
		setCookie(writer, "message", "")

		// TODO: Hash passwords in /create_post/ and compare hashes here

		// User just clicked "join" on /lobby/
		if request.Method != http.MethodPost {
			err := templates.ExecuteTemplate(writer, "password.html", templateInput)
			hdl(err)
			return

		} else if request.FormValue("password") != password {
			setCookie(writer, "message", "Error: Incorrect password.")
			redirect(writer, request, request.URL.Path)
			return
		}

		// They have the correct password, so we just proceed as usual
	}

	addPlayer(gameName, username, db)
	redirect(writer, request, fmt.Sprintf("/play/%v/", gameName))
}
