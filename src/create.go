package main

import (
	"database/sql"
	"net/http"

	"math/rand"
	"time"

	"fmt"
	"regexp"
)

// Create game page
func createGetHandler(writer http.ResponseWriter, request *http.Request) {
	// TODO: Check if the player is already in a game here, before create_post

	if getCookie(request, "username") == "" {
		setCookie(writer, "message", "Log in to create a game.")
		redirect(writer, request, "/")
	}

	message := getCookie(request, "message")
	setCookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "create.html", message)
	hdl(err)
}

// Create game form submission URL endpoint
func createPostHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	if request.Method != http.MethodPost {
		return
	}

	username := getCookie(request, "username")
	if username == "" {
		setCookie(writer, "message", "Error: You must be logged in to create a game.")
		redirect(writer, request, "/")
	}

	name := request.FormValue("name")
	password := request.FormValue("password")

	// Ensure standard-looking game names
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(name) {
		setCookie(writer, "message", "Error: Your game's name must match '^[a-zA-Z0-9 _-]+$'.")
		redirect(writer, request, "/create/")
		return
	}

	if len(name) > 15 {
		setCookie(writer, "message", "Error: Don't try to circumvent client-side validation, you goblin.")
		redirect(writer, request, "/create/")
	}

	existingName := inGame(username, db)
	if existingName != "" {
		message := fmt.Sprintf("Error: You're already in a game ('%v').", existingName)
		setCookie(writer, "message", message)

		redirect(writer, request, "/create/")
		return
	}

	delete_at := int(time.Now().Unix()) + 3600
	result, err := db.Exec("INSERT INTO games (name, password, delete_at) VALUES (?, ?, ?)",
		name, password, delete_at)

	if err != nil {
		setCookie(writer, "message", "Error: Game already exists.")
		redirect(writer, request, "/create/")
		return
	}

	game_id, err := result.LastInsertId()
	hdl(err)

	// TODO: Set some kind of seed for rand, because it's using the same
	// numbers every time

	for i := 0; i < 10; i++ {
		a := rand.Intn(12) + 1 // 1 to 12
		b := rand.Intn(12) + 1 // 1 to 12

		_, err := db.Exec(`INSERT INTO questions (game_id, text, answer, progress)
		                  VALUES (?, ?, ?, ?)`, game_id, fmt.Sprintf("%v × %v", a, b),
			a*b, i)
		hdl(err)
	}

	addPlayer(name, username, db)

	redirect(writer, request, fmt.Sprintf("/play/%v/", name))
}
