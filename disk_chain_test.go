package markov

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

func TestDiskChainWriter(t *testing.T) {
	f, cleanup := tempFile(t)
	defer cleanup()

	writer, err := NewDiskChainWriter(f)
	if err != nil {
		t.Fatalf("NewDiskChainWriter failed: %v", err)
	}

	testReadWriteChain(t, writer)

	writer2, err := OpenDiskChainWriter(f)
	if err != nil {
		t.Fatalf("OpenDiskChainWriter failed: %v", err)
	}

	testReadChain(t, writer2)
}

func TestDiskChain(t *testing.T) {
	f, cleanup := tempFile(t)
	defer cleanup()

	writer, err := NewDiskChainWriter(f)
	if err != nil {
		t.Fatalf("NewDiskChainWriter failed: %v", err)
	}

	testWriteChain(t, writer)

	reader, err := ReadDiskChain(f)
	if err != nil {
		t.Fatalf("NewDiskChain failed: %v", err)
	}

	testReadChain(t, reader)
}

func TestDiskCopy(t *testing.T) {
	src := &MemoryChain{}
	testWriteChain(t, src)

	f, cleanup := tempFile(t)
	defer cleanup()

	dest, err := NewDiskChainWriter(f)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = Copy(dest, src)
	if err != nil {
		t.Fatalf("Copy failed with error: %v", err)
	}

	testReadChain(t, dest)
}

func BenchmarkDiskCopy(b *testing.B) {
	src := NewMemoryChain(b.N)
	err := Feed(src, normalDistGenerator(b.N, b.N*2))
	if err != nil {
		b.Fatalf("got error: %v", err)
	}

	f, cleanup := tempFile(b)
	defer cleanup()

	dest, err := NewDiskChainWriter(f)
	if err != nil {
		b.Fatalf("error: %v", err)
	}

	b.ResetTimer()

	err = Copy(dest, src)
	if err != nil {
		b.Fatalf("Copy failed with error: %v", err)
	}
}
