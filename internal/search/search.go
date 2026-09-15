package search

import (
	"encoding/json"
	"io"
)

// FileName is the name of the index file under the static directory.
const FileName = "search-index.js"

// Entry is one searchable symbol.
type Entry struct {
	// Name is the symbol name, with the receiver type for methods.
	Name string `json:"name"`

	// Kind is the symbol kind: package, type, func, method, const, or var.
	Kind string `json:"kind"`

	// Package is the import path that declares the symbol.
	Package string `json:"package"`

	// Path is the page path relative to the output root.
	Path string `json:"path"`

	// Unexported marks a symbol that cannot be named outside its package.
	// Such symbols rank after the exported ones with the same score.
	Unexported bool `json:"unexported,omitempty"`
}

// Index accumulates entries.
type Index struct {
	entries []Entry
}

// Add records an entry.
func (i *Index) Add(e Entry) {
	i.entries = append(i.entries, e)
}

// Len returns the number of entries.
func (i *Index) Len() int {
	return len(i.entries)
}

// Write writes the index as a script that assigns the entries to
// window.godoc2SearchIndex. It returns any error from encoding or writing.
func (i *Index) Write(w io.Writer) error {
	if _, err := io.WriteString(w, "window.godoc2SearchIndex = "); err != nil {
		return err
	}
	entries := i.entries
	if entries == nil {
		entries = []Entry{}
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	_, err = io.WriteString(w, ";\n")
	return err
}
