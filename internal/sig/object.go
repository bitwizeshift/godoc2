package sig

import (
	"go/types"
)

// Object renders the declaration of a type from its type-checker object,
// for types whose syntax is not available, such as types from other modules.
// Struct fields and interface methods are printed one per line with the same
// cleanup as [Printer.Type].
func (p *Printer) Object(obj *types.TypeName) Rendered {
	w := p.writer()
	w.str("type ")
	w.linked(obj.Name(), p.Self)
	named, _ := obj.Type().(*types.Named)
	if named != nil {
		w.typeParamList(named.TypeParams())
	}
	if obj.IsAlias() {
		w.str(" = ")
		w.typ(types.Unalias(obj.Type()))
		return w.rendered()
	}
	w.str(" ")
	switch u := obj.Type().Underlying().(type) {
	case *types.Struct:
		w.structObj(u)
	case *types.Interface:
		w.interfaceObj(u)
	default:
		w.typ(u)
	}
	return w.rendered()
}

// typ writes a type expression from a type-checker type, linking named and
// basic types.
func (w *writer) typ(t types.Type) {
	switch t := t.(type) {
	case *types.Basic:
		w.linked(t.Name(), w.typeURL(t))
	case *types.Named:
		w.namedType(t)
	case *types.Alias:
		w.namedAlias(t)
	case *types.Pointer:
		w.str("*")
		w.typ(t.Elem())
	case *types.Slice:
		w.str("[]")
		w.typ(t.Elem())
	case *types.Array:
		w.str("[" + itoa(t.Len()) + "]")
		w.typ(t.Elem())
	case *types.Map:
		w.str("map[")
		w.typ(t.Key())
		w.str("]")
		w.typ(t.Elem())
	case *types.Chan:
		w.chanObj(t)
	case *types.Signature:
		w.str("func")
		w.signature(t)
	case *types.Interface:
		w.inlineInterface(t)
	case *types.Struct:
		w.inlineStruct(t)
	case *types.TypeParam:
		w.str(t.Obj().Name())
	case *types.Tuple:
		w.tuple(t, true)
	default:
		w.str(types.TypeString(t, w.qualifier))
	}
}

// namedType writes a qualified type name with its type arguments.
func (w *writer) namedType(t *types.Named) {
	w.qualified(t.Obj())
	if args := t.TypeArgs(); args.Len() > 0 {
		w.str("[")
		for i := range args.Len() {
			if i > 0 {
				w.str(", ")
			}
			w.typ(args.At(i))
		}
		w.str("]")
	}
}

func (w *writer) namedAlias(t *types.Alias) {
	w.qualified(t.Obj())
}

// qualified writes a type name, prefixed with its package name when it is
// declared outside the current package, and links it.
func (w *writer) qualified(obj *types.TypeName) {
	text := obj.Name()
	if q := w.qualifier(obj.Pkg()); q != "" {
		text = q + "." + text
	}
	url := ""
	if w.printer.Resolver != nil {
		url, _ = w.printer.Resolver.URL(w.printer.From, obj)
	}
	w.linked(text, url)
}

func (w *writer) chanObj(t *types.Chan) {
	switch t.Dir() {
	case types.RecvOnly:
		w.str("<-chan ")
	case types.SendOnly:
		w.str("chan<- ")
	default:
		w.str("chan ")
	}
	w.typ(t.Elem())
}

// signature writes the parameters and results of a function type.
func (w *writer) signature(sig *types.Signature) {
	w.str("(")
	w.params(sig)
	w.str(")")
	results := sig.Results()
	if results.Len() == 0 {
		return
	}
	w.str(" ")
	if results.Len() == 1 && results.At(0).Name() == "" {
		w.typ(results.At(0).Type())
		return
	}
	w.tuple(results, true)
}

// params writes the parameter list, marking a variadic final parameter.
func (w *writer) params(sig *types.Signature) {
	p := sig.Params()
	for i := range p.Len() {
		if i > 0 {
			w.str(", ")
		}
		v := p.At(i)
		if v.Name() != "" {
			w.str(v.Name() + " ")
		}
		if sig.Variadic() && i == p.Len()-1 {
			if s, ok := v.Type().(*types.Slice); ok {
				w.str("...")
				w.typ(s.Elem())
				continue
			}
		}
		w.typ(v.Type())
	}
}

func (w *writer) tuple(t *types.Tuple, parens bool) {
	if parens {
		w.str("(")
	}
	for i := range t.Len() {
		if i > 0 {
			w.str(", ")
		}
		v := t.At(i)
		if v.Name() != "" {
			w.str(v.Name() + " ")
		}
		w.typ(v.Type())
	}
	if parens {
		w.str(")")
	}
}

func (w *writer) typeParamList(tps *types.TypeParamList) {
	if tps == nil || tps.Len() == 0 {
		return
	}
	w.str("[")
	for i := range tps.Len() {
		if i > 0 {
			w.str(", ")
		}
		tp := tps.At(i)
		w.str(tp.Obj().Name() + " ")
		w.typ(tp.Constraint())
	}
	w.str("]")
}

// structObj writes the exported fields of a struct type, one per line.
func (w *writer) structObj(t *types.Struct) {
	w.str("struct {")
	var members []func()
	hidden := false
	for f := range t.Fields() {
		if !f.Exported() {
			hidden = true
			continue
		}
		members = append(members, func() { w.fieldObj(f, t) })
	}
	w.block(members, hidden, "// contains unexported fields")
}

func (w *writer) fieldObj(f *types.Var, t *types.Struct) {
	if !f.Embedded() {
		w.str(f.Name() + " ")
	}
	w.typ(f.Type())
	for i := range t.NumFields() {
		if t.Field(i) == f && t.Tag(i) != "" {
			w.str(" `" + t.Tag(i) + "`")
		}
	}
}

// interfaceObj writes the exported methods and embedded interfaces of an
// interface type, one per line.
func (w *writer) interfaceObj(t *types.Interface) {
	w.str("interface {")
	var members []func()
	hidden := false
	for i := range t.NumEmbeddeds() {
		e := t.EmbeddedType(i)
		members = append(members, func() { w.typ(e) })
	}
	for m := range t.ExplicitMethods() {
		if !m.Exported() {
			hidden = true
			continue
		}
		members = append(members, func() { w.methodObj(m) })
	}
	w.block(members, hidden, "// contains unexported methods")
}

func (w *writer) methodObj(m *types.Func) {
	w.str(m.Name())
	if sig, ok := m.Type().(*types.Signature); ok {
		w.signature(sig)
	}
}

// inlineInterface writes an interface literal on one line.
func (w *writer) inlineInterface(t *types.Interface) {
	if t.NumMethods() == 0 && t.NumEmbeddeds() == 0 {
		w.str("interface{}")
		return
	}
	w.str("interface{ ")
	first := true
	for i := range t.NumEmbeddeds() {
		if !first {
			w.str("; ")
		}
		first = false
		w.typ(t.EmbeddedType(i))
	}
	for m := range t.ExplicitMethods() {
		if !first {
			w.str("; ")
		}
		first = false
		w.methodObj(m)
	}
	w.str(" }")
}

// inlineStruct writes a struct literal on one line.
func (w *writer) inlineStruct(t *types.Struct) {
	if t.NumFields() == 0 {
		w.str("struct{}")
		return
	}
	w.str("struct{ ")
	for i, f := range enumerate(t.Fields()) {
		if i > 0 {
			w.str("; ")
		}
		w.fieldObj(f, t)
	}
	w.str(" }")
}
