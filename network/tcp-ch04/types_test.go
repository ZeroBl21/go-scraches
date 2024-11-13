package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
	binary1 := Binary("Clear is better than clever.")
	binary2 := Binary("Don't panic.")

	string1 := String("Errors are values.")

	payloads := []Payload{&binary1, &binary2, &string1}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		for _, payload := range payloads {
			if _, err := payload.WriteTo(conn); err != nil {
				t.Error(err)
				break
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T %[1]q", actual)
	}
}

func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := buf.WriteByte(BinaryType); err != nil {
		t.Fatal(err)
	}

	err := binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 GB
	if err != nil {
		t.Fatal(err)
	}

	var bin Binary

	if _, err := bin.ReadFrom(buf); err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
