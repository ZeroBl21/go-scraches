package log

import (
	"errors"
	"io"
	"os"
	"testing"

	api "github.com/ZeroBl21/dsg/ch01/proglog/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	for sceneario, fn := range map[string]func(
		t *testing.T,
		log *Log,
	){
		"append and read a record succeeds": testAppendRead,
		"offset out of range error":         testOutOfRangeErr,
		"init with existing segments":       testInitExisting,
		"reader":                            testReader,
		"truncate":                          testTruncate,
	} {
		t.Run(sceneario, func(t *testing.T) {
			dir := t.TempDir()
			defer os.RemoveAll(dir)

			cfg := Config{}
			cfg.Segment.MaxStoreBytes = 32

			log, err := NewLog(dir, cfg)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	append := &api.Record{Value: []byte("hello world")}

	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, append.Value, read.Value)
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	read, err := log.Read(1)
	require.Nil(t, read)

	var apiErr api.ErrOffsetOutOfRange
	require.True(t, errors.As(err, &apiErr))

	require.Equal(t, uint64(1), apiErr.Offset)
}

func testInitExisting(t *testing.T, o *Log) {
	append := &api.Record{Value: []byte("hello world")}

	for i := 0; i < 3; i++ {
		_, err := o.Append(append)
		require.NoError(t, err)
	}
	require.NoError(t, o.Close())

	off, err := o.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = o.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

	n, err := NewLog(o.Dir, o.Config)
	require.NoError(t, err)

	off, err = n.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = n.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)
}

func testReader(t *testing.T, log *Log) {
	append := &api.Record{Value: []byte("hello world")}

	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	reader := log.Reader()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, append.Value, read.Value)
}

func testTruncate(t *testing.T, log *Log) {
	append := &api.Record{Value: []byte("hello world")}

	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err)
}
