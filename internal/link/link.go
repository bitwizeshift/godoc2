package link

import (
	"go/token"
	"go/types"

	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
)

// ExternalBase is the documentation site used for packages outside the
// site.
const ExternalBase = "https://pkg.go.dev/"

// Resolver maps objects to documentation URLs.
type Resolver struct {
	site     *model.Site
	packages map[string]*model.Package
}

// New returns a [Resolver] for the modules of site.
func New(site *model.Site) *Resolver {
	pkgs := map[string]*model.Package{}
	for _, p := range site.Packages() {
		pkgs[p.ImportPath] = p
	}
	return &Resolver{site: site, packages: pkgs}
}

// Local reports whether the package at importPath belongs to the site.
func (r *Resolver) Local(importPath string) bool {
	_, ok := r.packages[importPath]
	return ok
}

// Package returns the site package at importPath, or nil.
func (r *Resolver) Package(importPath string) *model.Package {
	return r.packages[importPath]
}

// PackageNamed returns the site package with the given package name when
// exactly one exists, and nil otherwise.
func (r *Resolver) PackageNamed(name string) *model.Package {
	var found *model.Package
	for _, p := range r.site.Packages() {
		if p.Name != name {
			continue
		}
		if found != nil {
			return nil
		}
		found = p
	}
	return found
}

// URL returns the href from the page at from to the documentation of obj.
// It reports false for struct fields, local variables, and objects that are
// not documented: unexported objects outside the site, and unexported
// objects inside the site when the site excludes them. A method is
// documented only when its receiver type is.
func (r *Resolver) URL(from string, obj types.Object) (string, bool) {
	if obj == nil {
		return "", false
	}
	if obj.Pkg() == nil {
		return ExternalBase + "builtin#" + obj.Name(), true
	}
	recv := receiverName(obj)
	if v, ok := obj.(*types.Var); ok && (v.IsField() || !isPackageLevel(obj)) {
		return "", false
	}
	if !isPackageLevel(obj) && recv == "" {
		return "", false
	}
	p, local := r.packages[obj.Pkg().Path()]
	if !r.documented(obj.Name(), local) || (recv != "" && !r.documented(recv, local)) {
		return "", false
	}
	if local {
		return r.local(from, p, recv, obj.Name()), true
	}
	return r.external(obj.Pkg().Path(), recv, obj.Name()), true
}

// documented reports whether a page exists for the identifier name: it is
// exported, or it belongs to the site and the site includes unexported
// identifiers.
func (r *Resolver) documented(name string, local bool) bool {
	return token.IsExported(name) || (local && r.site.Unexported)
}

// PackageURL returns the href from the page at from to the documentation of
// the package at importPath.
func (r *Resolver) PackageURL(from, importPath string) string {
	if p, ok := r.packages[importPath]; ok {
		return pathmap.Rel(from, pathmap.Package(p.Module.Path, p.RelPath))
	}
	return r.externalPackage(importPath)
}

// Lookup returns the href from the page at from to the object named name in
// pkg, or to its method when method is not empty. It reports false when the
// object does not exist or is not documented.
func (r *Resolver) Lookup(from string, pkg *types.Package, name, method string) (string, bool) {
	obj := pkg.Scope().Lookup(name)
	if obj == nil {
		return "", false
	}
	if method == "" {
		return r.URL(from, obj)
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return "", false
	}
	m, _, _ := types.LookupFieldOrMethod(tn.Type(), true, pkg, method)
	fn, ok := m.(*types.Func)
	if !ok {
		return "", false
	}
	return r.URL(from, fn)
}

func (r *Resolver) local(from string, p *model.Package, recv, name string) string {
	var to string
	if recv != "" {
		to = pathmap.Method(p.Module.Path, p.RelPath, recv, name)
	} else {
		to = pathmap.Symbol(p.Module.Path, p.RelPath, name)
	}
	return pathmap.Rel(from, to)
}

func (r *Resolver) external(importPath, recv, name string) string {
	fragment := name
	if recv != "" {
		fragment = recv + "." + name
	}
	return r.externalPackage(importPath) + "#" + fragment
}

func (r *Resolver) externalPackage(importPath string) string {
	url := ExternalBase + importPath
	if dep, ok := r.site.Deps[importPath]; ok && dep.Version != "" {
		url += "@" + dep.Version
	}
	return url
}

// receiverName returns the receiver type name of a method, or empty for any
// other object.
func receiverName(obj types.Object) string {
	fn, ok := obj.(*types.Func)
	if !ok {
		return ""
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return ""
	}
	t := sig.Recv().Type()
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Alias:
		return t.Obj().Name()
	default:
		return ""
	}
}

// isPackageLevel reports whether obj is declared in a package scope.
func isPackageLevel(obj types.Object) bool {
	return obj.Parent() != nil && obj.Parent() == obj.Pkg().Scope()
}
