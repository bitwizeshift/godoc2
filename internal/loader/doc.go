// Package loader loads Go modules into the documentation [model].
//
// The loader runs [golang.org/x/tools/go/packages.Load] once for the given
// patterns, groups the matched packages by the module that owns them, and
// computes the exported documentation of each package with [go/doc].
// Examples are read from the _test.go files next to each package. A go.work
// file can supply the patterns, one "<dir>/..." per module the file uses.
//
// [model]: github.com/bitwizeshift/godoc2/internal/model
package loader
