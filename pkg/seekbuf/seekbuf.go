// file as a buffer
package seekbuf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Buffer type
type Buffer struct {
	reader io.Reader
	data   []byte
	len    int64
	pos    int64
	mode   Mode
	f      *os.File
	lock   sync.Mutex
}

var _ io.ReaderAt = (*Buffer)(nil)
var _ io.ReadSeeker = (*Buffer)(nil)
var _ io.Closer = (*Buffer)(nil)

// Mode cache mode
type Mode int

const (
	// FileMode cache reader's data to file
	FileMode = Mode(0)
	// MemoryMode cache reader's data to memory
	MemoryMode = Mode(1)
)

var (
	// ErrModeNotFound mode not found
	ErrModeNotFound = errors.New("mode not found")
)

// Open buffer and use reader as data source
func Open(r io.Reader, m Mode) (*Buffer, error) {
	switch m {
	case FileMode:
		f, err := os.CreateTemp("", "seekbuf-")
		if err != nil {
			return nil, err
		}
		return &Buffer{reader: r, mode: m, f: f}, nil
	case MemoryMode:
		return &Buffer{reader: r, mode: m}, nil
	}
	return nil, ErrModeNotFound
}

// Close and release
func (b *Buffer) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.data = nil
	b.reader = nil
	if b.mode == FileMode {
		name := b.f.Name()
		b.f.Close()
		b.f = nil
		os.Remove(name)
	}
	return nil
}

func (s *Buffer) ReadAt(p []byte, off int64) (n int, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	total := off + int64(len(p))
	if total > s.len {
		more := total - s.len
		switch s.mode {
		case MemoryMode:
			buf := &bytes.Buffer{}
			rn, e := io.CopyN(buf, s.reader, more)
			err = e
			s.len += int64(rn)
			s.data = append(s.data, buf.Bytes()...)
		case FileMode:
			rn, e := io.CopyN(s.f, s.reader, more)
			err = e
			s.len += int64(rn)
		}
	}

	if s.mode == FileMode {
		return s.f.ReadAt(p, off)
	}
	return copy(p, s.data[off:]), err
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	n, err = b.ReadAt(p, b.pos)
	b.lock.Lock()
	b.pos += int64(n)
	b.lock.Unlock()
	return n, err
}

// Seek sets the offset for the next Read or Write on the buffer to offset, interpreted according to whence:
// 0 means relative to the origin of the buffer, 1 means relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	o := offset
	switch whence {
	case io.SeekCurrent:
		// if o > 0 && b.pos+o >= int64(len(b.data)) {
		// 	return -1, fmt.Errorf("invalid offset %d", offset)
		// }
		b.pos += o
	case io.SeekStart:
		// if o > 0 && o >= int64(len(b.data)) {
		// 	return -1, fmt.Errorf("invalid offset %d", offset)
		// }
		b.pos = o
	case io.SeekEnd:
		// if int64(len(b.data))+o < 0 {
		// 	return -1, fmt.Errorf("invalid offset %d", offset)
		// }
		b.pos = int64(len(b.data)) + o
	default:
		return -1, fmt.Errorf("invalid whence %d", whence)
	}

	return int64(b.pos), nil
}

// return current buffer len
func (b *Buffer) Size() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.len
}
