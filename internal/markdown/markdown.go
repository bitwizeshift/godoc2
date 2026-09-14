package markdown

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"

	"github.com/bitwizeshift/godoc2/internal/markdown/doclink"
)

// Renderer parses and renders doc comments.
type Renderer struct {
	parser parser.Parser
	html   html.Renderer
}

// New returns a [Renderer] configured for Go doc comments.
func New() *Renderer {
	return &Renderer{
		parser: parser.New(
			parser.WithAutoHeadingID(),
			parser.WithExtensions(extension.GFMParser, doclink.Parser),
		),
		html: html.New(
			html.WithExtensions(extension.GFMHTMLRenderer, doclink.HTMLRenderer),
		),
	}
}

// Parse parses a doc comment, resolving doc links through scope.
func (r *Renderer) Parse(doc string, scope doclink.Scope) *Document {
	source := []byte(doc)
	root := r.parser.Parse(source, parser.WithContext(doclink.NewContext(scope)))
	return &Document{renderer: r, source: source, root: root}
}

// Heading is a heading of a rendered document.
type Heading struct {
	Level int
	ID    string
	Text  string
}

// Document is a parsed doc comment.
type Document struct {
	renderer *Renderer
	source   []byte
	root     ast.Node
}

// HTML renders the whole document. It returns any error from the renderer.
func (d *Document) HTML() (template.HTML, error) {
	return d.render(d.root)
}

// Summary renders the first paragraph of the document, or nothing when the
// document has no paragraph. It returns any error from the renderer.
func (d *Document) Summary() (template.HTML, error) {
	for c := d.root.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindParagraph {
			return d.render(c)
		}
	}
	return "", nil
}

// Headings lists the headings of the document in order.
func (d *Document) Headings() []Heading {
	var result []Heading
	for c := d.root.FirstChild(); c != nil; c = c.NextSibling() {
		h, ok := c.(*ast.Heading)
		if !ok {
			continue
		}
		id := ""
		if v, ok := h.Attribute("id"); ok {
			id = v.Value(d.source)
		}
		result = append(result, Heading{Level: h.Level, ID: id, Text: d.text(h)})
	}
	return result
}

func (d *Document) render(n ast.Node) (template.HTML, error) {
	var buf bytes.Buffer
	if err := d.renderer.html.Render(&buf, d.source, n); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// text collects the plain text of the inline children of n.
func (d *Document) text(n ast.Node) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *ast.Text:
			b.WriteString(c.Value.Value(d.source))
		case *ast.CodeSpan:
			b.WriteString(c.Value.Value(d.source))
		case *doclink.Node:
			b.WriteString(c.Text)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}
