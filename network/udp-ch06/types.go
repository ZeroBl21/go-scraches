package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	DatagramSize = 516              // Maximum supperted datagram size
	BlockSize    = DatagramSize - 4 // Datagram minus 4 bytes header
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1
	_
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const (
	ErrUnkown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	EWrrIllegalOp
	ErrUnkownID
	ErrFileExists
	ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode     string
}

func (rr ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if rr.Mode != "" {
		mode = rr.Mode
	}

	// OpCode + Filename + 0b + Mode + 0b
	cap := 2 + 2 + len(rr.Filename) + 1 + len(rr.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	if err := binary.Write(b, binary.BigEndian, OpRRQ); err != nil {
		return nil, err
	}

	if _, err := b.WriteString(rr.Filename); err != nil {
		return nil, err
	}

	if err := b.WriteByte(0); err != nil {
		return nil, err
	}

	if _, err := b.WriteString(mode); err != nil {
		return nil, err
	}

	if err := b.WriteByte(0); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (rr *ReadReq) UnmarhsalBinary(p []byte) error {
	reader := bytes.NewBuffer(p)

	var code OpCode

	err := binary.Read(reader, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	rr.Filename, err = reader.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove 0 Byte Delimiter
	rr.Filename = strings.TrimRight(rr.Filename, "\x00")
	if len(rr.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	rr.Mode, err = reader.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	rr.Mode = strings.TrimRight(rr.Mode, "\x00")
	if len(rr.Mode) == 0 {
		return errors.New("invalid RRQ")
	}

	actual := strings.ToLower(rr.Mode)
	if actual != "octet" {
		return errors.New("only binary transfers supperted")
	}

	return nil
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++

	if err := binary.Write(b, binary.BigEndian, OpData); err != nil {
		return nil, err
	}

	if err := binary.Write(b, binary.BigEndian, d.Block); err != nil {
		return nil, err
	}

	_, err := io.CopyN(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil
}

func (d *Data) UnmarhsalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opcode OpCode

	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}

	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
	cap := 2 + 2 // OpCode + Block Number

	b := new(bytes.Buffer)
	b.Grow(cap)

	if err := binary.Write(b, binary.BigEndian, OpAck); err != nil {
		return nil, err
	}

	if err := binary.Write(b, binary.BigEndian, a); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a Ack) UnmarhsalBinary(payload []byte) error {
	var code OpCode

	reader := bytes.NewReader(payload)

	if err := binary.Read(reader, binary.BigEndian, &code); err != nil {
		return err
	}

	if code != OpAck {
		return errors.New("invalid ACK")
	}

	return binary.Read(reader, binary.BigEndian, a)
}

type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	// OpCode + Error Code + Message + 0b
	cap := 2 + 2 + len(e.Message) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	if err := binary.Write(b, binary.BigEndian, OpErr); err != nil {
		return nil, err
	}

	if err := binary.Write(b, binary.BigEndian, e.Error); err != nil {
		return nil, err
	}

	if _, err := b.WriteString(e.Message); err != nil {
		return nil, err
	}

	if err := b.WriteByte(0); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e Err) UnmarhsalBinary(payload []byte) error {
	reader := bytes.NewBuffer(payload)

	var code OpCode

	if err := binary.Read(reader, binary.BigEndian, &code); err != nil {
		return err
	}

	if code != OpErr {
		return errors.New("invalid ERROR")
	}

	err := binary.Read(reader, binary.BigEndian, &e.Error)
	if err != nil {
		return err
	}

	e.Message, err = reader.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00")

	return err
}
