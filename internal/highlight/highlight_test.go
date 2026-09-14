package highlight_test

import (
	"html/template"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/highlight"
	"github.com/bitwizeshift/godoc2/internal/sig"
)

func TestHighlighter_Code(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		rendered sig.Rendered
		want     template.HTML
	}{
		{
			name: "no links",
			rendered: sig.Rendered{
				Text: "type ID = string",
			},
			want: `<pre class="chroma"><code><span class="kd">type</span> <span class="nx">ID</span> <span class="p">=</span> <span class="kt">string</span></code></pre>`,
		},
		{
			name: "single identifier link",
			rendered: sig.Rendered{
				Text: "func New() *Circle",
				Links: []sig.Span{
					{Start: 12, End: 18, URL: "Circle.html"},
				},
			},
			want: `<pre class="chroma"><code><span class="kd">func</span> <span class="nf">New</span><span class="p">()</span> <span class="o">*</span><a href="Circle.html"><span class="nx">Circle</span></a></code></pre>`,
		},
		{
			name: "link splits a token",
			rendered: sig.Rendered{
				Text: "ab",
				Links: []sig.Span{
					{Start: 1, End: 2, URL: "b.html"},
				},
			},
			want: `<pre class="chroma"><code><span class="nx">a</span><a href="b.html"><span class="nx">b</span></a></code></pre>`,
		},
		{
			name: "link across several tokens",
			rendered: sig.Rendered{
				Text: "var W io.Writer",
				Links: []sig.Span{
					{Start: 6, End: 15, URL: "https://pkg.go.dev/io#Writer"},
				},
			},
			want: `<pre class="chroma"><code><span class="kd">var</span> <span class="nx">W</span> <a href="https://pkg.go.dev/io#Writer"><span class="nx">io</span><span class="p">.</span><span class="nx">Writer</span></a></code></pre>`,
		},
		{
			name: "escapes html",
			rendered: sig.Rendered{
				Text: "chan<- int",
			},
			want: `<pre class="chroma"><code><span class="kd">chan</span><span class="o">&lt;-</span> <span class="kt">int</span></code></pre>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := highlight.New()

			// Act
			out, err := sut.Code(tc.rendered)

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Highlighter.Code(...) = %v, want nil", got)
			}
			if got, want := out, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Highlighter.Code(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestHighlighter_Source_WritesLinkableLines(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := highlight.New()
	var out strings.Builder

	// Act
	err := sut.Source(&out, []byte("package x\n\nvar A = 1\n"))
	text := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Highlighter.Source(...) = %v, want nil", got)
	}
	if got, want := strings.Contains(text, `id="L3"`), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.Source(...) contains id L3 = %v, want %v", got, want)
	}
	if got, want := strings.Contains(text, `href="#L1"`), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.Source(...) contains href L1 = %v, want %v", got, want)
	}
}

func TestHighlighter_CSS_WritesBothThemes(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := highlight.New()
	var out strings.Builder

	// Act
	err := sut.CSS(&out)
	css := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Highlighter.CSS(...) = %v, want nil", got)
	}
	if got, want := strings.Contains(css, ".chroma .k {"), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.CSS(...) has light keyword rule = %v, want %v", got, want)
	}
	if got, want := strings.Contains(css, `[data-theme="dark"] .chroma .k {`), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.CSS(...) has dark keyword rule = %v, want %v", got, want)
	}
	if got, want := strings.Contains(css, `:root:not([data-theme="light"]) .chroma .k {`), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.CSS(...) has system dark keyword rule = %v, want %v", got, want)
	}
	if got, want := strings.Contains(css, `[data-theme="dark"] .chroma .s {`), true; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.CSS(...) has dark string rule = %v, want %v", got, want)
	}
	if got, want := strings.Contains(css, ".chroma.light"), false; !cmp.Equal(got, want) {
		t.Errorf("Highlighter.CSS(...) has mode class = %v, want %v", got, want)
	}
}
