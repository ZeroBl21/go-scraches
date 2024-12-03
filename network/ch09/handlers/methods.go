package handlers

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"
)

type Methods map[string]http.Handler

func (h Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(req io.ReadCloser) {
		io.Copy(io.Discard, req)
		req.Close()
	}(r.Body)

	if handler, ok := h[r.Method]; ok {
		if handler == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		handler.ServeHTTP(w, r)
		return
	}

	w.Header().Add("Allow", h.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Methods) allowedMethods() string {
	allowed := make([]string, 0, len(h))

	for key := range h {
		allowed = append(allowed, key)
	}
	sort.Strings(allowed)

	return strings.Join(allowed, ", ")
}

func DefaultMethodsHandler() http.Handler {
	return Methods{
		http.MethodGet: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello, friend!"))
			},
		),
		http.MethodPost: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "Hello, %s!", html.EscapeString(string(b)))
			},
		),
	}
}
