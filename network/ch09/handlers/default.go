package handlers

import (
	"html/template"
	"io"
	"net/http"
)

var templ = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandlers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r io.ReadCloser) {
			io.Copy(io.Discard, r)
			r.Close()
		}(r.Body)

		var b []byte

		switch r.Method {
		case http.MethodGet:
			b = []byte("friend")
		case http.MethodPost:
			var err error
			b, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		templ.Execute(w, string(b))
	})
}
