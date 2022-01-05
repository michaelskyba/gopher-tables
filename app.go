package main

import (
	"html/template"
	"net/http"

	"fmt"
	"log"
	"regexp"
	"strings"

	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"os"
)

var templates = template.Must(template.ParseFiles(
	"html/index.html",
	"html/404.html",
	"html/login.html",
	"html/register.html",
	"html/profile.html",
	"html/lobby.html",
	"html/play.html",
	"html/create.html"))

func handle(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func set_cookie(writer http.ResponseWriter, name, value string) {
	http.SetCookie(writer, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

func get_cookie(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)

	if err != nil {
		return ""
	}
	return cookie.Value
}

func redirect(writer http.ResponseWriter, request *http.Request, path string) {
	http.Redirect(writer, request, path, http.StatusSeeOther)
}

// Add a player to a room
func add_player(game_name, username string, db *sql.DB) {

	// TODO: Don't let a player join if they're already in a room

	var game_id, user_id int

	// Find the game ID
	rows, err := db.Query("SELECT id FROM games WHERE name = ?", game_name)
	handle(err)

	if rows.Next() {
		err = rows.Scan(&game_id)
		handle(err)
	}
	rows.Close()

	// Find user ID
	rows, err = db.Query("SELECT id FROM accounts WHERE username = ?", username)
	handle(err)

	if rows.Next() {
		err = rows.Scan(&user_id)
		handle(err)
	}
	rows.Close()

	_, err = db.Exec("INSERT INTO players (game_id, user_id) VALUES (?, ?)",
		game_id, user_id)
}

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {
	username := get_cookie(request, "username")
	message := get_cookie(request, "message")

	if request.URL.Path == "/" {
		set_cookie(writer, "message", "")

		template_input := struct {
			Message  string
			LoggedIn bool
		}{
			message,
			username != "",
		}

		err := templates.ExecuteTemplate(writer, "index.html", template_input)
		handle(err)

	} else {
		err := templates.ExecuteTemplate(writer, "404.html", request.URL.Path)
		handle(err)
	}
}

// Log in page
func login_get_handler(writer http.ResponseWriter, request *http.Request) {

	username := get_cookie(request, "username")
	message := get_cookie(request, "message")

	if username != "" {
		set_cookie(writer, "message", "You're already logged in.")
		redirect(writer, request, "/")
		return
	}

	set_cookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "login.html", message)
	handle(err)
}

// Log in URL point for submitting the log in form
func login_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method != http.MethodPost {
		return
	}

	form_username := request.FormValue("username")
	form_password := request.FormValue("password")

	if form_username == "" {
		set_cookie(writer, "message", "Error: You have entered invalid credentials.")
		redirect(writer, request, "/login/")
		return
	}

	rows, err := db.Query("SELECT password FROM accounts WHERE username = ?", form_username)
	handle(err)
	defer rows.Close()

	success := true
	if rows.Next() {
		var password string
		err = rows.Scan(&password)
		handle(err)

		if password != form_password {
			success = false
		}

	} else {
		success = false
	}

	if !success {
		set_cookie(writer, "message", "Error: You have entered invalid credentials.")
		redirect(writer, request, "/login/")
		return
	}

	// TODO:
	// Hash their login password and store it as a cookie.
	// Then, when checking their username, check this pair.
	// This would prevent people from impersonating someone by adding their
	// username as their 'username' cookie manually.

	set_cookie(writer, "username", form_username)
	set_cookie(writer, "message", "You have successfully logged in.")
	redirect(writer, request, "/")
}

// Register page
func register_get_handler(writer http.ResponseWriter, request *http.Request) {

	username := get_cookie(request, "username")
	message := get_cookie(request, "message")

	if username != "" {
		set_cookie(writer, "message", "You're already logged in.")
		redirect(writer, request, "/")
		return
	}

	set_cookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "register.html", message)
	handle(err)
}

// Register URL pointing for submitting POST request form
func register_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	form_username := request.FormValue("username")
	form_password := request.FormValue("password")
	form_confirm := request.FormValue("confirm")

	if form_password != form_confirm {
		set_cookie(writer, "message", "Error: Your passwords don't match.")
		redirect(writer, request, "/register/")
		return
	}

	// Have standard-looking usernames
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(form_username) {
		set_cookie(writer, "message", "Error: Your username must match '^[a-zA-Z0-9 _-]+$'.")
		redirect(writer, request, "/register/")
		return
	}

	rows, err := db.Query("SELECT username FROM accounts WHERE username = ?", form_username)
	handle(err)
	defer rows.Close()
	if rows.Next() {
		set_cookie(writer, "message", "Error: That username is taken.")
		redirect(writer, request, "/register/")
		return
	}

	// TODO: Hash password instead of storing in plaintext

	_, err = db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", form_username, form_password)
	handle(err)

	// Log in
	set_cookie(writer, "username", form_username)
	set_cookie(writer, "message", "You have successfully registered.")
	redirect(writer, request, "/")
}

// Profile
func profile_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	username := get_cookie(request, "username")

	if username == "" {
		set_cookie(writer, "message", "Log in to see your profile.")
		redirect(writer, request, "/")
		return
	}

	// Find user's win count
	// Breaks if the user injects a non-existent username as browser cookie

	rows, err := db.Query("SELECT wins FROM accounts WHERE username = ?", username)

	var wins int
	if rows.Next() {
		err = rows.Scan(&wins)
		handle(err)
	}

	type profile struct {
		Username string
		Wins     int
	}

	err = templates.ExecuteTemplate(writer, "profile.html", profile{username, wins})
	handle(err)
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if get_cookie(request, "username") == "" {
		set_cookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
		return
	}

	var current struct {
		Games []string
	}

	// Get list of games from database
	rows, err := db.Query("SELECT name FROM games")
	handle(err)
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		handle(err)

		current.Games = append(current.Games, name)
	}

	err = templates.ExecuteTemplate(writer, "lobby.html", current)
	handle(err)
}

// /join/<name>/, accessed when pressing "Join" on a game
func join_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	username := get_cookie(request, "username")
	if username == "" {
		set_cookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
		return
	}

	path := strings.Split(request.URL.Path, "/")
	if len(path) != 4 {
		set_cookie(writer, "message", "Visit the lobby (press 'Play') to join a game.")
		redirect(writer, request, "/")
		return
	}
	var game_name = path[2]

	// Get list of games from database
	rows, err := db.Query("SELECT password FROM games WHERE name = ?", game_name)
	handle(err)
	defer rows.Close()
	if rows.Next() {
		var password string
		err = rows.Scan(&password)
		handle(err)

	} else {
		set_cookie(writer, "message", "Error: That game was not found.")
		redirect(writer, request, "/")
		return
	}

	add_player(game_name, username, db)

	// TODO:
	// Serve an HTML page with a password form if the game has a password
	// The password they input should be validated in a handler, maybe room_password_handler
	// If the password is correct, it should be stored as a cookie and then the user
	// should be redirected to /play/

	redirect(writer, request, fmt.Sprintf("/play/%v/", game_name))
}

func play_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if get_cookie(request, "username") == "" {
		set_cookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
	}

	// TODO: Check if the player has joined or not in the players table

	// TODO:
	// Validate the game password cookie the player has if
	// this room is password-protected

	// play.html will be sending requests to /progress/ to see progress.
	// It will need to send the game_id to /progress/, so we need to give that ID
	// to the client through a template value.

	path := strings.Split(request.URL.Path, "/")

	// TODO: Check for a valid URL path before querying database
	// Valid: /play/game_name/

	rows, err := db.Query("SELECT id FROM games WHERE name = ?", path[2])
	handle(err)

	var game_id int
	if rows.Next() {
		err = rows.Scan(&game_id)
		handle(err)
	}

	// TODO: Error if room doesn't exist

	err = templates.ExecuteTemplate(writer, "play.html", game_id)
	handle(err)
}

// Create game page
func create_get_handler(writer http.ResponseWriter, request *http.Request) {

	if get_cookie(request, "username") == "" {
		set_cookie(writer, "message", "Log in to create a game.")
		redirect(writer, request, "/")
	}

	message := get_cookie(request, "message")
	set_cookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "create.html", message)
	handle(err)
}

// API for the /play/ client to send requests to with AJAX
// This will return the progress of the players so that the client can render them
// Valid: /progress/game_id/
func progress_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	path := strings.Split(request.URL.Path, "/")
	game_id := path[2]

	// TODO: Return error if the URL is invalid (e.g. localhost:8000/progress/)
	// TODO: Return error if the user hasn't joined
	// TODO: Use the URL argument for the game_id instead of hardcoding "0"

	// First map user IDs to progress
	user_ids := map[int]int{}

	rows, err := db.Query("SELECT user_id, progress FROM players WHERE game_id = ?", game_id)
	handle(err)

	for rows.Next() {
		var user_id, progress int
		err = rows.Scan(&user_id, &progress)

		user_ids[user_id] = progress
	}
	rows.Close()

	// Get the usernames for those user IDs
	usernames := map[string]int{}

	for user_id, progress := range user_ids {

		rows, err = db.Query("SELECT username FROM accounts WHERE id = ?", user_id)
		handle(err)

		if rows.Next() {
			var username string
			err = rows.Scan(&username)
			handle(err)

			usernames[username] = progress
		}

		rows.Close()
	}

	encoder := json.NewEncoder(writer)
	encoder.Encode(usernames)
}

// Create game form submission URL endpoint
func create_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method != http.MethodPost {
		return
	}

	username := get_cookie(request, "username")
	if username == "" {
		set_cookie(writer, "message", "Error: You must be logged in to create a room")
		redirect(writer, request, "/")
	}

	name := request.FormValue("name")
	password := request.FormValue("password")

	// Ensure standard-looking room names
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(name) {
		set_cookie(writer, "message", "Error: Your room's name must match '^[a-zA-Z0-9 _-]+$'.")
		redirect(writer, request, "/create/")
		return
	}

	if len(name) > 15 {
		set_cookie(writer, "message", "Error: Don't try to circumvent client-side validation, you goblin")
		redirect(writer, request, "/create/")
	}

	// TODO: Check if room name already exists

	_, err := db.Exec("INSERT INTO games (name, password) VALUES (?, ?)", name, password)
	handle(err)

	// TODO: Declare this player as having joined the game in the players database

	redirect(writer, request, fmt.Sprintf("/play/%v/", name))
}

// Log out
func logout_handler(writer http.ResponseWriter, request *http.Request) {

	if get_cookie(request, "username") == "" {
		set_cookie(writer, "message", "You're already logged out.")
	} else {
		set_cookie(writer, "username", "")
		set_cookie(writer, "message", "You have successfully logged out.")
	}

	redirect(writer, request, "/")
}

func main() {
	config := mysql.Config{
		User:                 "michael",
		Passwd:               "password",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "db",
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	handle(err)

	// Check for connection
	err = db.Ping()
	handle(err)

	server := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", server))

	http.HandleFunc("/", home_handler)
	http.HandleFunc("/login/", login_get_handler)
	http.HandleFunc("/login_post/", func(writer http.ResponseWriter, request *http.Request) {
		login_post_handler(writer, request, db)
	})
	http.HandleFunc("/register/", register_get_handler)
	http.HandleFunc("/register_post/", func(writer http.ResponseWriter, request *http.Request) {
		register_post_handler(writer, request, db)
	})
	http.HandleFunc("/profile/", func(writer http.ResponseWriter, request *http.Request) {
		profile_handler(writer, request, db)
	})
	http.HandleFunc("/lobby/", func(writer http.ResponseWriter, request *http.Request) {
		lobby_handler(writer, request, db)
	})
	http.HandleFunc("/join/", func(writer http.ResponseWriter, request *http.Request) {
		join_handler(writer, request, db)
	})
	http.HandleFunc("/play/", func(writer http.ResponseWriter, request *http.Request) {
		play_handler(writer, request, db)
	})
	http.HandleFunc("/progress/", func(writer http.ResponseWriter, request *http.Request) {
		progress_handler(writer, request, db)
	})
	http.HandleFunc("/create/", create_get_handler)
	http.HandleFunc("/create_post/", func(writer http.ResponseWriter, request *http.Request) {
		create_post_handler(writer, request, db)
	})
	http.HandleFunc("/logout/", logout_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
