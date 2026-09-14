package render

import (
	"go/ast"

	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/relate"
)

// typeItem returns the list entry of a type.
func (b *builder) typeItem(t *model.Type) item {
	href := b.rel(pathmap.Symbol(b.r.module.Path, t.Pkg.RelPath, t.Name))
	summary, full := b.summaryAndFull(t.Doc)
	return item{
		ID:         "type." + t.Name,
		Name:       t.Name,
		Href:       href,
		Code:       b.code(b.printer(href).Type(t)),
		SourceHref: b.sourceHref(t.Spec),
		Summary:    summary,
		Full:       full,
	}
}

// funcItem returns the list entry of a function or method. The id prefix
// distinguishes the section the entry appears in.
func (b *builder) funcItem(prefix string, f *model.Func) item {
	href := b.rel(funcPath(b.r.module.Path, f))
	summary, full := b.summaryAndFull(f.Doc)
	return item{
		ID:         prefix + "." + f.Name,
		Name:       f.Name,
		Href:       href,
		Code:       b.code(b.printer(href).Func(f)),
		SourceHref: b.sourceHref(f.Decl),
		Summary:    summary,
		Full:       full,
	}
}

// valueItem returns the list entry of a constant or variable.
func (b *builder) valueItem(prefix string, v *model.Value) item {
	href := b.rel(pathmap.Symbol(b.r.module.Path, v.Pkg.RelPath, v.Name))
	summary, full := b.summaryAndFull(v.Doc)
	var pos ast.Node
	if v.Spec != nil {
		pos = v.Spec
	}
	return item{
		ID:         prefix + "." + v.Name,
		Name:       v.Name,
		Href:       href,
		Code:       b.code(b.printer(href).Value(v)),
		SourceHref: b.sourceHref(pos),
		Summary:    summary,
		Full:       full,
	}
}

func (b *builder) typeItems(types []*model.Type) []item {
	var items []item
	for _, t := range types {
		items = append(items, b.typeItem(t))
	}
	return items
}

func (b *builder) funcItems(prefix string, fns []*model.Func) []item {
	var items []item
	for _, f := range fns {
		items = append(items, b.funcItem(prefix, f))
	}
	return items
}

func (b *builder) valueItems(prefix string, vals []*model.Value) []item {
	var items []item
	for _, v := range vals {
		items = append(items, b.valueItem(prefix, v))
	}
	return items
}

// funcGroupsSection lists functions grouped by package, with each entry
// rendered in the scope of its own package.
func (b *builder) funcGroupsSection(id, title, prefix string, groups []relate.FuncGroup) *section {
	if len(groups) == 0 {
		return nil
	}
	s := &section{ID: id, Title: title, Kind: kindRelations}
	for _, g := range groups {
		rg := relationGroup{Path: g.Path}
		for _, f := range g.Funcs {
			rg.Items = append(rg.Items, b.withPackage(f.Pkg).funcItem(prefix, f))
		}
		s.Groups = append(s.Groups, rg)
	}
	return s
}

// relationSection lists implementation relations grouped by package. For an
// Implements section the implementing type is the page type t; for an
// Implementations section it is the entry itself.
func (b *builder) relationSection(id, title string, groups []relate.Group, t *model.Type) *section {
	if len(groups) == 0 {
		return nil
	}
	s := &section{ID: id, Title: title, Kind: kindRelations}
	for _, g := range groups {
		rg := relationGroup{Path: g.Path}
		for _, impl := range g.Impls {
			receiver := t.Name
			if id == "implementations" {
				receiver = impl.Obj.Name()
			}
			if impl.Pointer {
				receiver = "*" + receiver
			}
			rg.Items = append(rg.Items, b.implItem(id, impl, receiver))
		}
		s.Groups = append(s.Groups, rg)
	}
	return s
}

// implItem returns the entry of one implementation relation: the declaration
// of the related interface or type, and the form of the implementing type
// (T or *T) as the receiver badge. Types inside the module render from their
// syntax with a source link; other types render from type information.
func (b *builder) implItem(prefix string, impl relate.Impl, receiver string) item {
	href, _ := b.r.resolver.URL(b.from, impl.Obj)
	it := item{
		ID:       prefix + "." + impl.Obj.Pkg().Path() + "." + impl.Obj.Name(),
		Name:     impl.Obj.Name(),
		Href:     href,
		Receiver: receiver,
	}
	if impl.Local == nil {
		it.Code = b.code(b.printer(href).Object(impl.Obj))
		return it
	}
	local := b.withPackage(impl.Local.Pkg)
	it.Code = local.code(local.printer(href).Type(impl.Local))
	it.SourceHref = local.sourceHref(impl.Local.Spec)
	it.Summary, it.Full = local.summaryAndFull(impl.Local.Doc)
	it.Internal = impl.Local.Pkg.Internal()
	return it
}
