package sig

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"

	"github.com/bitwizeshift/godoc2/internal/model"
)

// valueSpec writes one constant or variable as a single declaration line.
func (w *writer) valueSpec(v *model.Value) {
	w.str(v.Kind.String())
	w.str(" ")
	w.linked(v.Name, w.printer.Self)
	if t := w.valueType(v); t != "" {
		w.str(" ")
		w.str(t)
	}
	w.valueInit(v)
}

// valueType returns the declared type expression of the value when the spec
// names one, and the inferred type otherwise. Untyped constants yield an
// empty string.
func (w *writer) valueType(v *model.Value) string {
	if v.Spec != nil && v.Spec.Type != nil {
		probe := &writer{printer: w.printer}
		probe.expr(v.Spec.Type)
		w.links = append(w.links, shift(probe.links, w.buf.Len()+1)...)
		return probe.buf.String()
	}
	if v.Obj == nil {
		return ""
	}
	t := v.Obj.Type()
	if b, ok := t.(*types.Basic); ok && b.Info()&types.IsUntyped != 0 {
		return ""
	}
	text := types.TypeString(t, w.qualifier)
	if url := w.typeURL(t); url != "" {
		start := w.buf.Len() + 1
		w.links = append(w.links, Span{Start: start, End: start + len(text), URL: url})
	}
	return text
}

// typeURL resolves the documentation of a named type, a pointer to one, or
// a basic type. Other types yield an empty string.
func (w *writer) typeURL(t types.Type) string {
	p := w.printer
	if p.Resolver == nil {
		return ""
	}
	var obj types.Object
	switch t := t.(type) {
	case *types.Named:
		obj = t.Obj()
	case *types.Alias:
		obj = t.Obj()
	case *types.Basic:
		obj = types.Universe.Lookup(t.Name())
	default:
		return ""
	}
	url, ok := p.Resolver.URL(p.From, obj)
	if !ok {
		return ""
	}
	return url
}

// qualifier prints package-qualified names with the package name, and omits
// the qualifier for the current package.
func (w *writer) qualifier(pkg *types.Package) string {
	if pkg == nil || (w.printer.Pkg != nil && pkg == w.printer.Pkg.TypesPkg) {
		return ""
	}
	return pkg.Name()
}

// valueInit writes the initializer. Constants that are defined with iota, or
// that repeat an earlier expression, show their evaluated value.
func (w *writer) valueInit(v *model.Value) {
	if e := w.initExpr(v); e != nil && !usesIota(e) {
		w.str(" = ")
		w.expr(e)
		return
	}
	c, ok := v.Obj.(*types.Const)
	if !ok {
		return
	}
	w.str(" = ")
	w.str(constString(c.Val()))
}

// initExpr returns the expression that initializes the value in its spec, or
// nil when the spec has none.
func (w *writer) initExpr(v *model.Value) ast.Expr {
	if v.Spec == nil {
		return nil
	}
	for i, name := range v.Spec.Names {
		if name.Name == v.Name && i < len(v.Spec.Values) {
			return v.Spec.Values[i]
		}
	}
	return nil
}

// usesIota reports whether the expression mentions iota.
func usesIota(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "iota" {
			found = true
		}
		return !found
	})
	return found
}

// constString renders a constant value as Go source.
func constString(v constant.Value) string {
	switch v.Kind() {
	case constant.String:
		return strconv.Quote(constant.StringVal(v))
	case constant.Int:
		return v.ExactString()
	default:
		return v.String()
	}
}

func shift(links []Span, by int) []Span {
	result := make([]Span, len(links))
	for i, l := range links {
		result[i] = Span{Start: l.Start + by, End: l.End + by, URL: l.URL}
	}
	return result
}
