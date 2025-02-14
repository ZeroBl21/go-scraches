package main

import "net/http"

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content := "There's an API here"

	replyTextContent(w, r, http.StatusOK, content)
}

func replyTextContent(w http.ResponseWriter, _ *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}
