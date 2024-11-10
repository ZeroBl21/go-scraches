package main

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	deadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	dial := net.Dialer{
		Control: func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(5*time.Second + time.Millisecond)
			return nil
		},
	}

	conn, err := dial.DialContext(ctx, "tcp", "10.0.0.0:80")
	if nil == err {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Error(err)
	}
	if !netErr.Timeout() {
		t.Errorf("error is not a timeout: %v", err)
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
