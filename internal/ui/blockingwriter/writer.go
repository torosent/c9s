// Package blockingwriter wraps an *os.File with retry logic for partial
// writes and EAGAIN errors so it behaves as if it were a fully blocking
// writer.
//
// This is needed because bubbletea v2's renderer issues large frames
// (~10 KB) in a single Write call. After tea.ExecProcess restores the
// terminal on macOS, the underlying *os.File for stdout can be left in
// non-blocking mode, causing writes that exceed the kernel TTY buffer
// (~1 KB) to return EAGAIN with a short count. The renderer treats that
// as a fatal error, drops the rest of the frame, and never recovers —
// which manifests as a TUI showing only a few rows of a stale frame
// after the user exits an exec'd shell.
//
// The wrapper preserves the underlying file descriptor (via Fd) and the
// io.ReadWriteCloser surface so bubbletea's terminal-detection code
// (which type-asserts to a term.File interface) keeps working.
package blockingwriter

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"
)

// File mirrors github.com/charmbracelet/x/term.File so callers don't
// have to import that package just to express the interface bubbletea
// expects from p.output.
type File interface {
	io.ReadWriteCloser
	Fd() uintptr
}

// New returns a File that wraps f and retries on EAGAIN / EWOULDBLOCK /
// short writes until all bytes are written. Read, Close and Fd are
// passed through unchanged.
func New(f *os.File) File {
	return &blockingWriter{f: f}
}

type blockingWriter struct {
	f *os.File
}

// retryDelay is how long to sleep between EAGAIN retries.
// 100 µs is short enough to be unnoticeable in interactive use yet
// long enough to let the kernel drain the TTY buffer.
const retryDelay = 100 * time.Microsecond

// maxBackoff caps the per-call wait so a permanently-blocked writer
// can't hang the renderer indefinitely.
const maxBackoff = 250 * time.Millisecond

func (b *blockingWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	total := 0
	deadline := time.Now().Add(maxBackoff)
	for total < len(p) {
		n, err := b.f.Write(p[total:])
		total += n
		if err == nil {
			if n == 0 {
				// Shouldn't happen for a regular file, but bail out
				// instead of looping forever.
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

func (b *blockingWriter) Read(p []byte) (int, error) { return b.f.Read(p) }
func (b *blockingWriter) Close() error               { return b.f.Close() }
func (b *blockingWriter) Fd() uintptr                { return b.f.Fd() }

