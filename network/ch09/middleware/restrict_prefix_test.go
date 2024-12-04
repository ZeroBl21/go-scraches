package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
	handler := http.StripPrefix(
		"/static/",
		RestrictPrefix(".", http.FileServer(http.Dir("../files/"))),
	)

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for idx, c := range testCases {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		wrec := httptest.NewRecorder()
		handler.ServeHTTP(wrec, req)

		actual := wrec.Result().StatusCode
		if c.code != actual {
			t.Errorf("%d: expected %d; actual %d", idx, c.code, actual)
		}
	}
}
