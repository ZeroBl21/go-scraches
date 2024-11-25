package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	DatagramSize = 516              // Maximum supported datagram size
	BlockSize    = DatagramSize - 4 // Datagram minus 4 bytes header
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1
	_            // no WRQ
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
	cap := 2 + 2 + len(rr.Filename) + 1 + len(mode) + 1

	buf := new(bytes.Buffer)
	buf.Grow(cap)

	if err := binary.Write(buf, binary.BigEndian, OpRRQ); err != nil {
		return nil, err
	}

	if _, err := buf.WriteString(rr.Filename); err != nil {
		return nil, err
	}

	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}

	if _, err := buf.WriteString(mode); err != nil {
		return nil, err
	}

	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (rr *ReadReq) UnmarshalBinary(payload []byte) error {
	reader := bytes.NewBuffer(payload)

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
		return errors.New("only binary transfers supported")
	}

	return nil
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Grow(DatagramSize)

	// Increment Block Number
	d.Block++

	if err := binary.Write(buf, binary.BigEndian, OpData); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, d.Block); err != nil {
		return nil, err
	}

	_, err := io.CopyN(buf, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *Data) UnmarshalBinary(payload []byte) error {
	if l := len(payload); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opCode OpCode

	err := binary.Read(bytes.NewReader(payload[:2]), binary.BigEndian, &opCode)
	if err != nil || opCode != OpData {
		return errors.New("invalid DATA")
	}

	err = binary.Read(bytes.NewReader(payload[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}

	d.Payload = bytes.NewBuffer(payload[4:])

	return nil
}

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
	// OpCode + Block
	cap := 2 + 2

	buf := new(bytes.Buffer)
	buf.Grow(cap)

	if err := binary.Write(buf, binary.BigEndian, OpAck); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, a); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(payload []byte) error {
	var opCode OpCode

	reader := bytes.NewReader(payload)

	if err := binary.Read(reader, binary.BigEndian, &opCode); err != nil {
		return err
	}

	if opCode != OpAck {
		return errors.New("invalid ACK")
	}

	return binary.Read(reader, binary.BigEndian, a)
}

type Err struct {
	Error   OpCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	// OpCode + ErrorCode + Message + 0b
	cap := 2 + 2 + len(e.Message) + 1

	buf := new(bytes.Buffer)
	buf.Grow(cap)

	if err := binary.Write(buf, binary.BigEndian, OpErr); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, e.Error); err != nil {
		return nil, err
	}

	if _, err := buf.WriteString(e.Message); err != nil {
		return nil, err
	}

	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *Err) UnmarshalBinary(payload []byte) error {
	reader := bytes.NewBuffer(payload)

	var opCode OpCode

	if err := binary.Read(reader, binary.BigEndian, &opCode); err != nil {
		return err
	}
	if opCode != OpErr {
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
