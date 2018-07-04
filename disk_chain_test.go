package markov

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func tempFile(t *testing.T) (io.ReadWriteSeeker, func()) {
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

func TestDiskChain(t *testing.T) {
	f, cleanup := tempFile(t)
	defer cleanup()

	writer, err := NewDiskChainWriter(f)
	if err != nil {
		t.Fatalf("NewDiskChainWriter failed: %v", err)
	}

	testReadWriteChain(t, writer)
}
