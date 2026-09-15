package sig

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode/utf8"

	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/model"
)

// MaxWidth is the widest one-line signature, in runes, before parameters are
// broken onto separate lines.
const MaxWidth = 80

// Span marks a byte range of rendered text that links to a URL.
type Span struct {
	Start int
	End   int
	URL   string
}

// Rendered is the text of a declaration and its link spans, in order of
// appearance.
type Rendered struct {
	Text  string
	Links []Span
}

// Printer renders declarations of one package for one output page.
type Printer struct {
	// Resolver maps identifiers to URLs.
	Resolver *link.Resolver

	// Pkg is the package that declares the printed entities.
	Pkg *model.Package

	// From is the output path of the page the text is rendered on.
	From string

	// Self is the URL the declared name links to, or empty for no link.
	Self string

	// Unexported writes the unexported fields and methods of struct and
	// interface declarations in place of the comment that stands for them.
	Unexported bool
}

// Func renders a function or method signature.
func (p *Printer) Func(f *model.Func) Rendered {
	w := p.writer()
	w.funcDecl(f.Decl, false)
	if w.width() > MaxWidth {
		w = p.writer()
		w.funcDecl(f.Decl, true)
	}
	return w.rendered()
}

// Type renders a type declaration.
func (p *Printer) Type(t *model.Type) Rendered {
	w := p.writer()
	w.typeSpec(t.Spec)
	return w.rendered()
}

// Value renders a single constant or variable declaration.
func (p *Printer) Value(v *model.Value) Rendered {
	w := p.writer()
	w.valueSpec(v)
	return w.rendered()
}

// Field renders a single struct field: its name, type, and tag. An embedded
// field renders as its type alone.
func (p *Printer) Field(f *model.Field) Rendered {
	w := p.writer()
	w.fieldDecl(f)
	return w.rendered()
}

func (p *Printer) writer() *writer {
	return &writer{printer: p}
}

// writer accumulates rendered text and link spans.
type writer struct {
	printer *Printer
	buf     strings.Builder
	links   []Span
	indent  int
}

func (w *writer) rendered() Rendered {
	return Rendered{Text: w.buf.String(), Links: w.links}
}

func (w *writer) width() int {
	return utf8.RuneCountInString(w.buf.String())
}

func (w *writer) str(s string) {
	w.buf.WriteString(s)
}

func (w *writer) newline() {
	w.buf.WriteByte('\n')
	for range w.indent {
		w.buf.WriteByte('\t')
	}
}

// blank writes an empty line without indentation.
func (w *writer) blank() {
	w.buf.WriteByte('\n')
}

// linked writes text and records a span for it when url is not empty.
func (w *writer) linked(text, url string) {
	start := w.buf.Len()
	w.str(text)
	if url != "" {
		w.links = append(w.links, Span{Start: start, End: w.buf.Len(), URL: url})
	}
}

// urlOf resolves the object an identifier refers to.
func (w *writer) urlOf(id *ast.Ident) string {
	p := w.printer
	if p.Pkg == nil || p.Pkg.Info == nil || p.Resolver == nil {
		return ""
	}
	obj := p.Pkg.Info.Uses[id]
	if obj == nil {
		return ""
	}
	url, ok := p.Resolver.URL(p.From, obj)
	if !ok {
		return ""
	}
	return url
}

// expr writes a type or value expression.
func (w *writer) expr(e ast.Expr) {
	switch e := e.(type) {
	case *ast.Ident:
		w.linked(e.Name, w.urlOf(e))
	case *ast.SelectorExpr:
		w.selector(e)
	case *ast.StarExpr:
		w.str("*")
		w.expr(e.X)
	case *ast.ParenExpr:
		w.str("(")
		w.expr(e.X)
		w.str(")")
	case *ast.ArrayType:
		w.str("[")
		if e.Len != nil {
			w.expr(e.Len)
		}
		w.str("]")
		w.expr(e.Elt)
	case *ast.Ellipsis:
		w.str("...")
		if e.Elt != nil {
			w.expr(e.Elt)
		}
	case *ast.MapType:
		w.str("map[")
		w.expr(e.Key)
		w.str("]")
		w.expr(e.Value)
	case *ast.ChanType:
		w.chanType(e)
	case *ast.FuncType:
		w.str("func")
		w.funcType(e, false)
	case *ast.IndexExpr:
		w.expr(e.X)
		w.str("[")
		w.expr(e.Index)
		w.str("]")
	case *ast.IndexListExpr:
		w.expr(e.X)
		w.str("[")
		w.exprList(e.Indices)
		w.str("]")
	case *ast.BasicLit:
		w.str(e.Value)
	case *ast.UnaryExpr:
		w.str(e.Op.String())
		w.expr(e.X)
	case *ast.BinaryExpr:
		w.expr(e.X)
		w.str(" " + e.Op.String() + " ")
		w.expr(e.Y)
	case *ast.CallExpr:
		w.expr(e.Fun)
		w.str("(")
		w.exprList(e.Args)
		if e.Ellipsis.IsValid() {
			w.str("...")
		}
		w.str(")")
	default:
		w.str(types.ExprString(e))
	}
}

func (w *writer) exprList(list []ast.Expr) {
	for i, e := range list {
		if i > 0 {
			w.str(", ")
		}
		w.expr(e)
	}
}

// selector writes a qualified identifier, linking the whole selector to the
// selected object when it resolves.
func (w *writer) selector(e *ast.SelectorExpr) {
	url := w.urlOf(e.Sel)
	start := w.buf.Len()
	w.expr(e.X)
	w.str(".")
	w.str(e.Sel.Name)
	if url != "" {
		w.links = append(w.links, Span{Start: start, End: w.buf.Len(), URL: url})
	}
}

func (w *writer) chanType(e *ast.ChanType) {
	switch e.Dir {
	case ast.RECV:
		w.str("<-chan ")
	case ast.SEND:
		w.str("chan<- ")
	default:
		w.str("chan ")
	}
	w.expr(e.Value)
}

// funcDecl writes a function or method declaration. When broken is true the
// parameters, and the results if they are still too wide, are placed one per
// line.
func (w *writer) funcDecl(d *ast.FuncDecl, broken bool) {
	w.str("func ")
	if d.Recv != nil {
		w.str("(")
		w.fieldList(d.Recv.List, ", ")
		w.str(") ")
	}
	w.linked(d.Name.Name, w.printer.Self)
	w.typeParams(d.Type.TypeParams)
	w.funcType(d.Type, broken)
}

// funcType writes the parameter and result lists of a function type.
func (w *writer) funcType(t *ast.FuncType, broken bool) {
	if broken {
		w.brokenFieldList(t.Params)
	} else {
		w.str("(")
		w.fieldList(fields(t.Params), ", ")
		w.str(")")
	}
	w.results(t.Results, broken)
}

func (w *writer) results(r *ast.FieldList, broken bool) {
	if r == nil || len(r.List) == 0 {
		return
	}
	w.str(" ")
	if len(r.List) == 1 && len(r.List[0].Names) == 0 {
		w.expr(r.List[0].Type)
		return
	}
	if broken && w.lastLineWidth()+w.inlineWidth(r) > MaxWidth {
		w.brokenFieldList(r)
		return
	}
	w.str("(")
	w.fieldList(r.List, ", ")
	w.str(")")
}

// inlineWidth measures a field list rendered on one line.
func (w *writer) inlineWidth(fl *ast.FieldList) int {
	probe := &writer{printer: w.printer}
	probe.str("(")
	probe.fieldList(fl.List, ", ")
	probe.str(")")
	return probe.width()
}

func (w *writer) lastLineWidth() int {
	text := w.buf.String()
	if i := strings.LastIndexByte(text, '\n'); i >= 0 {
		text = text[i+1:]
	}
	return utf8.RuneCountInString(text)
}

// brokenFieldList writes each field on its own line inside parentheses.
func (w *writer) brokenFieldList(fl *ast.FieldList) {
	w.str("(")
	list := fields(fl)
	if len(list) == 0 {
		w.str(")")
		return
	}
	w.indent++
	for _, f := range list {
		w.newline()
		w.field(f)
		w.str(",")
	}
	w.indent--
	w.newline()
	w.str(")")
}

func fields(fl *ast.FieldList) []*ast.Field {
	if fl == nil {
		return nil
	}
	return fl.List
}

// fieldList writes fields separated by sep.
func (w *writer) fieldList(list []*ast.Field, sep string) {
	for i, f := range list {
		if i > 0 {
			w.str(sep)
		}
		w.field(f)
	}
}

// field writes the names and type of one field.
func (w *writer) field(f *ast.Field) {
	for i, name := range f.Names {
		if i > 0 {
			w.str(", ")
		}
		w.str(name.Name)
	}
	if len(f.Names) > 0 {
		w.str(" ")
	}
	w.expr(f.Type)
}

func (w *writer) typeParams(tp *ast.FieldList) {
	if tp == nil || len(tp.List) == 0 {
		return
	}
	w.str("[")
	w.fieldList(tp.List, ", ")
	w.str("]")
}

// typeSpec writes a type declaration.
func (w *writer) typeSpec(s *ast.TypeSpec) {
	w.str("type ")
	w.linked(s.Name.Name, w.printer.Self)
	w.typeParams(s.TypeParams)
	if s.Assign.IsValid() {
		w.str(" = ")
	} else {
		w.str(" ")
	}
	switch t := s.Type.(type) {
	case *ast.StructType:
		w.structType(t)
	case *ast.InterfaceType:
		w.interfaceType(t)
	default:
		w.expr(s.Type)
	}
}

// structType writes the visible fields of a struct, one per line, and a
// comment in place of the hidden fields. The [go/doc] filter removes the
// unexported fields and marks the struct incomplete when the loader excludes
// them.
func (w *writer) structType(t *ast.StructType) {
	w.str("struct {")
	var members []func()
	hidden := t.Incomplete
	for _, f := range fields(t.Fields) {
		if !w.visible(f) {
			hidden = true
			continue
		}
		members = append(members, func() { w.structField(f) })
	}
	w.block(members, hidden, "// contains unexported fields")
}

func (w *writer) structField(f *ast.Field) {
	w.field(f)
	w.tag(f)
}

// fieldDecl writes one named field of a struct: the name, type, and tag. An
// embedded field is written as its type alone.
func (w *writer) fieldDecl(f *model.Field) {
	if !f.Embedded {
		w.str(f.Name)
		w.str(" ")
	}
	w.expr(f.Spec.Type)
	w.tag(f.Spec)
}

// tag writes the struct tag of f, when it has one.
func (w *writer) tag(f *ast.Field) {
	if f.Tag != nil {
		w.str(" ")
		w.str(f.Tag.Value)
	}
}

// interfaceType writes the visible methods and embedded types of an
// interface, one per line, and a comment in place of the hidden methods. The
// [go/doc] filter removes the unexported methods and marks the interface
// incomplete when the loader excludes them.
func (w *writer) interfaceType(t *ast.InterfaceType) {
	w.str("interface {")
	var members []func()
	hidden := t.Incomplete
	for _, f := range fields(t.Methods) {
		if !w.visible(f) {
			hidden = true
			continue
		}
		members = append(members, func() { w.interfaceMember(f) })
	}
	w.block(members, hidden, "// contains unexported methods")
}

func (w *writer) interfaceMember(f *ast.Field) {
	if len(f.Names) == 0 {
		w.expr(f.Type)
		return
	}
	w.str(f.Names[0].Name)
	if ft, ok := f.Type.(*ast.FuncType); ok {
		w.funcType(ft, false)
		return
	}
	w.expr(f.Type)
}

// block writes members separated by blank lines inside braces, followed by
// the hidden comment when any member was omitted.
func (w *writer) block(members []func(), hidden bool, comment string) {
	if len(members) == 0 && !hidden {
		w.str("}")
		return
	}
	w.indent++
	for i, member := range members {
		if i > 0 {
			w.blank()
		}
		w.newline()
		member()
	}
	if hidden {
		if len(members) > 0 {
			w.blank()
		}
		w.newline()
		w.str(comment)
	}
	w.indent--
	w.newline()
	w.str("}")
}

// visible reports whether a struct field or interface member is written: it
// is exported, or the printer writes unexported members.
func (w *writer) visible(f *ast.Field) bool {
	return w.printer.Unexported || fieldExported(f)
}

// fieldExported reports whether a struct field or interface member is
// exported. Embedded members are exported when their type name is.
func fieldExported(f *ast.Field) bool {
	if len(f.Names) > 0 {
		return f.Names[0].IsExported()
	}
	return ast.IsExported(embeddedName(f.Type))
}

// embeddedName returns the type name of an embedded field or interface.
func embeddedName(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return embeddedName(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.IndexExpr:
		return embeddedName(e.X)
	case *ast.IndexListExpr:
		return embeddedName(e.X)
	default:
		return ""
	}
}
