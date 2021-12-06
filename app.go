package main

import (
	"log"
	"net/http"
	"html/template"
)

type template_values struct {
	Url string
}

func render_template(filename string, writer http.ResponseWriter, values template_values) {
	page, _ := template.ParseFiles(filename)
	page.Execute(writer, values)
}

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {

	// Normal
	if request.URL.Path == "/" {
		render_template("html/index.html", writer, template_values{})

	// 404
	} else {
		values := template_values{request.URL.Path}
		render_template("html/404.html", writer, values)
	}
}

// Log in page
func login_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("html/login.html", writer, template_values{})
}

// Register page
func register_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("html/register.html", writer, template_values{})
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request) {
	render_template("html/lobby.html", writer, template_values{})
}

func main() {
	http.HandleFunc("/", home_handler)
	http.HandleFunc("/login/", login_handler)
	http.HandleFunc("/register/", register_handler)
	http.HandleFunc("/lobby/", lobby_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
