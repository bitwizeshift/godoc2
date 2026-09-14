// Package docfile finds the Markdown files that document a directory in
// place of a package doc comment.
//
// [Find] locates the file of one directory. The loader calls it for every
// package without a doc comment and for a module root without a package.
package docfile
