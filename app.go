package main

import (
	"fmt"
	"os"
	"log"
	"net/http"
	"html/template"

	"github.com/go-sql-driver/mysql"
	"database/sql"
)

var templates = template.Must(template.ParseFiles(
	"html/index.html",
	"html/404.html",
	"html/login.html",
	"html/register.html",
	"html/lobby.html"))

type template_values struct {
	Message string
	LoggedIn bool
}

type account struct {
	ID int
	username string
	password string
}

func handle(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func set_cookie(writer http.ResponseWriter, name, value string) {
	http.SetCookie(writer, &http.Cookie{
		              Name: name,
		              Value: value,
		              Path: "/",
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

	if request.Method == http.MethodPost {

		form_username := request.FormValue("username")
		form_password := request.FormValue("password")

		rows, err := db.Query("SELECT * FROM accounts WHERE username = ?", form_username)
		handle(err)
		defer rows.Close()

		success := false
		for rows.Next() {
			var current account
			err = rows.Scan(&current.ID, &current.username, &current.password)
			handle(err)

			if current.password == form_password {
				success = true
			}
		}

		if success {
				set_cookie(writer, "username", form_username)
				set_cookie(writer, "message", "Successfully logged in")
				redirect(writer, request, "/")
		} else {
				set_cookie(writer, "message", "Error: Invalid credentials")
				redirect(writer, request, "/login/")
		}
	}
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

	// form_username := request.FormValue("username")
	form_password := request.FormValue("password")
	form_confirm := request.FormValue("confirm")

	// Passwords don't match
	if form_password != form_confirm {
		set_cookie(writer, "message", "Error: Passwords don't match")
		redirect(writer, request, "/register/")
	}

	set_cookie(writer, "message", "Request received")
	redirect(writer, request, "/register/")
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request) {

	signed_in := true
	username, message := get_template_values(request)
	if username == "" {
		signed_in = false
	}

	render_template("lobby.html", writer, template_values{message, signed_in})
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
		User:   "michael",
		Passwd: "password",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "db",
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
	http.HandleFunc("/login_post/", func (writer http.ResponseWriter, request *http.Request) {
		            login_post_handler(writer, request, db)
	})
	http.HandleFunc("/register/", register_get_handler)
	http.HandleFunc("/register_post/", func (writer http.ResponseWriter, request *http.Request) {
		            register_post_handler(writer, request, db)
	})
	http.HandleFunc("/lobby/", lobby_handler)
	http.HandleFunc("/logout/", logout_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
