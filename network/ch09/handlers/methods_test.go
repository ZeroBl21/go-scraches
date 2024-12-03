package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestSimpleHTTPServer(t *testing.T) {
	srv := &http.Server{
		Addr: "127.0.0.1:8081",
		Handler: http.TimeoutHandler(
			DefaultMethodsHandler(),
			2*time.Minute,
			"",
		),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{
			http.MethodPost, bytes.NewBufferString("<world>"),
			http.StatusOK, "Hello, &lt;world&gt;!",
		},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}

	client := new(http.Client)
	path := fmt.Sprintf("http://%s/", srv.Addr)

	for i, c := range testCases {
		req, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}

		if resp.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
		}
		resp.Body.Close()

		if c.response != string(body) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, body)
		}
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
