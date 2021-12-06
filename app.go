package main

import (
	"fmt"
	"log"
	"net/http"
	// "html/template"
)

// Home page
func home_handler(writer http.ResponseWriter, request *http.Request) {
	var HTML string

	if request.URL.Path == "/" {
		HTML = "<h1>Welcome to Gopher Tables!</h1><a href='/login'>Log in</a>"
	} else {
		HTML = fmt.Sprintf("<h1>404 - Page not found: %s</h1>", request.URL.Path)
	}

	fmt.Fprintf(writer, HTML)
	// fmt.Println(request)
}

// Log in page
func login_handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, request.URL.Path)
}

// Register page
func register_handler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, request.URL.Path)
}

func main() {
	http.HandleFunc("/", home_handler)
	http.HandleFunc("/login/", login_handler)
	http.HandleFunc("/register/", register_handler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
