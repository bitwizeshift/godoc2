package progress

import (
	"fmt"
	"io"
)

// Reporter receives generation progress.
type Reporter interface {
	// Stage reports that the named stage started.
	Stage(name string)

	// File reports that the file at path was written.
	File(path string)
}

// Writer is a [Reporter] that prints to an [io.Writer].
type Writer struct {
	w       io.Writer
	verbose bool
}

// NewWriter returns a [Reporter] that prints stages to w. When verbose is
// true it also prints each written file.
func NewWriter(w io.Writer, verbose bool) *Writer {
	return &Writer{w: w, verbose: verbose}
}

// Stage prints the stage name. A write error is dropped: progress output is
// advisory and must not stop generation.
func (w *Writer) Stage(name string) {
	_, _ = fmt.Fprintln(w.w, name)
}

// File prints the file path, indented, when verbose output is enabled. A
// write error is dropped for the same reason as in [Writer.Stage].
func (w *Writer) File(path string) {
	if !w.verbose {
		return
	}
	_, _ = fmt.Fprintf(w.w, "  %s\n", path)
}

var _ Reporter = (*Writer)(nil)

// Discard returns a [Reporter] that drops every report.
func Discard() Reporter {
	return discard{}
}

type discard struct{}

func (discard) Stage(string) {}
func (discard) File(string)  {}
