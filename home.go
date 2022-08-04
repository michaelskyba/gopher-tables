package main

import "net/http"

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
