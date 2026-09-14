package doclink

import (
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// Scope resolves the targets of doc links for one document.
type Scope interface {
	// Symbol returns the URL of name, or of its method when method is not
	// empty, in the package identified by pkg. An empty pkg means the current
	// package. It reports false when the symbol does not resolve.
	Symbol(pkg, name, method string) (string, bool)

	// Package returns the URL of the package identified by pkg, which is
	// either an import path or the name of an imported package. It reports
	// false when the package does not resolve.
	Package(pkg string) (string, bool)
}

// Node is a resolved doc link.
type Node struct {
	ast.BaseInline

	// Text is the link text without the brackets.
	Text string

	// URL is the resolved destination.
	URL string
}

// Kind is the node kind of [Node].
var Kind = ast.NewNodeKind("DocLink")

// Kind returns [Kind].
func (n *Node) Kind() ast.NodeKind { return Kind }

// Dump implements [ast.Node].
func (n *Node) Dump(_ []byte) *ast.NodeDump {
	return ast.NewNodeDump(n, map[string]any{"Text": n.Text, "URL": n.URL})
}

// FirstRune returns the first rune of the link text.
func (n *Node) FirstRune(_ []byte) (rune, bool) {
	r, size := utf8.DecodeRuneInString(n.Text)
	return r, size > 0
}

// LastRune returns the last rune of the link text.
func (n *Node) LastRune(_ []byte) (rune, bool) {
	r, size := utf8.DecodeLastRuneInString(n.Text)
	return r, size > 0
}

// NewNode returns a doc link node.
func NewNode(txt, url string) *Node {
	n := &Node{Text: txt, URL: url}
	n.Init(n)
	return n
}

var _ ast.InlineNode = (*Node)(nil)

var scopeKey = parser.NewContextKey()

// NewContext returns a parse context that resolves doc links through scope.
func NewContext(scope Scope) parser.Context {
	pc := parser.NewContext()
	pc.Set(scopeKey, scope)
	return pc
}

// Priority places the doc link parser before the standard link parser.
const Priority = 150

// Parser is the parser extension.
var Parser parser.Extension = parserExtension{}

type parserExtension struct{}

func (parserExtension) ParserOptions(*parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(util.Prioritized[parser.InlineParser](inlineParser{}, Priority)),
	}
}

// HTMLRenderer is the HTML renderer extension.
var HTMLRenderer html.Extension = rendererExtension{}

type rendererExtension struct{}

func (rendererExtension) RendererOptions(*html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
			Kind: renderer.NodeRendererFunc(renderNode),
		}),
	}
}

// linkPattern matches "[", an optional "*", dot-separated identifiers where
// the first may be an import path, and "]".
var linkPattern = regexp.MustCompile(`^\[(\*?)([A-Za-z_][A-Za-z0-9_./-]*)\]`)

type inlineParser struct{}

func (inlineParser) Trigger() []byte {
	return []byte{'['}
}

func (inlineParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	scope, _ := pc.Get(scopeKey).(Scope)
	if scope == nil {
		return nil
	}
	line, _ := block.PeekLine()
	m := linkPattern.FindSubmatch(line)
	if m == nil || followedByLinkSyntax(line[len(m[0]):]) {
		return nil
	}
	label := string(m[2])
	if _, defined := pc.LinkDefinition(strings.ToLower(label)); defined {
		return nil
	}
	url, ok := resolve(scope, label)
	if !ok {
		return nil
	}
	block.Advance(len(m[0]))
	return NewNode(string(m[1])+label, url)
}

// followedByLinkSyntax reports whether the bracketed text is the label of a
// Markdown inline or reference link.
func followedByLinkSyntax(rest []byte) bool {
	return len(rest) > 0 && (rest[0] == '(' || rest[0] == '[' || rest[0] == ':')
}

// resolve splits the label into its package, name, and method parts and
// asks the scope for the URL.
func resolve(scope Scope, label string) (string, bool) {
	pkgPath, rest, hasSlash := splitPath(label)
	parts := strings.Split(rest, ".")
	switch {
	case hasSlash && rest == "":
		return scope.Package(pkgPath)
	case hasSlash && len(parts) == 1:
		return scope.Symbol(pkgPath, parts[0], "")
	case hasSlash && len(parts) == 2:
		return scope.Symbol(pkgPath, parts[0], parts[1])
	case hasSlash:
		return "", false
	}
	switch len(parts) {
	case 1:
		if url, ok := scope.Symbol("", parts[0], ""); ok {
			return url, true
		}
		return scope.Package(parts[0])
	case 2:
		if startsLower(parts[0]) {
			if url, ok := scope.Symbol(parts[0], parts[1], ""); ok {
				return url, true
			}
		}
		return scope.Symbol("", parts[0], parts[1])
	case 3:
		return scope.Symbol(parts[0], parts[1], parts[2])
	default:
		return "", false
	}
}

// splitPath separates an import path prefix, which contains a slash, from the
// dotted symbol that follows its last element.
func splitPath(label string) (pkgPath, rest string, ok bool) {
	slash := strings.LastIndex(label, "/")
	if slash < 0 {
		return "", label, false
	}
	last := label[slash+1:]
	dot := strings.Index(last, ".")
	if dot < 0 {
		return label, "", true
	}
	return label[:slash+1+dot], last[dot+1:], true
}

func startsLower(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLower(r)
}

func renderNode(w io.Writer, _ []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*Node)
	bw := w.(util.BufWriter)
	_, _ = bw.WriteString(`<a href="`)
	_, _ = html.ContextTextWriter(rc).WriteString(node.URL)
	_, _ = bw.WriteString(`"><code>`)
	_, _ = html.ContextTextWriter(rc).WriteString(node.Text)
	_, _ = bw.WriteString(`</code></a>`)
	return ast.WalkContinue, nil
}
