package main

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})

	go func() {
		defer func() { sync <- struct{}{} }()

		dial := net.Dialer{
			Control: func(_, _ string, _ syscall.RawConn) error {
				time.Sleep(time.Second)
				return nil
			},
		}

		conn, err := dial.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}

		conn.Close()
		t.Error("connection did not time out")
	}()

	cancel()
	<-sync

	if ctx.Err() != context.Canceled {
		t.Errorf("expected cancel context; actual: %q", ctx.Err())
	}
}
