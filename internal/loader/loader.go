package loader

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/model"
)

var (
	// ErrNoModule is returned when no matched package belongs to a module.
	ErrNoModule = errors.New("no module")

	// ErrLoad is returned when a package reports load or type errors.
	ErrLoad = errors.New("load")
)

// Config configures a [Load] call.
type Config struct {
	// Dir is the working directory for pattern resolution. Empty means the
	// current directory.
	Dir string

	// Patterns are the package patterns, such as "./...".
	Patterns []string

	// Workspace is the path of a go.work file, relative to Dir. When it is
	// set, the go command runs in that workspace and every module the file
	// uses is added to the patterns as "<dir>/...".
	Workspace string

	// Unexported includes the unexported identifiers of every package in the
	// site, after the exported ones. The init and main functions are never
	// included.
	Unexported bool
}

// Load calls [Load] with the configuration.
func (c Config) Load(ctx context.Context) (*model.Site, error) {
	return Load(ctx, c)
}

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedImports |
	packages.NeedModule

// Load loads the packages matched by cfg and returns the site of the modules
// that own them. Every module with at least one matched package is part of
// the site, so a pattern that names a dependency documents that dependency.
//
// It returns [ErrNoModule] if no matched package belongs to a module, and an
// error wrapping [ErrLoad] if the workspace file cannot be read, a pattern
// did not match, or a package failed to load or type-check.
func Load(ctx context.Context, cfg Config) (*model.Site, error) {
	dir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoad, err)
	}
	patterns, env, err := workspace(dir, cfg)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Context: ctx,
		Mode:    loadMode,
		Dir:     cfg.Dir,
		Env:     env,
		Fset:    fset,
	}, patterns...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoad, err)
	}
	if err := packageErrors(pkgs); err != nil {
		return nil, err
	}

	site := &model.Site{Dir: dir, Fset: fset, Deps: dependencies(pkgs), Unexported: cfg.Unexported}
	for _, pkg := range pkgs {
		if pkg.Module == nil {
			continue
		}
		mod := site.Module(pkg.Module.Path)
		if mod == nil {
			mod = &model.Module{Path: pkg.Module.Path, Version: pkg.Module.Version, Dir: pkg.Module.Dir}
			site.Modules = append(site.Modules, mod)
		}
		p, err := newPackage(fset, mod, pkg, cfg.Unexported)
		if err != nil {
			return nil, err
		}
		mod.Packages = append(mod.Packages, p)
	}
	if len(site.Modules) == 0 {
		return nil, ErrNoModule
	}
	slices.SortFunc(site.Modules, func(lhs, rhs *model.Module) int {
		return strings.Compare(lhs.Path, rhs.Path)
	})
	for _, mod := range site.Modules {
		slices.SortFunc(mod.Packages, func(lhs, rhs *model.Package) int {
			return strings.Compare(lhs.ImportPath, rhs.ImportPath)
		})
		if err := moduleDocFile(mod); err != nil {
			return nil, err
		}
	}
	if err := siteDocFile(site); err != nil {
		return nil, err
	}
	return site, nil
}

// moduleDocFile attaches the Markdown file of the module root when no
// package lives there.
func moduleDocFile(mod *model.Module) error {
	if mod.Root() != nil {
		return nil
	}
	f, err := docfile.Find(mod.Dir)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrLoad, mod.Path, err)
	}
	mod.DocFile = f
	return nil
}

// siteDocFile attaches the Markdown file of the site directory when the site
// holds more than one module.
func siteDocFile(site *model.Site) error {
	if len(site.Modules) < 2 {
		return nil
	}
	f, err := docfile.Find(site.Dir)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrLoad, site.Dir, err)
	}
	site.DocFile = f
	return nil
}

// packageErrors returns an error wrapping [ErrLoad] for the first package
// that reports errors.
func packageErrors(pkgs []*packages.Package) error {
	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			continue
		}
		msgs := make([]string, 0, len(pkg.Errors))
		for _, e := range pkg.Errors {
			msgs = append(msgs, e.Error())
		}
		return fmt.Errorf("%w: %s: %s", ErrLoad, pkg.PkgPath, strings.Join(msgs, "; "))
	}
	return nil
}

// dependencies walks the import graph and records the module of every
// imported package.
func dependencies(pkgs []*packages.Package) map[string]model.Dependency {
	deps := map[string]model.Dependency{}
	seen := map[string]bool{}
	var visit func(*packages.Package)
	visit = func(pkg *packages.Package) {
		if seen[pkg.PkgPath] {
			return
		}
		seen[pkg.PkgPath] = true
		if pkg.Module != nil && !pkg.Module.Main {
			deps[pkg.PkgPath] = model.Dependency{
				Path:    pkg.Module.Path,
				Version: pkg.Module.Version,
			}
		}
		for _, imp := range pkg.Imports {
			visit(imp)
		}
	}
	for _, pkg := range pkgs {
		visit(pkg)
	}
	return deps
}

// newPackage builds the model of one loaded package. unexported includes
// the unexported identifiers of the package.
func newPackage(fset *token.FileSet, mod *model.Module, pkg *packages.Package, unexported bool) (*model.Package, error) {
	dir := packageDir(pkg)
	testFiles, err := parseTestFiles(fset, dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrLoad, pkg.PkgPath, err)
	}
	files := slices.Concat(pkg.Syntax, testFiles)
	var mode doc.Mode
	if unexported {
		mode = doc.AllDecls
	}
	dpkg, err := doc.NewFromFiles(fset, files, pkg.PkgPath, mode)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrLoad, pkg.PkgPath, err)
	}

	p := &model.Package{
		ImportPath: pkg.PkgPath,
		Name:       pkg.Name,
		RelPath:    strings.TrimPrefix(strings.TrimPrefix(pkg.PkgPath, mod.Path), "/"),
		Dir:        dir,
		Module:     mod,
		Doc:        dpkg.Doc,
		TypesPkg:   pkg.Types,
		Info:       pkg.TypesInfo,
		Files:      sourceFiles(pkg),
	}
	if strings.TrimSpace(p.Doc) == "" {
		p.DocFile, err = docfile.Find(dir)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrLoad, pkg.PkgPath, err)
		}
	}
	c := &converter{fset: fset, pkg: p, scope: pkg.Types.Scope(), unexported: unexported}
	p.Consts = c.values(dpkg.Consts, model.KindConst)
	p.Vars = c.values(dpkg.Vars, model.KindVar)
	p.Examples = c.examples(dpkg.Examples)
	for _, t := range dpkg.Types {
		p.Consts = append(p.Consts, c.values(t.Consts, model.KindConst)...)
		p.Vars = append(p.Vars, c.values(t.Vars, model.KindVar)...)
		p.Types = append(p.Types, c.typ(t))
		p.Funcs = append(p.Funcs, c.funcs(t.Funcs, nil)...)
	}
	p.Funcs = append(p.Funcs, c.funcs(dpkg.Funcs, nil)...)
	sortByName(p.Consts, func(v *model.Value) string { return v.Name })
	sortByName(p.Vars, func(v *model.Value) string { return v.Name })
	sortByName(p.Types, func(t *model.Type) string { return t.Name })
	sortByName(p.Funcs, func(f *model.Func) string { return f.Name })
	return p, nil
}

// packageDir returns the directory that holds the package files.
func packageDir(pkg *packages.Package) string {
	if pkg.Dir != "" {
		return pkg.Dir
	}
	if len(pkg.GoFiles) > 0 {
		return filepath.Dir(pkg.GoFiles[0])
	}
	return ""
}

// sourceFiles lists the Go files of the package, including test files,
// sorted by name.
func sourceFiles(pkg *packages.Package) []*model.File {
	var files []*model.File
	for _, path := range pkg.GoFiles {
		files = append(files, &model.File{Name: filepath.Base(path), Path: path})
	}
	dir := packageDir(pkg)
	for _, name := range testFileNames(dir) {
		files = append(files, &model.File{Name: name, Path: filepath.Join(dir, name)})
	}
	sortByName(files, func(f *model.File) string { return f.Name })
	return files
}

// testFileNames lists the _test.go files in dir. A directory that cannot be
// read yields no names.
func testFileNames(dir string) []string {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
			names = append(names, e.Name())
		}
	}
	return names
}

// parseTestFiles parses the _test.go files in dir with comments, so that
// examples can be extracted from them.
func parseTestFiles(fset *token.FileSet, dir string) ([]*ast.File, error) {
	var files []*ast.File
	for _, name := range testFileNames(dir) {
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// sortByName orders items by name, with the exported names before the
// unexported ones.
func sortByName[T any](items []T, name func(T) string) {
	slices.SortStableFunc(items, func(lhs, rhs T) int {
		return model.CompareNames(name(lhs), name(rhs))
	})
}

// converter turns [go/doc] entities into model entities for one package.
type converter struct {
	fset  *token.FileSet
	pkg   *model.Package
	scope *types.Scope

	// unexported includes the unexported identifiers.
	unexported bool
}

// documented reports whether the identifier name is part of the model: it
// is exported, or the converter includes unexported identifiers. The blank
// identifier is never documented.
func (c *converter) documented(name string) bool {
	if name == "_" {
		return false
	}
	return c.unexported || token.IsExported(name)
}

func (c *converter) values(vals []*doc.Value, kind model.ValueKind) []*model.Value {
	var result []*model.Value
	for _, v := range vals {
		for _, spec := range v.Decl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if !c.documented(name.Name) {
					continue
				}
				result = append(result, &model.Value{
					Name: name.Name,
					Kind: kind,
					Doc:  valueDoc(v, vs),
					Decl: v.Decl,
					Spec: vs,
					Obj:  c.scope.Lookup(name.Name),
					Pkg:  c.pkg,
				})
			}
		}
	}
	return result
}

// valueDoc returns the documentation of one spec in a value group, falling
// back to the group documentation.
func valueDoc(v *doc.Value, spec *ast.ValueSpec) string {
	if spec.Doc != nil {
		return spec.Doc.Text()
	}
	return v.Doc
}

func (c *converter) typ(t *doc.Type) *model.Type {
	spec := typeSpec(t)
	obj, _ := c.scope.Lookup(t.Name).(*types.TypeName)
	mt := &model.Type{
		Name:     t.Name,
		Kind:     typeKind(spec, obj),
		Doc:      t.Doc,
		Decl:     t.Decl,
		Spec:     spec,
		Obj:      obj,
		Examples: c.examples(t.Examples),
		Pkg:      c.pkg,
	}
	mt.Methods = c.funcs(t.Methods, mt)
	sortByName(mt.Methods, func(f *model.Func) string { return f.Name })
	mt.Fields = c.fields(mt)
	return mt
}

// fields lists the documented fields of a struct type in declaration order.
func (c *converter) fields(t *model.Type) []*model.Field {
	if t.Spec == nil {
		return nil
	}
	st, ok := t.Spec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	var result []*model.Field
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			if id := embeddedIdent(f.Type); id != nil && c.documented(id.Name) {
				result = append(result, c.field(t, f, id, true))
			}
			continue
		}
		for _, name := range f.Names {
			if c.documented(name.Name) {
				result = append(result, c.field(t, f, name, false))
			}
		}
	}
	return result
}

func (c *converter) field(t *model.Type, f *ast.Field, name *ast.Ident, embedded bool) *model.Field {
	var obj *types.Var
	if c.pkg.Info != nil {
		obj, _ = c.pkg.Info.Defs[name].(*types.Var)
	}
	return &model.Field{
		Name:     name.Name,
		Doc:      fieldDoc(f),
		Embedded: embedded,
		Spec:     f,
		Obj:      obj,
		Type:     t,
	}
}

// fieldDoc returns the doc comment of a field, or else its line comment.
func fieldDoc(f *ast.Field) string {
	if f.Doc != nil {
		return f.Doc.Text()
	}
	if f.Comment != nil {
		return f.Comment.Text()
	}
	return ""
}

// embeddedIdent returns the identifier that names the type of an embedded
// field: T, *T, pkg.T, or an instantiation of one of these. It returns nil
// for any other expression.
func embeddedIdent(e ast.Expr) *ast.Ident {
	switch e := e.(type) {
	case *ast.Ident:
		return e
	case *ast.StarExpr:
		return embeddedIdent(e.X)
	case *ast.SelectorExpr:
		return e.Sel
	case *ast.IndexExpr:
		return embeddedIdent(e.X)
	case *ast.IndexListExpr:
		return embeddedIdent(e.X)
	default:
		return nil
	}
}

// typeSpec returns the spec of the named type inside its declaration group.
func typeSpec(t *doc.Type) *ast.TypeSpec {
	for _, spec := range t.Decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == t.Name {
			return ts
		}
	}
	return nil
}

func typeKind(spec *ast.TypeSpec, obj *types.TypeName) model.TypeKind {
	if spec != nil && spec.Assign.IsValid() {
		return model.KindAlias
	}
	if obj == nil {
		return model.KindOther
	}
	switch obj.Type().Underlying().(type) {
	case *types.Struct:
		return model.KindStruct
	case *types.Interface:
		return model.KindInterface
	default:
		return model.KindOther
	}
}

// funcs converts the functions or methods of one [go/doc] list. The init and
// main functions are never converted: neither can be named by other code.
func (c *converter) funcs(fns []*doc.Func, recv *model.Type) []*model.Func {
	var result []*model.Func
	for _, f := range fns {
		if f.Decl == nil || !c.documented(f.Name) {
			continue
		}
		if recv == nil && (f.Name == "init" || f.Name == "main") {
			continue
		}
		obj := c.funcObject(f, recv)
		if obj == nil {
			continue
		}
		result = append(result, &model.Func{
			Name:     f.Name,
			Doc:      f.Doc,
			Decl:     f.Decl,
			Obj:      obj,
			Recv:     recv,
			Examples: c.examples(f.Examples),
			Pkg:      c.pkg,
		})
	}
	return result
}

// funcObject finds the type-checker object of a function or method.
func (c *converter) funcObject(f *doc.Func, recv *model.Type) *types.Func {
	if recv == nil {
		obj, _ := c.scope.Lookup(f.Name).(*types.Func)
		return obj
	}
	if recv.Obj == nil {
		return nil
	}
	named, ok := recv.Obj.Type().(*types.Named)
	if !ok {
		return nil
	}
	for m := range named.Methods() {
		if m.Name() == f.Name {
			return m
		}
	}
	return nil
}

func (c *converter) examples(exs []*doc.Example) []*model.Example {
	var result []*model.Example
	for _, ex := range exs {
		result = append(result, &model.Example{
			Name:   ex.Name,
			Suffix: ex.Suffix,
			Doc:    ex.Doc,
			Code:   formatExample(c.fset, ex),
			Output: ex.Output,
		})
	}
	return result
}
