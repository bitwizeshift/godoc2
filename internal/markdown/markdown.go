package markdown

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"

	"github.com/bitwizeshift/godoc2/internal/markdown/doclink"
)

// Renderer parses and renders doc comments and Markdown files.
type Renderer struct {
	parser   parser.Parser
	markdown parser.Parser
	html     html.Renderer
	unsafe   html.Renderer
}

// New returns a [Renderer] configured for Go doc comments. Code blocks that
// name a language render through h. A nil h renders every code block as
// plain text.
func New(h Highlighter) *Renderer {
	fences := html.WithNodeRendererDecorator(ast.KindCodeBlock, fenceDecorator(h))
	return &Renderer{
		parser: parser.New(
			parser.WithAutoHeadingID(),
			parser.WithExtensions(extension.GFMParser, doclink.Parser),
		),
		markdown: parser.New(
			parser.WithAutoHeadingID(),
			parser.WithExtensions(extension.GFMParser),
		),
		html: html.New(
			html.WithExtensions(extension.GFMHTMLRenderer, doclink.HTMLRenderer),
			fences,
		),
		unsafe: html.New(
			html.WithExtensions(extension.GFMHTMLRenderer),
			html.WithUnsafe(),
			fences,
		),
	}
}

// Parse parses a doc comment, resolving doc links through scope.
func (r *Renderer) Parse(doc string, scope doclink.Scope) *Document {
	source := []byte(doc)
	root := r.parser.Parse(source, parser.WithContext(doclink.NewContext(scope)))
	return &Document{html: r.html, source: source, root: root}
}

// ParseMarkdown parses a Markdown file as GitHub Flavored Markdown. Doc
// links such as [Name] are not resolved, and raw HTML is kept as written.
func (r *Renderer) ParseMarkdown(text string) *Document {
	source := []byte(text)
	root := r.markdown.Parse(source)
	return &Document{html: r.unsafe, source: source, root: root}
}

// Heading is a heading of a rendered document.
type Heading struct {
	Level int
	ID    string
	Text  string
}

// Document is a parsed doc comment.
type Document struct {
	html   html.Renderer
	source []byte
	root   ast.Node
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

// RewriteLinks replaces the destination of every link and image with the
// result of fn applied to the written destination.
func (d *Document) RewriteLinks(fn func(dest string) string) {
	_ = ast.Walk(d.root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Link:
			n.Destination = d.rewrite(n.Destination, fn)
		case *ast.Image:
			n.Destination = d.rewrite(n.Destination, fn)
		}
		return ast.WalkContinue, nil
	})
}

func (d *Document) rewrite(dest text.SingleLineValue, fn func(string) string) text.SingleLineValue {
	written := dest.Value(d.source)
	result := fn(written)
	if result == written {
		return dest
	}
	return text.NewSingleLineValueFromString(result, nil)
}

func (d *Document) render(n ast.Node) (template.HTML, error) {
	var buf bytes.Buffer
	if err := d.html.Render(&buf, d.source, n); err != nil {
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
