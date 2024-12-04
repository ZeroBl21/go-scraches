package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
	handler := http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			time.Sleep(time.Minute)
		}),
		time.Second,
		"Timed out while reading response",
	)

	req := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	wrec := httptest.NewRecorder()
	handler.ServeHTTP(wrec, req)

	resp := wrec.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code: %q", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if actual := string(body); actual != "Timed out while reading response" {
		t.Logf("unexpected body: %q", actual)
	}
}
