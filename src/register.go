package main

import (
	"database/sql"
	"net/http"
	"regexp"
)

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
