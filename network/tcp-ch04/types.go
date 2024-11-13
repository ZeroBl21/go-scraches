package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

type Binary []byte

func (m Binary) Bytes() []byte  { return m }
func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, BinaryType)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	n += 4

	payloadSize, err := w.Write(m)

	return n + int64(payloadSize), err
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var n int64 = 1

	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}

	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*m = make([]byte, size)
	payloadSize, err := r.Read(*m)

	return n + int64(payloadSize), err
}

type String string

func (m String) Bytes() []byte  { return []byte(m) }
func (m String) String() string { return string(m) }

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType)
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return 0, err
	}
	n += 4

	payload, err := w.Write([]byte(m))

	return n + int64(payload), err
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var n int64 = 1

	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, nil
	}
	n += 4

	buf := make([]byte, size)
	payload, err := r.Read(buf)
	if err != nil {
		return n, err
	}

	*m = String(buf)

	return n + int64(payload), nil
}

func decode(r io.Reader) (Payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload

	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}

	_, err = payload.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
