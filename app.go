package main

import (
	"log"
	"net/http"
	"html/template"
)

type template_values struct {
	Url string
}

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {

	// Normal
	if request.URL.Path == "/" {
		page, _ := template.ParseFiles("html/index.html")
		page.Execute(writer, template_values{})

	// 404
	} else {
		page, _ := template.ParseFiles("html/404.html")
		page.Execute(writer, template_values{request.URL.Path})
	}
}

// Log in page
func login_handler(writer http.ResponseWriter, request *http.Request) {
	page, _ := template.ParseFiles("html/login.html")
	page.Execute(writer, template_values{})
}

// Register page
func register_handler(writer http.ResponseWriter, request *http.Request) {
	page, _ := template.ParseFiles("html/register.html")
	page.Execute(writer, template_values{})
}

// Lobby
func lobby_handler(writer http.ResponseWriter, request *http.Request) {
	page, _ := template.ParseFiles("html/lobby.html")
	page.Execute(writer, template_values{})
}

func main() {
	http.HandleFunc("/", home_handler)
	http.HandleFunc("/login/", login_handler)
	http.HandleFunc("/register/", register_handler)
	http.HandleFunc("/lobby/", lobby_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
