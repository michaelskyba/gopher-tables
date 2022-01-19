package main

import (
	"html/template"
	"net/http"

	"time"
	"math/rand"

	"fmt"
	"log"
	"regexp"
	"strings"
	"strconv"

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
	"html/password.html",
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

// Check if a player is already in a game
func is_already_in_game(username string, db *sql.DB) (bool, string) {
	rows, err := db.Query(`SELECT games.name FROM games
	                      INNER JOIN players  ON games.id    = players.game_id
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      WHERE accounts.username = ?`, username)
	handle(err)

	if rows.Next() {
		var game_name string
		err = rows.Scan(&game_name)
		handle(err)

		return true, game_name
	}

	return false, ""
}

// Add a player to a game
func add_player(game_name, username string, db *sql.DB) {

	// Find the game ID and user ID

	var game_id, user_id int

	rows, err := db.Query("SELECT id FROM games WHERE name = ?", game_name)
	handle(err)

	if rows.Next() {
		err = rows.Scan(&game_id)
		handle(err)
	}
	rows.Close()

	rows, err = db.Query("SELECT id FROM accounts WHERE username = ?", username)
	handle(err)

	if rows.Next() {
		err = rows.Scan(&user_id)
		handle(err)
	}
	rows.Close()

	_, err = db.Exec("UPDATE players SET progress = -1 WHERE user_id = ?", user_id)
	handle(err)

	_, err = db.Exec("INSERT INTO players (game_id, user_id) VALUES (?, ?)",
		game_id, user_id)
	handle(err)
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
	// An exception might be for /progress/ if it's slowing down the response
	// speed significantly, because /progress/ needs to be called rapidly.

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

	// TODO: Hash password instead of storing in plaintext

	_, err := db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", form_username, form_password)
	if err != nil {
		set_cookie(writer, "message", "Error: That username is taken.")
		redirect(writer, request, "/register/")
		return
	}

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

	// Make sure game exists and get password
	var password string
	rows, err := db.Query("SELECT password FROM games WHERE name = ?", game_name)
	handle(err)
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&password)
		handle(err)

	} else {
		set_cookie(writer, "message", "Error: That game was not found.")
		redirect(writer, request, "/")
		return
	}

	// TODO: Don't require a password if you've already joined the game

	if password != "" {

		template_input := struct {
			Name string
			Message string
		}{
			game_name,
			"",
		}

		// TODO: Hash passwords in /create_post/ and compare hashes here

		// User just clicked "join" on /lobby/
		if request.Method != http.MethodPost {
			err := templates.ExecuteTemplate(writer, "password.html", template_input)
			handle(err)

			return

		// Wrong password
		} else if request.FormValue("password") != password {

			// TODO: Use the message cookie and a redirect() on a POST
			// submission instead of rendering a template on POST. This makes
			// it so the user can reload the page and have the message cleared
			// without seeing a re-submit form prompt, but it's also just cleaner.

			template_input.Message = "Error: Incorrect password"

			err := templates.ExecuteTemplate(writer, "password.html", template_input)
			handle(err)

			return
		}

		// They have the correct password, so we just proceed as usual
	}

	yes, existing_name := is_already_in_game(username, db)
	if yes && existing_name != game_name {
		message := fmt.Sprintf("Error: You're already in a game ('%v')", existing_name)
		set_cookie(writer, "message", message)

		redirect(writer, request, "/")
		return

	} else if existing_name != game_name {
		add_player(game_name, username, db)
	}

	redirect(writer, request, fmt.Sprintf("/play/%v/", game_name))
}

func play_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	username := get_cookie(request, "username")
	if username == "" {
		set_cookie(writer, "message", "Log in to play.")
		redirect(writer, request, "/")
	}

	// TODO: Check if the player has joined or not in the players table

	// TODO:
	// Validate the game password cookie the player has if
	// this game is password-protected

	// play.html will be sending requests to /progress/ to see progress.
	// It will need to send the game_id to /progress/, so we need to give that ID
	// to the client through a template value.

	// We want the trailing slash, so 3, not 2
	path := strings.Split(request.URL.Path, "/")
	if len(path) < 3 {
		set_cookie(writer, "message", "Error: Visit the lobby to join a game")
		redirect(writer, request, "/")
	}

	game_name := path[2]

	rows, err := db.Query("SELECT id FROM games WHERE name = ?", game_name)
	handle(err)

	var game_id int
	if rows.Next() {
		err = rows.Scan(&game_id)
		handle(err)
	}

	// TODO: Error if game doesn't exist

	// Don't send an empty string if the player has won
	rows, err = db.Query(`SELECT players.progress FROM players
	                     INNER JOIN accounts ON accounts.id = players.user_id
	                     WHERE players.progress > 9
	                     AND accounts.username = ?`, username)
	handle(err)

	if rows.Next() {
		template_input := struct {
			Name string
			ID int
			Question string
		}{
			game_name,
			game_id,
			"winner",
		}

		err = templates.ExecuteTemplate(writer, "play.html", template_input)
		handle(err)
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
		handle(err)
	}

	template_input := struct {
		Name string
		ID int
		Question string
	}{
		game_name,
		game_id,
		question,
	}

	err = templates.ExecuteTemplate(writer, "play.html", template_input)
	handle(err)
}

// Basically, a user creates a game and that user's progress is set to -1. /play/
// renders the page and tells the client that the current question is "".
// play.js knows that "" means that we're waiting for a player to join, so it says that.
// Once a new player joins, we can start playing, so answer.js sends a request here, to
// /init_question/. init_question_handler needs to set the player's progress to 0
// and return the first question, which the user can now start to solve. The rest of
// the questions are provided by /answer/ on correct answer submissions.
//
// This is the best way I can think of for fetching the first question in /play/...
// We can't give the question to the client from the start or else
// the host will have an advantage using a userscript.
// You can argue that it doesn't matter because it's already possible to have
// a userscript that solves multiplication for you, but that's different, since it
// hijacks the core mechanic of the game instead of a specific implementation detail.
func init_question_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	// TODO: Make sure the player is logged in
	// TODO: Make sure the player has joined a game
	// TODO: Make sure the game has two players

	// TODO: Ensure valid URL (/init_question/<game id>/)
	// Now that I think about it, is providing the game ID useless if we're already
	// going to be confirming it using the username cookie? At that point, we might
	// as well just figure it out on the server side, right?
	// Yeah, it /is/ useless. Get rid of game_id as an argument (same with /progress/).

	username := get_cookie(request, "username")

	// Set progress to first real value "0" instead of -1
	_, err   := db.Exec(`UPDATE players
	                    INNER JOIN accounts ON accounts.id = players.user_id
	                    SET players.progress = 0
	                    WHERE accounts.username = ?`, username)
	handle(err)

	// Display first question
	rows, err := db.Query(`SELECT text FROM questions
	                      INNER JOIN games    ON games.id    = questions.game_id
	                      INNER JOIN players  ON games.id    = players.game_id
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      WHERE questions.progress = 0
	                      AND   accounts.username  = ?`, username)
	handle(err)

	if rows.Next() {
		var question string
		err = rows.Scan(&question)
		handle(err)

		fmt.Fprintln(writer, question)
	}
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
	// TODO: Return error if the user isn't signed in
	// TODO: Return error if the user hasn't joined

	// TODO: Use game associated with username cookie instead of a URL argument
	// Why? Well, we have to query the database either way to see if they
	// have joined, so it makes no sense to pass around a game ID on top of that

	progress_set := map[string]int{}

	rows, err := db.Query(`SELECT players.progress, accounts.username
	                      FROM players
	                      INNER JOIN accounts ON accounts.id = players.user_id
	                      INNER JOIN games    ON games.id    = players.game_id
	                      WHERE players.game_id = ?`, game_id)
	handle(err)

	for rows.Next() {
		var progress int
		var username string
		err = rows.Scan(&progress, &username)

		progress_set[username] = progress
	}
	rows.Close()

	encoder := json.NewEncoder(writer)
	encoder.Encode(progress_set)
}

// API for the /play/ client to send requests to with AJAX
// This is used when answering questions - you submit your answer here to check if it
// was right. If it was, your progress will be updated
func answer_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	// TODO: Return error if the URL is invalid
	// Valid: /answer/your_answer_here/

	// TODO: Return error if the user isn't signed in

	username := get_cookie(request, "username")

	path := strings.Split(request.URL.Path, "/")
	answer_input, err := strconv.Atoi(path[2])

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
	handle(err)

	if rows.Next() {
		err = rows.Scan(&user_id, &progress, &answer)
		handle(err)
	}

	if answer_input == answer {

		_, err = db.Exec("UPDATE players SET progress = ? WHERE user_id = ?",
		                  progress + 1, user_id)
		handle(err)

		// The player won
		// It's > 8 and not > 9 because we haven't updated the progress variable
		// and instead use progress + 1 when talking to SQL
		if progress > 8 {

			_, err = db.Exec("UPDATE accounts SET wins = wins + 1 WHERE id = ?", user_id)
			handle(err)

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
			// as I said. answer_handler should add 3600 to the game's delete_at field
			// every time a player gets an answer correct. This would be done after
			// checking to make sure the game isn't over yet (i.e. one of the players
			// is at a score of > 9) so that the game wouldn't be able to run forever.
			// answer_handler would set the time to (current epoch) + 60 (1m) when a
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
			handle(err)

			return
		}

		// Return next question

		rows, err := db.Query(`SELECT text FROM questions
		                      INNER JOIN games    ON games.id    = questions.game_id
		                      INNER JOIN players  ON games.id    = players.game_id
		                      WHERE questions.progress = ?
		                      AND players.user_id = ?`, progress + 1, user_id)
		handle(err)

		var question string
		if rows.Next() {
			err = rows.Scan(&question)
			handle(err)
		}
		fmt.Fprintln(writer, question)

	} else {
		fmt.Fprintln(writer, "incorrect")
	}
}

// Create game form submission URL endpoint
func create_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method != http.MethodPost {
		return
	}

	username := get_cookie(request, "username")
	if username == "" {
		set_cookie(writer, "message", "Error: You must be logged in to create a game")
		redirect(writer, request, "/")
	}

	name := request.FormValue("name")
	password := request.FormValue("password")

	// Ensure standard-looking game names
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(name) {
		set_cookie(writer, "message", "Error: Your game's name must match '^[a-zA-Z0-9 _-]+$'.")
		redirect(writer, request, "/create/")
		return
	}

	if len(name) > 15 {
		set_cookie(writer, "message", "Error: Don't try to circumvent client-side validation, you goblin")
		redirect(writer, request, "/create/")
	}

	yes, existing_name := is_already_in_game(username, db)
	if yes {
		message := fmt.Sprintf("Error: You're already in a game ('%v')", existing_name)
		set_cookie(writer, "message", message)

		redirect(writer, request, "/create/")
		return
	}

	delete_at := int(time.Now().Unix()) + 3600
	result, err := db.Exec("INSERT INTO games (name, password, delete_at) VALUES (?, ?, ?)",
	                       name, password, delete_at)

	if err != nil {
		set_cookie(writer, "message", "Error: Game already exists")
		redirect(writer, request, "/create/")
		return
	}

	game_id, err := result.LastInsertId()
	handle(err)

	for i := 0; i < 10; i++ {
		a := rand.Intn(12) + 1 // 1 to 12
		b := rand.Intn(12) + 1 // 1 to 12

		_, err := db.Exec(`INSERT INTO questions (game_id, text, answer, progress)
		                  VALUES (?, ?, ?, ?)`, game_id, fmt.Sprintf("%v × %v", a, b),
		                  a * b, i)
		handle(err)
	}

	add_player(name, username, db)

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

// Delete games scheduled for deletion (either finished or AFK timeout)
func game_delete_timer(db *sql.DB) {
	for range time.Tick(time.Second * 10) {

		current := time.Now().Unix()

		_, err := db.Exec("DELETE FROM games WHERE delete_at < ?", current)
		handle(err)
	}
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

	// TODO: Use SQL Joins instead of manually fiddling with IDs

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
	http.HandleFunc("/init_question/", func(writer http.ResponseWriter, request *http.Request) {
		init_question_handler(writer, request, db)
	})
	http.HandleFunc("/progress/", func(writer http.ResponseWriter, request *http.Request) {
		progress_handler(writer, request, db)
	})
	http.HandleFunc("/answer/", func(writer http.ResponseWriter, request *http.Request) {
		answer_handler(writer, request, db)
	})
	http.HandleFunc("/create/", create_get_handler)
	http.HandleFunc("/create_post/", func(writer http.ResponseWriter, request *http.Request) {
		create_post_handler(writer, request, db)
	})
	http.HandleFunc("/logout/", logout_handler)

	go game_delete_timer(db)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
