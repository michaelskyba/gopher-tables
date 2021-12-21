package main

import (
	"html/template"
	"net/http"

	"fmt"
	"log"
	"strings"
	"regexp"

	"os"
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

var templates = template.Must(template.ParseFiles(
	"html/index.html",
	"html/404.html",
	"html/login.html",
	"html/register.html",
	"html/profile.html",
	"html/lobby.html"))

type template_values struct {
	Message  string
	LoggedIn bool
}

type lobby struct {
	Games []string
}

type account struct {
	ID       int
	username string
	password string
	wins     int
}

type profile struct {
	Username string
	Wins     int
}

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

// Returns the username and message of the current session
func get_template_values(request *http.Request) (string, string) {

	var username string
	var message string

	for _, cookie := range request.Cookies() {
		if cookie.Name == "username" {
			username = cookie.Value
		}
		if cookie.Name == "message" {
			message = cookie.Value
		}
	}

	return username, message
}

func redirect(writer http.ResponseWriter, request *http.Request, path string) {
	http.Redirect(writer, request, path, http.StatusSeeOther)
}

func render_template(filename string, writer http.ResponseWriter, values template_values) {
	err := templates.ExecuteTemplate(writer, filename, values)
	handle(err)
}

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {

	signed_in := true
	username, message := get_template_values(request)
	if username == "" {
		signed_in = false
	}

	// Normal
	if request.URL.Path == "/" {
		set_cookie(writer, "message", "")
		render_template("index.html", writer, template_values{message, signed_in})

	// 404
	} else {
		values := template_values{request.URL.Path, false}
		render_template("404.html", writer, values)
	}
}

// Log in page
func login_get_handler(writer http.ResponseWriter, request *http.Request) {

	username, message := get_template_values(request)

	// Already logged in
	if username != "" {
		redirect(writer, request, "/")
	}

	set_cookie(writer, "message", "")
	render_template("login.html", writer, template_values{message, false})
}

// Log in URL point for submitting the log in form
func login_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method != http.MethodPost {
		return
	}

	form_username := request.FormValue("username")
	form_password := request.FormValue("password")

	if form_username == "" {
		set_cookie(writer, "message", "Error: Invalid credentials")
		redirect(writer, request, "/login/")
	}

	rows, err := db.Query("SELECT * FROM accounts WHERE username = ?", form_username)
	handle(err)
	defer rows.Close()

	success := false
	for rows.Next() {
		var current account
		err = rows.Scan(&current.ID, &current.username, &current.password, &current.wins)
		handle(err)

		if current.password == form_password {
			success = true
		}
	}

	if !success {
		set_cookie(writer, "message", "Error: Invalid credentials")
		redirect(writer, request, "/login/")
		return
	}

	set_cookie(writer, "username", form_username)
	set_cookie(writer, "message", "Successfully logged in")
	redirect(writer, request, "/")
}

// Register page
func register_get_handler(writer http.ResponseWriter, request *http.Request) {

	username, message := get_template_values(request)

	// Redirect to homepage if already signed in
	if username != "" {
		redirect(writer, request, "/")
	}

	set_cookie(writer, "message", "")
	render_template("register.html", writer, template_values{message, false})
}

// Register URL pointing for submitting POST request form
func register_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	form_username := request.FormValue("username")
	form_password := request.FormValue("password")
	form_confirm := request.FormValue("confirm")

	// Passwords don't match
	if form_password != form_confirm {
		set_cookie(writer, "message", "Error: Passwords don't match")
		redirect(writer, request, "/register/")
		return
	}

	// Validate usernames before sending to SQL to avoid injection
	valid := regexp.MustCompile("^[a-zA-Z0-9 _-]+$")
	if !valid.MatchString(form_username) {
		set_cookie(writer, "message", "Error: Username must match '^[a-zA-Z0-9 _-]+$'")
		redirect(writer, request, "/register/")
		return
	}

	// Check if username taken
	rows, err := db.Query("SELECT * FROM accounts WHERE username = ?", form_username)
	handle(err)
	defer rows.Close()
	if rows.Next() {
		set_cookie(writer, "message", "Error: Username taken")
		redirect(writer, request, "/register/")
		return
	}

	// Add user to database
	_, err = db.Exec("INSERT INTO accounts (username, password) VALUES (?, ?)", form_username, form_password)
	handle(err)

	// Log in
	set_cookie(writer, "username", form_username)
	set_cookie(writer, "message", "Successfully registered")
	redirect(writer, request, "/")
}

// Profile
func profile_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {
	username, _ := get_template_values(request)

	// Not logged in
	if username == "" {
		redirect(writer, request, "/")
		return
	}

	// Find user's win count
	// Breaks if the user injects a non-existent username as browser cookie

	rows, err := db.Query("SELECT * FROM accounts WHERE username = ?", username)

	var current account
	rows.Next()
	err = rows.Scan(&current.ID, &current.username, &current.password, &current.wins)
	handle(err)

	// Send data to template
	// Can't use render_template because it needs template_values
	err = templates.ExecuteTemplate(writer, "profile.html", profile{username, current.wins})
	handle(err)
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	username, _ := get_template_values(request)
	if username == "" {
		redirect(writer, request, "/")
		return
	}
	current := lobby{}

	// Get list of games from database
	rows, err := db.Query("SELECT * FROM games")
	handle(err)
	defer rows.Close()

	for rows.Next() {
		var name, password string
		var id int

		err = rows.Scan(&id, &name, &password)
		handle(err)

		current.Games = append(current.Games, name)
	}

	err = templates.ExecuteTemplate(writer, "lobby.html", current)
	handle(err)
}

// /join/x, accessed when pressing "Join" on a game
func join_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	username, _ := get_template_values(request)
	if username == "" {
		redirect(writer, request, "/")
		return
	}
	current := lobby{}

	path := strings.Split(request.URL.Path, "/")
	if len(path) != 4 {
		redirect(writer, request, "/")
		return
	}

	// Get list of games from database
	rows, err := db.Query("SELECT * FROM games WHERE name = ?", path[2])
	handle(err)
	defer rows.Close()
	if rows.Next() {
		var name, password string
		var id int

		err = rows.Scan(&id, &name, &password)
		handle(err)

		current.Games = append(current.Games, name)

	// No game with that ID
	} else {
		set_cookie(writer, "message", "Error: Game not found")
		redirect(writer, request, "/")
		return
	}

	set_cookie(writer, "message", "Game found!")
	redirect(writer, request, "/")

	// err = templates.ExecuteTemplate(writer, "lobby.html", current)
	// handle(err)
}

// Log out
func logout_handler(writer http.ResponseWriter, request *http.Request) {
	username, _ := get_template_values(request)

	set_cookie(writer, "username", "")
	if username != "" {
		set_cookie(writer, "message", "Successfully logged out")
	}

	redirect(writer, request, "/")
}

func main() {
	// Database setup
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

	// Static file serving
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
	http.HandleFunc("/logout/", logout_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
