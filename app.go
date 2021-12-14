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
	Url string
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
	})
}

func render_template(filename string, writer http.ResponseWriter, values template_values) {
	err := templates.ExecuteTemplate(writer, filename, values)
	handle(err)
}

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {

	// Normal
	if request.URL.Path == "/" {
		render_template("index.html", writer, template_values{})

	// 404
	} else {
		values := template_values{request.URL.Path}
		render_template("404.html", writer, values)
	}
}

// Log in page
func login_get_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("login.html", writer, template_values{})
}

// Log in URL point for submitting the log in form
func login_post_handler(writer http.ResponseWriter, request *http.Request, db *sql.DB) {

	if request.Method == http.MethodPost {

		// fmt.Println(request.FormValue("username"))
		// fmt.Println(request.FormValue("password"))

		rows, err := db.Query("SELECT * FROM accounts;")
		handle(err)
		defer rows.Close()

		for rows.Next() {
			var current account
			err = rows.Scan(&current.ID, &current.username, &current.password)
			handle(err)
			// fmt.Println(current)
		}

		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}
}

// Register page
func register_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("register.html", writer, template_values{})
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("lobby.html", writer, template_values{})
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
	http.HandleFunc("/login_get/", login_get_handler)
	// http.HandleFunc("/login_post/", login_post_handler)
	http.HandleFunc("/login_post/", func (writer http.ResponseWriter, request *http.Request) {
		            login_post_handler(writer, request, db)
	})
	http.HandleFunc("/register/", register_handler)
	http.HandleFunc("/lobby/", lobby_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
