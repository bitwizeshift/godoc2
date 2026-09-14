package model

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"strings"
)

// Module is a loaded Go module together with its documented packages.
type Module struct {
	// Path is the module path from go.mod.
	Path string

	// Version is the module version, or empty for the main module.
	Version string

	// Dir is the root directory of the module on disk.
	Dir string

	// Packages are the packages of the module, sorted by import path.
	Packages []*Package

	// Fset is the file set shared by every package syntax tree.
	Fset *token.FileSet

	// Deps maps the import path of every loaded dependency package to the
	// module that provides it.
	Deps map[string]Dependency
}

// Dependency identifies the module that provides an imported package.
type Dependency struct {
	// Path is the module path.
	Path string

	// Version is the module version.
	Version string
}

// Package is one documented package of a [Module].
type Package struct {
	// ImportPath is the full import path.
	ImportPath string

	// Name is the package clause name.
	Name string

	// RelPath is the path relative to the module root, or empty for the root
	// package.
	RelPath string

	// Dir is the package directory on disk.
	Dir string

	// Doc is the raw package documentation.
	Doc string

	// Consts, Vars, Types, and Funcs hold the exported identifiers.
	Consts []*Value
	Vars   []*Value
	Types  []*Type
	Funcs  []*Func

	// Examples are the package-level examples.
	Examples []*Example

	// Files are the Go source files of the package, sorted by name.
	Files []*File

	// TypesPkg and Info are the type-checked package and its type information.
	TypesPkg *types.Package
	Info     *types.Info
}

// Internal reports whether the package sits under an internal directory.
func (p *Package) Internal() bool {
	for elem := range strings.SplitSeq(p.RelPath, "/") {
		if elem == "internal" {
			return true
		}
	}
	return false
}

// Tool reports whether the package is a main package.
func (p *Package) Tool() bool {
	return p.Name == "main"
}

// DisplayName returns the name used in headings and tables. Main packages are
// named after their directory.
func (p *Package) DisplayName() string {
	if p.Tool() {
		return path.Base(p.ImportPath)
	}
	return p.Name
}

// Summary returns the first paragraph of the package documentation.
func (p *Package) Summary() string {
	return FirstParagraph(p.Doc)
}

// TypeKind classifies a [Type] by its underlying declaration.
type TypeKind int

// Recognised values for [TypeKind].
const (
	KindOther TypeKind = iota
	KindStruct
	KindInterface
	KindAlias
)

// String returns the keyword shown in page headings for the kind.
func (k TypeKind) String() string {
	switch k {
	case KindStruct:
		return "struct"
	case KindInterface:
		return "interface"
	case KindAlias:
		return "alias"
	default:
		return "type"
	}
}

// Type is an exported type declaration.
type Type struct {
	Name string
	Kind TypeKind
	Doc  string

	Decl *ast.GenDecl
	Spec *ast.TypeSpec
	Obj  *types.TypeName

	// Methods are the exported methods declared on the type.
	Methods []*Func

	Examples []*Example

	Pkg *Package
}

// Summary returns the first paragraph of the type documentation.
func (t *Type) Summary() string {
	return FirstParagraph(t.Doc)
}

// Func is an exported function or method.
type Func struct {
	Name string
	Doc  string

	Decl *ast.FuncDecl
	Obj  *types.Func

	// Recv is the receiver type for methods, or nil for functions.
	Recv *Type

	Examples []*Example

	Pkg *Package
}

// Summary returns the first paragraph of the function documentation.
func (f *Func) Summary() string {
	return FirstParagraph(f.Doc)
}

// ValueKind distinguishes constants from variables.
type ValueKind int

// Recognised values for [ValueKind].
const (
	KindConst ValueKind = iota
	KindVar
)

// String returns the keyword for the kind.
func (k ValueKind) String() string {
	if k == KindVar {
		return "var"
	}
	return "const"
}

// Value is an exported constant or variable.
type Value struct {
	Name string
	Kind ValueKind
	Doc  string

	Decl *ast.GenDecl
	Spec *ast.ValueSpec
	Obj  types.Object

	Examples []*Example

	Pkg *Package
}

// Summary returns the first paragraph of the value documentation.
func (v *Value) Summary() string {
	return FirstParagraph(v.Doc)
}

// Example is a runnable example from a _test.go file.
type Example struct {
	// Name is the full example name, such as "ExampleFoo_bar".
	Name string

	// Suffix is the lower-case suffix after the underscore, or empty.
	Suffix string

	// Doc is the example documentation.
	Doc string

	// Code is the formatted example body.
	Code string

	// Output is the expected output, or empty.
	Output string
}

// File is one Go source file of a package.
type File struct {
	// Name is the base file name.
	Name string

	// Path is the absolute path on disk.
	Path string
}

// FirstParagraph returns the text up to the first blank line of doc.
func FirstParagraph(doc string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(doc), "\n\n")
	return first
}
