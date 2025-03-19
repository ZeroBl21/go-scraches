package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	write = []byte("hello world")
	width = uint64(len(write)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	file, err := os.CreateTemp("", "store_append_read_test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	s, err := newStore(file)
	require.NoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)

	s, err = newStore(file)
	require.NoError(t, err)
	testRead(t, s)
}

func testAppend(t *testing.T, s *store) {
	t.Helper()

	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(write)
		require.NoError(t, err)
		require.Equal(t, pos+n, width*i)
	}
}

func TestStoreClose(t *testing.T) {
	file, err := os.CreateTemp("", "store_close_test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	s, err := newStore(file)
	require.NoError(t, err)

	_, _, err = s.Append(write)
	require.NoError(t, err)

	file, beforeSize, err := openFile(t, file.Name())
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, afterSize, err := openFile(t, file.Name())
	require.NoError(t, err)
	require.True(t, afterSize > beforeSize)
}

func openFile(
	t *testing.T,
	filename string,
) (*os.File, int64, error) {
	t.Helper()

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}

	fs, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}

	return f, fs.Size(), nil
}

func testRead(t *testing.T, s *store) {
	t.Helper()

	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, write, read)

		pos += width
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()

	for i, off := uint64(1), int64(0); i < 4; i++ {
		buf := make([]byte, lenWidth)

		n, err := s.ReadAt(buf, off)
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)

		off += int64(n)
		size := enc.Uint64(buf)

		buf = make([]byte, size)
		n, err = s.ReadAt(buf, off)
		require.NoError(t, err)
		require.Equal(t, write, buf)
		require.Equal(t, int(size), n)

		off += int64(n)
	}
}
