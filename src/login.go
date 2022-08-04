package main

import (
	"database/sql"
	"net/http"
)

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
