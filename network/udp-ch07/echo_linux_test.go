package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEchoServerUnixPacket(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixpacket")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	// Server
	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	if err := os.Chmod(socket, os.ModeSocket|0666); err != nil {
		t.Fatal(err)
	}

	// Client
	conn, err := net.Dial("unixpacket", rAddr.String())
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second))

	msg := []byte("ping")
	for i := 0; i < 3; i++ {
		if _, err := conn.Write(msg); err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}

	for i := 0; i < 3; i++ {
		if _, err := conn.Write(msg); err != nil {
			t.Fatal(err)
		}
	}

	buf = make([]byte, 2)
	for i := 0; i < 3; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(msg[:2], buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg[:2], buf[:n])
		}
	}
}
