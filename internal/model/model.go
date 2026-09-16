package model

import (
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"strings"
)

// Site is the set of modules that one documentation site covers.
type Site struct {
	// Dir is the directory the package patterns were resolved from.
	Dir string

	// Modules are the loaded modules, sorted by path.
	Modules []*Module

	// DocFile is the Markdown file that documents the site root. It is set
	// only when the site holds more than one module.
	DocFile *DocFile

	// Fset is the file set shared by every package syntax tree.
	Fset *token.FileSet

	// Deps maps the import path of every loaded dependency package to the
	// module that provides it.
	Deps map[string]Dependency

	// Unexported reports whether the packages of the site hold their
	// unexported symbols as well as the exported ones.
	Unexported bool
}

// Module returns the module with the given path, or nil.
func (s *Site) Module(path string) *Module {
	for _, m := range s.Modules {
		if m.Path == path {
			return m
		}
	}
	return nil
}

// Packages returns the packages of every module in site order.
func (s *Site) Packages() []*Package {
	var pkgs []*Package
	for _, m := range s.Modules {
		pkgs = append(pkgs, m.Packages...)
	}
	return pkgs
}

// Module is a loaded Go module together with its documented packages.
type Module struct {
	// Path is the module path from go.mod.
	Path string

	// Version is the module version, or empty for a main module.
	Version string

	// Dir is the root directory of the module on disk.
	Dir string

	// Packages are the packages of the module, sorted by import path.
	Packages []*Package

	// DocFile is the Markdown file that documents the module root. It is set
	// only when no package lives in the root directory.
	DocFile *DocFile
}

// Root returns the package in the module root directory, or nil.
func (m *Module) Root() *Package {
	for _, p := range m.Packages {
		if p.RelPath == "" {
			return p
		}
	}
	return nil
}

// Doc returns the raw documentation of the module: the doc comment of its
// root package, or an empty string when the module has no root package.
func (m *Module) Doc() string {
	if root := m.Root(); root != nil {
		return root.Doc
	}
	return ""
}

// Deprecated returns the deprecation message of the root package, or an
// empty string when the module has no deprecated root package.
func (m *Module) Deprecated() string {
	return Deprecation(m.Doc())
}

// DocFile is a Markdown file that documents a directory in place of a
// package doc comment.
type DocFile struct {
	// Path is the file path on disk.
	Path string

	// Text is the file contents.
	Text string
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

	// Module is the module that owns the package.
	Module *Module

	// Doc is the raw package documentation.
	Doc string

	// DocFile is the Markdown file that documents the package. It is set only
	// when the package has no doc comment.
	DocFile *DocFile

	// Consts, Vars, Types, and Funcs hold the documented identifiers: the
	// exported ones, and the unexported ones as well when the site includes
	// them. Each list holds the exported identifiers first, then the
	// unexported ones, each run sorted by name.
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

// Deprecated returns the deprecation message of the package, or an empty
// string when the package is not deprecated.
func (p *Package) Deprecated() string {
	return Deprecation(p.Doc)
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

// Type is a type declaration.
type Type struct {
	Name string
	Kind TypeKind
	Doc  string

	Decl *ast.GenDecl
	Spec *ast.TypeSpec
	Obj  *types.TypeName

	// Methods are the documented methods declared on the type, exported
	// first.
	Methods []*Func

	// Fields are the documented fields of a struct type, in declaration order.
	// It is empty for every other kind of type.
	Fields []*Field

	Examples []*Example

	Pkg *Package
}

// Field is a field of a struct type. A declaration that names
// several fields, such as "A, B int", yields one Field per name.
type Field struct {
	Name string
	Doc  string

	// Embedded reports whether the field is declared by its type alone.
	Embedded bool

	// Spec is the declaration the field belongs to.
	Spec *ast.Field
	Obj  *types.Var

	// Type is the struct that declares the field.
	Type *Type
}

// Deprecated returns the deprecation message of the field, or an empty
// string when it is not deprecated.
func (f *Field) Deprecated() string {
	return Deprecation(f.Doc)
}

// Exported reports whether the field name is exported.
func (f *Field) Exported() bool {
	return token.IsExported(f.Name)
}

// Summary returns the first paragraph of the type documentation.
func (t *Type) Summary() string {
	return FirstParagraph(t.Doc)
}

// Deprecated returns the deprecation message of the type, or an empty
// string when it is not deprecated.
func (t *Type) Deprecated() string {
	return Deprecation(t.Doc)
}

// Interface reports whether the underlying type of t is an interface. An
// alias of an interface type is an interface as well.
func (t *Type) Interface() bool {
	if t.Obj == nil {
		return t.Kind == KindInterface
	}
	_, ok := types.Unalias(t.Obj.Type()).Underlying().(*types.Interface)
	return ok
}

// Exported reports whether the type name is exported.
func (t *Type) Exported() bool {
	return token.IsExported(t.Name)
}

// Func is a function or method.
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

// Deprecated returns the deprecation message of the func, or an empty
// string when it is not deprecated.
func (f *Func) Deprecated() string {
	return Deprecation(f.Doc)
}

// Exported reports whether the function or method name is exported.
func (f *Func) Exported() bool {
	return token.IsExported(f.Name)
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

// Value is a constant or variable.
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

// SentinelError reports whether the value is a variable of type error.
func (v *Value) SentinelError() bool {
	if v.Kind != KindVar || v.Obj == nil {
		return false
	}
	return types.Identical(v.Obj.Type(), types.Universe.Lookup("error").Type())
}

// Summary returns the first paragraph of the value documentation.
func (v *Value) Summary() string {
	return FirstParagraph(v.Doc)
}

// Deprecated returns the deprecation message of the value, or an empty
// string when it is not deprecated.
func (v *Value) Deprecated() string {
	return Deprecation(v.Doc)
}

// Exported reports whether the value name is exported.
func (v *Value) Exported() bool {
	return token.IsExported(v.Name)
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

// CompareNames orders two identifiers for display: exported names before
// unexported ones, then by name. It returns a negative, zero, or positive
// value as [strings.Compare] does.
func CompareNames(lhs, rhs string) int {
	if l, r := token.IsExported(lhs), token.IsExported(rhs); l != r {
		if l {
			return -1
		}
		return 1
	}
	return strings.Compare(lhs, rhs)
}

// FirstParagraph returns the text up to the first blank line of doc.
func FirstParagraph(doc string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(doc), "\n\n")
	return first
}

// deprecatedPrefix starts the paragraph that marks a deprecation, as the Go
// tools recognise it.
const deprecatedPrefix = "Deprecated: "

// Deprecation returns the deprecation message of doc: the text of the first
// paragraph that starts with "Deprecated: ", without the prefix and with its
// lines joined by spaces. It returns an empty string when doc has no such
// paragraph.
func Deprecation(doc string) string {
	for para := range strings.SplitSeq(strings.TrimSpace(doc), "\n\n") {
		para = strings.TrimSpace(para)
		if msg, ok := strings.CutPrefix(para, deprecatedPrefix); ok {
			return strings.Join(strings.Fields(msg), " ")
		}
	}
	return ""
}
