package disk

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func tempFile(t testing.TB) (*os.File, func()) {
	path := filepath.Join(
		os.TempDir(),
		strconv.Itoa(int(time.Now().UnixNano())),
	)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("could not create file %q: %v", path, err)
	}

	cleanup := func() {
		f.Close()
		os.Remove(path)
	}

	return f, cleanup
}

func TestReadAppendWrite(t *testing.T) {
	rw, cleanup := tempFile(t)
	defer cleanup()

	const iterations = 10

	buf := make([]byte, 4)

	for i := uint32(0); i < iterations; i++ {
		binary.BigEndian.PutUint32(buf, i)

		off, err := Write(rw, -1, buf)
		if err != nil {
			t.Fatalf("Write error: %v", err)
		}

		if off != int64(i*4) {
			t.Errorf("got offset %d, want %d", off, i*4)
		}
	}

	for i := uint32(0); i < iterations; i++ {
		buf, err := Read(rw, int64(i*4), 4)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}

		actual := binary.BigEndian.Uint32(buf)
		if actual != i {
			t.Errorf("got %d, want %d", actual, i)
		}
	}

	var val uint32 = 1<<32 - 1
	binary.BigEndian.PutUint32(buf, val)
	_, err := Write(rw, 4, buf)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	buf, err = Read(rw, 4, 4)
	if err != nil {
		t.Errorf("Read error: %v", err)
	}

	actual := binary.BigEndian.Uint32(buf)
	if actual != val {
		t.Errorf("got %d, want %d", actual, val)
	}
}

func TestSectionHeader(t *testing.T) {
	cases := []struct {
		t   sectionType
		len uint32
	}{
		{0xe, 0xabcdef},
		{recordSection, 123},
		{0, 1<<24 - 1},
	}

	for _, c := range cases {
		buf := make([]byte, 4)
		putSectionHeader(buf, c.t, c.len)

		typ, len := sectionHeader(buf)

		if typ != c.t {
			t.Errorf("got %d, want %d", typ, c.t)
		}

		if len != c.len {
			t.Errorf("got %x, want %x", len, c.len)
		}
	}
}
