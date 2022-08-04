package main

import "net/http"

func setCookie(writer http.ResponseWriter, name, value string) {
	http.SetCookie(writer, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

func getCookie(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)

	if err != nil {
		return ""
	}
	return cookie.Value
}
