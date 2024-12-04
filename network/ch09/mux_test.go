package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimpleMux(t *testing.T) {
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	serveMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello friend.")
	})
	serveMux.HandleFunc("/hello/there/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Why, hello there.")
	})

	mux := drainAndClose(serveMux)

	testCases := []struct {
		path     string
		response string
		code     int
	}{
		{"http://test/", "", http.StatusNoContent},
		{"http://test/hello", "Hello friend.", http.StatusOK},
		{"http://test/hello/there/", "Why, hello there.", http.StatusOK},
		{
			"http://test/hello/there",
			"<a href=\"/hello/there/\">Moved Permanently</a>.\n\n",
			http.StatusMovedPermanently,
		},
		{"http://test/hello/there/you", "Why, hello there.", http.StatusOK},
		{"http://test/hello/and/goodbye", "", http.StatusNoContent},
		{"http://test/something/else/entirely", "", http.StatusNoContent},
		{"http://test/hello/you", "", http.StatusNoContent},
	}

	for idx, c := range testCases {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		wrec := httptest.NewRecorder()
		mux.ServeHTTP(wrec, req)

		resp := wrec.Result()
		if actual := resp.StatusCode; c.code != actual {
			t.Errorf("%d: expected code %d; actual %d", idx, c.code, actual)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		if actual := string(body); c.response != actual {
			t.Errorf("%d: expected response %q; actual %q", idx, c.response, actual)
		}
	}
}

func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		},
	)
}
