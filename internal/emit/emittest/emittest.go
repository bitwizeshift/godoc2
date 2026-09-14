package emittest

import (
	"bytes"
	"io"
	"sort"
	"sync"

	"github.com/bitwizeshift/godoc2/internal/emit"
)

// MemSink is an [emit.Sink] that keeps every created file in memory.
type MemSink struct {
	mu    sync.Mutex
	files map[string]*bytes.Buffer
}

// NewMemSink returns an empty in-memory sink.
func NewMemSink() *MemSink {
	return &MemSink{files: map[string]*bytes.Buffer{}}
}

// Create returns a writer that stores the content under path. Creating the
// same path twice replaces the earlier content.
func (s *MemSink) Create(path string) (io.WriteCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf := &bytes.Buffer{}
	s.files[path] = buf
	return nopCloser{buf}, nil
}

// Content returns the content written to path and whether the file exists.
func (s *MemSink) Content(path string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf, ok := s.files[path]
	if !ok {
		return "", false
	}
	return buf.String(), true
}

// Paths returns every created path, sorted.
func (s *MemSink) Paths() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	paths := make([]string, 0, len(s.files))
	for p := range s.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

var _ emit.Sink = (*MemSink)(nil)

// ErrSink returns an [emit.Sink] whose Create always fails with err.
func ErrSink(err error) emit.Sink {
	return errSink{err: err}
}

type errSink struct {
	err error
}

func (s errSink) Create(string) (io.WriteCloser, error) {
	return nil, s.err
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
