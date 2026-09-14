package render

import (
	"go/types"

	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/markdown/doclink"
	"github.com/bitwizeshift/godoc2/internal/model"
)

// packageScope resolves doc links written in the documentation of one
// package, from one output page.
type packageScope struct {
	resolver *link.Resolver
	pkg      *model.Package
	from     string
}

// Symbol resolves name in the current package, or in the imported package
// named or addressed by pkg.
func (s packageScope) Symbol(pkg, name, method string) (string, bool) {
	target := s.pkg.TypesPkg
	if pkg != "" {
		target = s.imported(pkg)
	}
	if target == nil {
		return "", false
	}
	return s.resolver.Lookup(s.from, target, name, method)
}

// Package resolves an imported package by name or import path.
func (s packageScope) Package(pkg string) (string, bool) {
	target := s.imported(pkg)
	if target == nil {
		return "", false
	}
	return s.resolver.PackageURL(s.from, target.Path()), true
}

// imported finds the imported package whose name or path is pkg.
func (s packageScope) imported(pkg string) *types.Package {
	if s.pkg.TypesPkg == nil {
		return nil
	}
	for _, imp := range s.pkg.TypesPkg.Imports() {
		if imp.Path() == pkg || imp.Name() == pkg {
			return imp
		}
	}
	if p := s.resolver.Package(pkg); p != nil {
		return p.TypesPkg
	}
	if p := s.resolver.PackageNamed(pkg); p != nil {
		return p.TypesPkg
	}
	return nil
}

var _ doclink.Scope = packageScope{}
