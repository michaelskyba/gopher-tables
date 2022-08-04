package main

import (
	"database/sql"
	"net/http"
)

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
