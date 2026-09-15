package markdown

import (
	"html/template"
	"io"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
)

// Highlighter renders the code of a fenced block written in a named
// language.
type Highlighter interface {
	Fence(lang, code string) (template.HTML, error)
}

// fenceRenderer renders a code block that names a language through h. A block
// without a language, and a block whose language h rejects, renders through
// next.
type fenceRenderer struct {
	h    Highlighter
	next html.NodeRenderer
}

// fenceDecorator returns the decorator that wraps the code block renderer
// with h. A nil h leaves the code block renderer as it is.
func fenceDecorator(h Highlighter) html.NodeRendererDecorator {
	return func(next html.NodeRenderer) html.NodeRenderer {
		if h == nil {
			return next
		}
		return &fenceRenderer{h: h, next: next}
	}
}

// Render writes the whole block on entering, because a code block has no
// children. The exit call writes nothing for a block with a language.
func (r *fenceRenderer) Render(w io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	block := n.(*ast.CodeBlock)
	lang, ok := block.Language(source)
	if !ok {
		return r.next.Render(w, source, n, entering, rc)
	}
	if !entering {
		return ast.WalkContinue, nil
	}
	out, err := r.h.Fence(lang, block.Value.Str(source))
	if err != nil {
		return r.fallback(w, source, n, rc)
	}
	if _, err := io.WriteString(w, string(out)+"\n"); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkContinue, nil
}

// fallback renders the block through next, entering and leaving in one call.
func (r *fenceRenderer) fallback(w io.Writer, source []byte, n ast.Node, rc renderer.Context) (ast.WalkStatus, error) {
	if _, err := r.next.Render(w, source, n, true, rc); err != nil {
		return ast.WalkStop, err
	}
	return r.next.Render(w, source, n, false, rc)
}

var _ html.NodeRenderer = (*fenceRenderer)(nil)
