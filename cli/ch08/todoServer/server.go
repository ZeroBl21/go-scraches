package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ZeroBl21/cli/ch02/todo"
)

var list = todo.List{}

func newMux(todoFile string) http.Handler {
	m := http.NewServeMux()

	if err := list.Get(todoFile); err != nil {
		panic(err)
	}

	m.HandleFunc("GET /{$}", rootHandler)

	m.HandleFunc("/", http.NotFound)

	return m
}

func replyTextContent(w http.ResponseWriter, _ *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func replyJSON(w http.ResponseWriter, r *http.Request, status int, content *todoResponse) {
	body, err := json.Marshal(content)
	if err != nil {
		replyError(w, r, http.StatusInternalServerError, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func replyError(w http.ResponseWriter, r *http.Request, status int, error string) {
	log.Printf("%s: %s Error: %d %s", r.URL, r.Method, status, error)

	http.Error(w, http.StatusText(status), status)
}
