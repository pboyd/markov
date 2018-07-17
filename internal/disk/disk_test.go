package disk

import (
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
