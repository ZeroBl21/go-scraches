package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	// Req 1 - This don't write the header
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
	}

	req := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()

	handler(w, req)
	t.Logf("Response status: %q", w.Result().Status)

	// Req 2 - This writes the header
	handler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}

	req = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()

	handler(w, req)
	t.Logf("Response status: %q", w.Result().Status)
}
