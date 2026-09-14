package emit

import (
	"io"
	"os"
	"path/filepath"
)

// Sink creates output files by slash-separated relative path.
type Sink interface {
	// Create creates the file at path, together with any missing parent
	// directories, and returns a writer for its content. The caller closes the
	// writer when the file is complete.
	Create(path string) (io.WriteCloser, error)
}

// DirSink is a [Sink] that writes files under a directory on disk.
type DirSink struct {
	root string
}

// NewDirSink returns a [Sink] rooted at root.
func NewDirSink(root string) *DirSink {
	return &DirSink{root: root}
}

// Create creates the file under the sink root. It returns any error from
// creating the parent directories or the file.
func (s *DirSink) Create(path string) (io.WriteCloser, error) {
	full := filepath.Join(s.root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return nil, err
	}
	return os.Create(full)
}

var _ Sink = (*DirSink)(nil)
