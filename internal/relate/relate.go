package relate

import (
	"go/types"
	"maps"
	"slices"
	"strings"

	"github.com/bitwizeshift/godoc2/internal/model"
)

// Impl is one side of an implementation relation: the interface a type
// implements, or the type that implements an interface.
type Impl struct {
	// Obj is the interface or the implementing type.
	Obj *types.TypeName

	// Pointer is true when only the pointer form of the implementing type
	// satisfies the interface.
	Pointer bool

	// Local is the documented type when Obj belongs to the site, or nil.
	Local *model.Type
}

// Group is the set of [Impl] entries declared in one package.
type Group struct {
	// Path is the import path, or empty for the package of the page.
	Path string

	// Impls are the entries, sorted by name.
	Impls []Impl
}

// Index answers relation queries for the modules of one site.
type Index struct {
	packages   map[string]*model.Package
	local      map[*types.TypeName]*model.Type
	interfaces []*types.TypeName
	named      []*types.TypeName

	// unexported includes the unexported types of the site as candidates.
	unexported bool
}

// New builds the index for site.
func New(site *model.Site) *Index {
	idx := &Index{
		packages:   map[string]*model.Package{},
		local:      map[*types.TypeName]*model.Type{},
		unexported: site.Unexported,
	}
	pkgs := site.Packages()
	for _, p := range pkgs {
		idx.packages[p.ImportPath] = p
		for _, t := range p.Types {
			if t.Obj != nil {
				idx.local[t.Obj] = t
			}
		}
	}
	seen := map[*types.Package]bool{}
	for _, p := range pkgs {
		idx.scan(p.TypesPkg, seen)
	}
	return idx
}

// scan collects the candidate types of pkg and of every package it imports,
// skipping internal packages outside the site. Unexported types are
// candidates only inside a site that includes them.
func (idx *Index) scan(pkg *types.Package, seen map[*types.Package]bool) {
	if pkg == nil || seen[pkg] {
		return
	}
	seen[pkg] = true
	_, local := idx.packages[pkg.Path()]
	if !local && isInternalPath(pkg.Path()) {
		return
	}
	for _, name := range pkg.Scope().Names() {
		tn, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok || tn.IsAlias() {
			continue
		}
		if !tn.Exported() && !(local && idx.unexported) {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok || named.TypeParams().Len() > 0 {
			continue
		}
		if iface, ok := named.Underlying().(*types.Interface); ok {
			if iface.NumMethods() > 0 && iface.IsMethodSet() {
				idx.interfaces = append(idx.interfaces, tn)
			}
			continue
		}
		idx.named = append(idx.named, tn)
	}
	for _, imp := range pkg.Imports() {
		idx.scan(imp, seen)
	}
}

func isInternalPath(path string) bool {
	for elem := range strings.SplitSeq(path, "/") {
		if elem == "internal" {
			return true
		}
	}
	return false
}

// Instances returns the constants and variables of the package of t whose
// type is t or a pointer to t: the exported constants, then the exported
// variables, then the unexported ones in the same order.
func (idx *Index) Instances(t *model.Type) []*model.Value {
	if t.Obj == nil || t.Pkg == nil {
		return nil
	}
	var result []*model.Value
	for _, v := range slices.Concat(t.Pkg.Consts, t.Pkg.Vars) {
		if v.Obj != nil && sameType(v.Obj.Type(), t.Obj) {
			result = append(result, v)
		}
	}
	slices.SortStableFunc(result, func(lhs, rhs *model.Value) int {
		return compareExported(lhs.Exported(), rhs.Exported())
	})
	return result
}

// compareExported orders an exported entity before an unexported one.
func compareExported(lhs, rhs bool) int {
	switch {
	case lhs == rhs:
		return 0
	case lhs:
		return -1
	default:
		return 1
	}
}

// FuncGroup is the set of functions declared in one package.
type FuncGroup struct {
	// Path is the import path, or empty for the package of the type.
	Path string

	// Funcs are the functions, sorted by name.
	Funcs []*model.Func
}

// Constructors returns the functions of the site whose first result is t
// or a pointer to t, grouped by package. For an interface t, the functions
// whose first result is a concrete type that satisfies t are included. The
// group of the package of t comes first with an empty path, then the groups
// of the other packages of its module, then the groups of the other modules,
// each run in import path order.
func (idx *Index) Constructors(t *model.Type) []FuncGroup {
	return idx.funcGroups(t, func(f *model.Func) bool {
		return isConstructor(f, t.Obj)
	})
}

// Utilities returns the functions of the site that take t or a pointer to
// t as a parameter and are not constructors, grouped by package in the same
// order as [Index.Constructors].
func (idx *Index) Utilities(t *model.Type) []FuncGroup {
	return idx.funcGroups(t, func(f *model.Func) bool {
		return !isConstructor(f, t.Obj) && takesParam(f, t.Obj)
	})
}

// funcGroups collects the site functions accepted by match, grouped by
// package in [Index.Constructors] order.
func (idx *Index) funcGroups(t *model.Type, match func(*model.Func) bool) []FuncGroup {
	if t.Obj == nil || t.Pkg == nil {
		return nil
	}
	var groups []FuncGroup
	ranks := map[string]int{}
	for _, path := range slices.Sorted(maps.Keys(idx.packages)) {
		p := idx.packages[path]
		var funcs []*model.Func
		for _, f := range p.Funcs {
			if match(f) {
				funcs = append(funcs, f)
			}
		}
		if len(funcs) == 0 {
			continue
		}
		g := FuncGroup{Path: path, Funcs: funcs}
		if p == t.Pkg {
			g.Path = ""
		}
		ranks[g.Path] = distance(t, p)
		groups = append(groups, g)
	}
	slices.SortStableFunc(groups, func(lhs, rhs FuncGroup) int {
		return compareGroups(ranks[lhs.Path], lhs.Path, ranks[rhs.Path], rhs.Path)
	})
	return groups
}

// Distances between the package of a page type and the package of a related
// entity, in display order.
const (
	samePackage = iota
	sameModule
	otherModule
)

// distance ranks the package p against the package of t. A nil p is a
// package outside the site.
func distance(t *model.Type, p *model.Package) int {
	switch {
	case p == nil:
		return otherModule
	case p == t.Pkg:
		return samePackage
	case p.Module == t.Pkg.Module:
		return sameModule
	default:
		return otherModule
	}
}

// compareGroups orders two groups by distance, then by import path.
func compareGroups(lhsRank int, lhsPath string, rhsRank int, rhsPath string) int {
	if lhsRank != rhsRank {
		return lhsRank - rhsRank
	}
	return strings.Compare(lhsPath, rhsPath)
}

// Implements returns the non-empty interfaces that t, or a pointer to t,
// satisfies. Interfaces are not checked against other interfaces.
func (idx *Index) Implements(t *model.Type) []Impl {
	if t.Obj == nil || isInterface(t.Obj) {
		return nil
	}
	var result []Impl
	for _, iface := range idx.interfaces {
		if impl, ok := implements(t.Obj, iface); ok {
			result = append(result, Impl{Obj: iface, Pointer: impl, Local: idx.local[iface]})
		}
	}
	return result
}

// Implementations returns the concrete types that satisfy the interface t.
// It returns nil when t is not a non-empty interface.
func (idx *Index) Implementations(t *model.Type) []Impl {
	if t.Obj == nil || !isInterface(t.Obj) {
		return nil
	}
	var result []Impl
	for _, named := range idx.named {
		if impl, ok := implements(named, t.Obj); ok {
			result = append(result, Impl{Obj: named, Pointer: impl, Local: idx.local[named]})
		}
	}
	return result
}

// Groups sorts impls into groups by package. The group of the package of t
// comes first with an empty path, then the groups of the other packages of
// its module, then the groups of the other modules and of packages outside
// the site, each run in import path order.
func Groups(impls []Impl, t *model.Type) []Group {
	byPath := map[string]*Group{}
	ranks := map[string]int{}
	var result []*Group
	for _, impl := range impls {
		path := impl.Obj.Pkg().Path()
		if t.Obj != nil && impl.Obj.Pkg() == t.Obj.Pkg() {
			path = ""
		}
		g, ok := byPath[path]
		if !ok {
			g = &Group{Path: path}
			byPath[path] = g
			ranks[path] = distance(t, localPackage(impl))
			result = append(result, g)
		}
		g.Impls = append(g.Impls, impl)
	}
	for _, g := range result {
		slices.SortFunc(g.Impls, func(lhs, rhs Impl) int {
			return model.CompareNames(lhs.Obj.Name(), rhs.Obj.Name())
		})
	}
	slices.SortFunc(result, func(lhs, rhs *Group) int {
		return compareGroups(ranks[lhs.Path], lhs.Path, ranks[rhs.Path], rhs.Path)
	})
	groups := make([]Group, 0, len(result))
	for _, g := range result {
		groups = append(groups, *g)
	}
	return groups
}

// localPackage returns the site package that declares impl, or nil.
func localPackage(impl Impl) *model.Package {
	if impl.Local == nil {
		return nil
	}
	return impl.Local.Pkg
}

// implements reports whether typ or *typ satisfies iface, and whether the
// pointer form was needed.
func implements(typ, iface *types.TypeName) (pointer, ok bool) {
	target, isIface := types.Unalias(iface.Type()).Underlying().(*types.Interface)
	if !isIface || typ == iface || target.NumMethods() == 0 {
		return false, false
	}
	t := typ.Type()
	if types.Implements(t, target) {
		return false, true
	}
	if types.Implements(types.NewPointer(t), target) {
		return true, true
	}
	return false, false
}

func isInterface(tn *types.TypeName) bool {
	_, ok := types.Unalias(tn.Type()).Underlying().(*types.Interface)
	return ok
}

// isConstructor reports whether the first result of f is target, a pointer
// to it, or a concrete type that satisfies the interface target.
func isConstructor(f *model.Func, target *types.TypeName) bool {
	sig, ok := f.Obj.Type().(*types.Signature)
	if !ok || sig.Results().Len() == 0 {
		return false
	}
	result := sig.Results().At(0).Type()
	return sameType(result, target) || satisfies(result, target)
}

// satisfies reports whether the concrete type typ implements the non-empty
// interface target. typ is checked as written: a value type does not satisfy
// an interface that only its pointer type implements.
func satisfies(typ types.Type, target *types.TypeName) bool {
	iface, ok := types.Unalias(target.Type()).Underlying().(*types.Interface)
	if !ok || iface.NumMethods() == 0 || types.IsInterface(typ) {
		return false
	}
	return types.Implements(typ, iface)
}

// takesParam reports whether any parameter of f is target or a pointer to
// it.
func takesParam(f *model.Func, target *types.TypeName) bool {
	sig, ok := f.Obj.Type().(*types.Signature)
	if !ok {
		return false
	}
	for p := range sig.Params().Variables() {
		if sameType(p.Type(), target) {
			return true
		}
	}
	return false
}

// sameType reports whether typ is target, a pointer to target, or an
// instantiation of a generic target.
func sameType(typ types.Type, target *types.TypeName) bool {
	if p, ok := typ.(*types.Pointer); ok {
		typ = p.Elem()
	}
	typ = types.Unalias(typ)
	want := types.Unalias(target.Type())
	if types.Identical(typ, want) {
		return true
	}
	got, ok := typ.(*types.Named)
	wantNamed, wantOK := want.(*types.Named)
	return ok && wantOK && got.Origin() == wantNamed.Origin()
}
