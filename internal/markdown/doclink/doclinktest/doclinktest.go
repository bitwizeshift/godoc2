package doclinktest

import (
	"strings"

	"github.com/bitwizeshift/godoc2/internal/markdown/doclink"
)

// MapScope is a [doclink.Scope] backed by fixed tables.
//
// Symbols are keyed by "pkg.Name", "pkg.Name.Method", "Name", or
// "Name.Method". Packages are keyed by their import path or name.
type MapScope struct {
	Symbols  map[string]string
	Packages map[string]string
}

// Symbol looks the symbol up in the Symbols table.
func (s MapScope) Symbol(pkg, name, method string) (string, bool) {
	key := strings.Join(nonEmpty(pkg, name, method), ".")
	url, ok := s.Symbols[key]
	return url, ok
}

// Package looks the package up in the Packages table.
func (s MapScope) Package(pkg string) (string, bool) {
	url, ok := s.Packages[pkg]
	return url, ok
}

var _ doclink.Scope = MapScope{}

// EmptyScope returns a [doclink.Scope] that resolves nothing.
func EmptyScope() doclink.Scope {
	return MapScope{}
}

func nonEmpty(parts ...string) []string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
