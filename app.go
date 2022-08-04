package main

import (
	"html/template"
	"net/http"

	"math/rand"
	"time"

	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
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

func hdl(err error) {
	if err != nil {
		panic(err)
	}
}

func setCookie(writer http.ResponseWriter, name, value string) {
	http.SetCookie(writer, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

func getCookie(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)

	if err != nil {
		return ""
	}
	return cookie.Value
}

func redirect(writer http.ResponseWriter, request *http.Request, path string) {
	http.Redirect(writer, request, path, http.StatusSeeOther)
}

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

// Home page
func homeHandler(writer http.ResponseWriter, request *http.Request) {
	username := getCookie(request, "username")
	message := getCookie(request, "message")

	if request.URL.Path == "/" {
		setCookie(writer, "message", "")

		templateInput := struct {
			Message  string
			LoggedIn bool
		}{
			message,
			username != "",
		}

		err := templates.ExecuteTemplate(writer, "index.html", templateInput)
		hdl(err)

	} else {
		err := templates.ExecuteTemplate(writer, "404.html", request.URL.Path)
		hdl(err)
	}
}

// Log in page
func loginGetHandler(writer http.ResponseWriter, request *http.Request) {

	username := getCookie(request, "username")
	message := getCookie(request, "message")

	if username != "" {
		setCookie(writer, "message", "You're already logged in.")
		redirect(writer, request, "/")
		return
	}

	setCookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "login.html", message)
	hdl(err)
}

// Log in URL point for submitting the log in form
func loginPostHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method != http.MethodPost {
		return
	}

	formUsername := request.FormValue("username")
	formPassword := request.FormValue("password")

	if formUsername == "" {
		setCookie(writer, "message", "Error: You have entered invalid credentials.")
		redirect(writer, request, "/login/")
		return
	}

	rows, err := db.Query("SELECT password FROM accounts WHERE username = ?", formUsername)
	hdl(err)
	defer rows.Close()

	success := true
	if rows.Next() {
		var password string
		err = rows.Scan(&password)
		hdl(err)

		if password != formPassword {
			success = false
		}

	} else {
		success = false
	}

	if !success {
		setCookie(writer, "message", "Error: You have entered invalid credentials.")
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

	setCookie(writer, "username", formUsername)
	setCookie(writer, "message", "You have successfully logged in.")
	redirect(writer, request, "/")
}

// Register page
func registerGetHandler(writer http.ResponseWriter, request *http.Request) {

	username := getCookie(request, "username")
	message := getCookie(request, "message")

	if username != "" {
		setCookie(writer, "message", "You're already logged in.")
		redirect(writer, request, "/")
		return
	}

	setCookie(writer, "message", "")

	err := templates.ExecuteTemplate(writer, "register.html", message)
	hdl(err)
}

// Register URL pointing for submitting POST request form
func registerPostHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	formUsername := request.FormValue("username")
	formPassword := request.FormValue("password")
	formConfirm := request.FormValue("confirm")

	if formPassword != formConfirm {
		setCookie(writer, "message", "Error: Your passwords don't match.")
		redirect(writer, request, "/register/")
		return
	}

	// Have standard-looking usernames
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(formUsername) {
		setCookie(writer, "message", "Error: Your username must match '^[a-zA-Z0-9 _-]+$'.")
		redirect(writer, request, "/register/")
		return
	}

	// TODO: Hash password instead of storing in plaintext

	_, err := db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", formUsername, formPassword)
	if err != nil {
		setCookie(writer, "message", "Error: That username is taken.")
		redirect(writer, request, "/register/")
		return
	}

	// Log in
	setCookie(writer, "username", formUsername)
	setCookie(writer, "message", "You have successfully registered.")
	redirect(writer, request, "/")
}

// Profile
func profileHandler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	username := getCookie(request, "username")

	if username == "" {
		setCookie(writer, "message", "Log in to see your profile.")
		redirect(writer, request, "/")
		return
	}

	// Find user's win count
	// Breaks if the user injects a non-existent username as browser cookie

	rows, err := db.Query("SELECT wins FROM accounts WHERE username = ?", username)

	var wins int
	if rows.Next() {
		err = rows.Scan(&wins)
		hdl(err)
	}

	type profile struct {
		Username string
		Wins     int
	}

	err = templates.ExecuteTemplate(writer, "profile.html", profile{username, wins})
	hdl(err)
}

// Lobby
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
	if existingName == gameName {

		// Player has already joined - don't ask them for the password again
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

// Log out
func logoutHandler(writer http.ResponseWriter, request *http.Request) {

	if getCookie(request, "username") == "" {
		setCookie(writer, "message", "You're already logged out.")
	} else {
		setCookie(writer, "username", "")
		setCookie(writer, "message", "You have successfully logged out.")
	}

	redirect(writer, request, "/")
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
	hdl(err)

	// Check for connection
	err = db.Ping()
	hdl(err)

	server := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", server))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login/", loginGetHandler)
	http.HandleFunc("/login_post/", func(writer http.ResponseWriter, request *http.Request) {
		loginPostHandler(writer, request, db)
	})
	http.HandleFunc("/register/", registerGetHandler)
	http.HandleFunc("/register_post/", func(writer http.ResponseWriter, request *http.Request) {
		registerPostHandler(writer, request, db)
	})
	http.HandleFunc("/profile/", func(writer http.ResponseWriter, request *http.Request) {
		profileHandler(writer, request, db)
	})
	http.HandleFunc("/lobby/", func(writer http.ResponseWriter, request *http.Request) {
		lobbyHandler(writer, request, db)
	})
	http.HandleFunc("/join/", func(writer http.ResponseWriter, request *http.Request) {
		joinHandler(writer, request, db)
	})
	http.HandleFunc("/play/", func(writer http.ResponseWriter, request *http.Request) {
		playHandler(writer, request, db)
	})
	http.HandleFunc("/init_question/", func(writer http.ResponseWriter, request *http.Request) {
		initQuestionHandler(writer, request, db)
	})
	http.HandleFunc("/progress/", func(writer http.ResponseWriter, request *http.Request) {
		progressHandler(writer, request, db)
	})
	http.HandleFunc("/answer/", func(writer http.ResponseWriter, request *http.Request) {
		answerHandler(writer, request, db)
	})
	http.HandleFunc("/create/", createGetHandler)
	http.HandleFunc("/create_post/", func(writer http.ResponseWriter, request *http.Request) {
		createPostHandler(writer, request, db)
	})
	http.HandleFunc("/logout/", logoutHandler)

	go gameDeleteTimer(db)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
