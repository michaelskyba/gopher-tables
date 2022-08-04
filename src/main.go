package main

import (
	"html/template"
	"net/http"

	"database/sql"
	"log"

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

func redirect(writer http.ResponseWriter, request *http.Request, path string) {
	http.Redirect(writer, request, path, http.StatusSeeOther)
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
