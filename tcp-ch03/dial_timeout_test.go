package main

import (
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	dial := net.Dialer{
		Timeout: timeout,
		Control: func(_, addr string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection time out",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
	}

	return dial.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("connection did not time out")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Error(err)
	}

	if !netErr.Timeout() {
		t.Fatal("error is not timeout")
	}
}
