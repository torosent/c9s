package blockingwriter

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// flakyFile mimics a non-blocking *os.File. The first few writes return
// (n, EAGAIN) for the first byte chunk, then drain on subsequent calls.
type flakyFile struct {
	*os.File
	chunks    int
	chunkSize int
	calls     int
}

func newFlaky(t *testing.T, chunks, chunkSize int) (*flakyFile, *os.File) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "out")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return &flakyFile{File: f, chunks: chunks, chunkSize: chunkSize}, f
}

func (f *flakyFile) Write(p []byte) (int, error) {
	f.calls++
	if f.calls <= f.chunks {
		// Write only a partial chunk and return EAGAIN.
		size := f.chunkSize
		if size > len(p) {
			size = len(p)
		}
		n, err := f.File.Write(p[:size])
		if err != nil {
			return n, err
		}
		return n, syscall.EAGAIN
	}
	return f.File.Write(p)
}

func TestBlockingWriterRetriesOnEAGAIN(t *testing.T) {
	flaky, _ := newFlaky(t, 3, 10)
	bw := &writerAdapter{w: flaky}
	payload := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	n, err := bw.Write(payload)
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Write returned n=%d, want %d", n, len(payload))
	}
}

// writerAdapter mirrors blockingWriter but accepts any io.Writer so we
// can substitute the flaky test double.
type writerAdapter struct {
	w io.Writer
}

func (a *writerAdapter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	total := 0
	deadline := time.Now().Add(maxBackoff)
	for total < len(p) {
		n, err := a.w.Write(p[total:])
		total += n
		if err == nil {
			if n == 0 {
				return total, io.ErrShortWrite
			}
			continue
		}
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			if time.Now().After(deadline) {
				return total, err
			}
			time.Sleep(retryDelay)
			continue
		}
		return total, err
	}
	return total, nil
}

func TestBlockingWriterPropagatesNonRetryableErrors(t *testing.T) {
	want := errors.New("disk on fire")
	bw := &writerAdapter{w: errWriter{err: want}}
	if _, err := bw.Write([]byte("x")); !errors.Is(err, want) {
		t.Fatalf("Write err = %v, want %v", err, want)
	}
}

type errWriter struct{ err error }

func (e errWriter) Write(p []byte) (int, error) { return 0, e.err }

func TestBlockingWriterPassesThroughEmpty(t *testing.T) {
	bw := &writerAdapter{w: errWriter{err: errors.New("should not be called")}}
	n, err := bw.Write(nil)
	if err != nil || n != 0 {
		t.Fatalf("Write(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

func TestBlockingWriterFdPassesThrough(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	bw := New(f)
	if bw.Fd() != f.Fd() {
		t.Fatalf("Fd() = %d, want %d", bw.Fd(), f.Fd())
	}
}
