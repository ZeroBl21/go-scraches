package ch13

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type wideResponseWriter struct {
	http.ResponseWriter
	lenght, status int
}

func WideEventLog(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wideHeader := &wideResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wideHeader, r)

		addr, _, _ := net.SplitHostPort(r.RemoteAddr)
		logger.Info("example wide event",
			zap.Int("status_code", wideHeader.status),
			zap.Int("response_length", wideHeader.lenght),
			zap.Int64("content_length", r.ContentLength),
			zap.String("method", r.Method),
			zap.String("proto", r.Proto),
			zap.String("remote_addr", addr),
			zap.String("uri", r.RequestURI),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

func (w *wideResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (w *wideResponseWriter) Write(buf []byte) (int, error) {
	n, err := w.ResponseWriter.Write(buf)
	w.lenght += n

	if w.status == 0 {
		w.status = http.StatusOK
	}

	return n, err
}

func Example_wideLogEntry() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
	)
	defer zl.Sync()

	ts := httptest.NewServer(
		WideEventLog(zl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(r io.ReadCloser) {
				io.Copy(io.Discard, r)
				r.Close()
			}(r.Body)
			w.Write([]byte("Hello!"))
		},
		)),
	)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test")
	if err != nil {
		zl.Fatal(err.Error())
	}
	resp.Body.Close()

	// Output:
	// {"level":"info","msg":"example wide event","status_code":200,"response_length":6,"content_length":0,"method":"GET","proto":"HTTP/1.1","remote_addr":"127.0.0.1","uri":"/test","user_agent":"Go-http-client/1.1"}
}
