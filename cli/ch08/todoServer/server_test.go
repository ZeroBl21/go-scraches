package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupAPI(t *testing.T) (string, func()) {
	t.Helper()

	ts := httptest.NewServer(newMux(""))

	return ts.URL, func() {
		ts.Close()
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		expCode    int
		expItems   int
		expContent string
	}{
		{
			name: "GetRoot", path: "/",
			expCode:    http.StatusOK,
			expContent: "There's an API here",
		},
		{
			name: "NotFound", path: "/this/dont/exists",
			expCode: http.StatusNotFound,
		},
	}

	url, clenup := setupAPI(t)
	defer clenup()

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.Get(url + tt.path)
			if err != nil {
				t.Error(err)
			}
			defer r.Body.Close()

			if r.StatusCode != tt.expCode {
				t.Fatalf("Expected status code %q, got %q instead",
					http.StatusText(tt.expCode), http.StatusText(r.StatusCode))
			}

			switch {
			case strings.Contains(r.Header.Get("Content-Type"), "text/plain"):
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if !strings.Contains(string(body), tt.expContent) {
					t.Errorf("Expected %q, got %q.",
						tt.expContent, string(body))
				}

			default:
				t.Fatalf("Unsupported Content-Type: %q", r.Header.Get("Content-Type"))
			}
		})
	}
}
