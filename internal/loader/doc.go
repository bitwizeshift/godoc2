// Package loader loads a Go module into the documentation [model].
//
// The loader runs [golang.org/x/tools/go/packages.Load] once for the given
// patterns, keeps the packages of the main module, and computes the exported
// documentation of each package with [go/doc]. Examples are read from the
// _test.go files next to each package.
//
// [model]: github.com/bitwizeshift/godoc2/internal/model
package loader
