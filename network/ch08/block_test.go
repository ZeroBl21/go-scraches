package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// func TestBlockIndefinitely(t *testing.T) {
// 	testServer := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
// 	_, _ = http.Get(testServer.URL)
//
// 	t.Fatal("client did not indefinitely block")
// }

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(blockIndefinitely))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		testServer.URL,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	resp.Body.Close()
}

func blockIndefinitely(_ http.ResponseWriter, _ *http.Request) {
	select {}
}
