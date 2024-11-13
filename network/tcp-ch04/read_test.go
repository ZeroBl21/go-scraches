package main

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	payload := make([]byte, 1<<24) // 16 MB

	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != err {
			t.Log(err)
			return
		}
		defer conn.Close()

		if _, err = conn.Write(payload); err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1<<19) // 512 KB

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}

			break
		}

		t.Logf("read %d bytes", n)
	}

	conn.Close()
}
