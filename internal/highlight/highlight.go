package highlight

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v3"
	chromahtml "github.com/alecthomas/chroma/v3/formatters/html"
	"github.com/alecthomas/chroma/v3/lexers"

	"github.com/bitwizeshift/godoc2/internal/sig"
)

const (
	// LineIDPrefix prefixes the id attribute of every source line, so that
	// line 42 is reachable as "#L42".
	LineIDPrefix = "L"
)

// Highlighter renders Go code as HTML.
type Highlighter struct {
	lexer  chroma.Lexer
	source *chromahtml.Formatter
	light  *chroma.Style
	dark   *chroma.Style
}

// New returns a [Highlighter] for Go source.
func New() *Highlighter {
	return &Highlighter{
		lexer: chroma.Coalesce(lexers.Get("go")),
		source: chromahtml.New(
			chromahtml.WithClasses(true),
			chromahtml.WithModeClasses(true),
			chromahtml.WithLineNumbers(true),
			chromahtml.LineNumbersInTable(true),
			chromahtml.WithLinkableLineNumbers(true, LineIDPrefix),
			chromahtml.TabWidth(4),
		),
		light: lightStyle,
		dark:  darkStyle,
	}
}

// Code renders a declaration as a highlighted <pre> block. Every link span is
// wrapped in an anchor. It returns any error from the lexer.
func (h *Highlighter) Code(r sig.Rendered) (template.HTML, error) {
	tokens, err := h.lexer.Tokenise(nil, r.Text)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(`<pre class="chroma"><code>`)
	offset := 0
	links := r.Links
	for tok := range tokens {
		offset = h.writeToken(&b, tok, offset, &links)
	}
	b.WriteString("</code></pre>")
	return template.HTML(b.String()), nil
}

// Snippet renders a code fragment without links.
func (h *Highlighter) Snippet(code string) (template.HTML, error) {
	return h.Code(sig.Rendered{Text: code})
}

// writeToken writes one token, splitting it where a link span starts or
// ends inside it. It returns the byte offset after the token.
func (h *Highlighter) writeToken(b *strings.Builder, tok chroma.Token, offset int, links *[]sig.Span) int {
	text := tok.Value
	for text != "" {
		cut := len(text)
		if len(*links) > 0 {
			l := (*links)[0]
			switch {
			case l.Start == offset:
				fmt.Fprintf(b, `<a href="%s">`, html.EscapeString(l.URL))
				cut = min(cut, l.End-offset)
			case l.Start > offset:
				cut = min(cut, l.Start-offset)
			default:
				cut = min(cut, l.End-offset)
			}
		}
		h.writeSpan(b, tok.Type, text[:cut])
		offset += cut
		text = text[cut:]
		if len(*links) > 0 && (*links)[0].End == offset {
			b.WriteString("</a>")
			*links = (*links)[1:]
		}
	}
	return offset
}

func (h *Highlighter) writeSpan(b *strings.Builder, tt chroma.TokenType, text string) {
	escaped := html.EscapeString(text)
	if cls := class(tt); cls != "" && tt != chroma.TextWhitespace {
		fmt.Fprintf(b, `<span class="%s">%s</span>`, cls, escaped)
		return
	}
	b.WriteString(escaped)
}

// class returns the CSS class chroma assigns to a token type, walking up the
// type hierarchy the way the chroma HTML formatter does.
func class(t chroma.TokenType) string {
	for t != 0 {
		if cls, ok := chroma.StandardTypes[t]; ok {
			return cls
		}
		t = t.Parent()
	}
	return chroma.StandardTypes[t]
}

// Source writes a whole file with line numbers to w. It returns any error
// from the lexer or the writer.
func (h *Highlighter) Source(w io.Writer, src []byte) error {
	tokens, err := h.lexer.Tokenise(nil, string(src))
	if err != nil {
		return err
	}
	return h.source.Format(w, h.light, tokens)
}

// CSS writes the highlighting rules for both themes to w. Light rules apply
// by default. Dark rules apply under a data-theme="dark" ancestor, and under
// the dark system colour scheme unless data-theme="light" is set. Every token
// class is written for both themes so that a dark rule overrides its light
// counterpart.
func (h *Highlighter) CSS(w io.Writer) error {
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.WithAllClasses(true),
		chromahtml.WithModeClasses(true),
		chromahtml.WithLineNumbers(true),
		chromahtml.LineNumbersInTable(true),
		chromahtml.WithLinkableLineNumbers(true, LineIDPrefix),
		chromahtml.WithCSSComments(false),
	)
	var light, dark bytes.Buffer
	if err := formatter.WriteCSS(&light, h.light); err != nil {
		return err
	}
	if err := formatter.WriteCSS(&dark, h.dark); err != nil {
		return err
	}
	lightCSS := strings.NewReplacer(".chroma.light", ".chroma", ".bg.light", ".bg").Replace(light.String())
	darkCSS := dark.String() + resetRules(light.String(), dark.String())
	toggled := strings.NewReplacer(
		".chroma.dark", `[data-theme="dark"] .chroma`,
		".bg.dark", `[data-theme="dark"] .bg`,
	).Replace(darkCSS)
	system := strings.NewReplacer(
		".chroma.dark", `:root:not([data-theme="light"]) .chroma`,
		".bg.dark", `:root:not([data-theme="light"]) .bg`,
	).Replace(darkCSS)
	css := lightCSS + toggled + "@media (prefers-color-scheme: dark) {\n" + system + "}\n"
	_, err := io.WriteString(w, css)
	return err
}

var tokenRule = regexp.MustCompile(`(?m)^\.chroma\.\w+ (\.\S+) \{`)

// resetRules returns dark-scoped rules that reset every token class the light
// theme styles but the dark theme leaves unstyled, so that the light colour
// does not leak into the dark theme.
func resetRules(light, dark string) string {
	defined := map[string]bool{}
	for _, m := range tokenRule.FindAllStringSubmatch(dark, -1) {
		defined[m[1]] = true
	}
	var b strings.Builder
	for _, m := range tokenRule.FindAllStringSubmatch(light, -1) {
		if defined[m[1]] {
			continue
		}
		defined[m[1]] = true
		fmt.Fprintf(&b, ".chroma.dark %s { color: inherit; background-color: transparent }\n", m[1])
	}
	return b.String()
}
