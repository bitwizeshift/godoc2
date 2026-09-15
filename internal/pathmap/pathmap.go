package pathmap

import (
	"go/token"
	"path"
	"strings"
)

// IndexFile is the file name of module and package pages.
const IndexFile = "index.html"

// StaticDir is the directory that holds the shared assets.
const StaticDir = "static"

// Root returns the path of the root page of the site.
func Root() string {
	return IndexFile
}

// Module returns the path of the module page.
func Module(modulePath string) string {
	return path.Join(modulePath, IndexFile)
}

// Package returns the path of a package page. rel is the package path relative
// to the module root, or empty for the root package.
func Package(modulePath, rel string) string {
	return path.Join(modulePath, rel, IndexFile)
}

// PackageDir returns the output directory of a package.
func PackageDir(modulePath, rel string) string {
	return path.Join(modulePath, rel)
}

// UnexportedMark precedes an unexported name in a page path. The mark keeps
// the page of an unexported name apart from the page of an exported name
// that differs only in case, on file systems that ignore case.
const UnexportedMark = "~"

// Symbol returns the path of a type, function, constant, or variable page.
func Symbol(modulePath, rel, name string) string {
	return path.Join(modulePath, rel, elem(name)+".html")
}

// Method returns the path of a method page.
func Method(modulePath, rel, typeName, name string) string {
	return path.Join(modulePath, rel, elem(typeName)+"."+elem(name)+".html")
}

// elem returns the path element of a symbol name.
func elem(name string) string {
	if token.IsExported(name) {
		return name
	}
	return UnexportedMark + name
}

// Source returns the path of a rendered source file page.
func Source(modulePath, rel, file string) string {
	return path.Join(modulePath, rel, file+".html")
}

// File returns the path of a file copied from the module directory. rel is
// the file path relative to the module root.
func File(modulePath, rel string) string {
	return path.Join(modulePath, rel)
}

// Static returns the path of a shared asset.
func Static(name string) string {
	return path.Join(StaticDir, name)
}

// Rel returns the relative href from the page at from to the page at to.
func Rel(from, to string) string {
	fromDir := path.Dir(from)
	if fromDir == "." {
		return to
	}
	fromParts := strings.Split(fromDir, "/")
	toParts := strings.Split(to, "/")

	common := 0
	for common < len(fromParts) && common < len(toParts)-1 && fromParts[common] == toParts[common] {
		common++
	}
	var b strings.Builder
	for range len(fromParts) - common {
		b.WriteString("../")
	}
	b.WriteString(strings.Join(toParts[common:], "/"))
	return b.String()
}
