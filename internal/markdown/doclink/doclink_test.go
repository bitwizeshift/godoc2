package doclink_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"

	"github.com/bitwizeshift/godoc2/internal/markdown/doclink"
	"github.com/bitwizeshift/godoc2/internal/markdown/doclink/doclinktest"
)

func render(t testing.TB, scope doclink.Scope, source string) string {
	t.Helper()
	p := parser.New(parser.WithExtensions(doclink.Parser))
	r := html.New(html.WithExtensions(doclink.HTMLRenderer))
	doc := p.Parse([]byte(source), parser.WithContext(doclink.NewContext(scope)))
	var buf bytes.Buffer
	if err := r.Render(&buf, []byte(source), doc); err != nil {
		t.Fatalf("Render(...) = %v, want nil", err)
	}
	return buf.String()
}

func TestParser(t *testing.T) {
	t.Parallel()

	scope := doclinktest.MapScope{
		Symbols: map[string]string{
			"Circle":          "Circle.html",
			"Circle.Area":     "Circle.Area.html",
			"io.Writer":       "https://pkg.go.dev/io#Writer",
			"io.Writer.Write": "https://pkg.go.dev/io#Writer.Write",
			"net/http.Client": "https://pkg.go.dev/net/http#Client",
		},
		Packages: map[string]string{
			"net/http": "https://pkg.go.dev/net/http",
			"shapes":   "shapes/index.html",
		},
	}

	testCases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "local symbol",
			source: "See [Circle].",
			want:   `<p>See <a href="Circle.html"><code>Circle</code></a>.</p>` + "\n",
		},
		{
			name:   "local method",
			source: "See [Circle.Area].",
			want:   `<p>See <a href="Circle.Area.html"><code>Circle.Area</code></a>.</p>` + "\n",
		},
		{
			name:   "pointer form",
			source: "See [*Circle].",
			want:   `<p>See <a href="Circle.html"><code>*Circle</code></a>.</p>` + "\n",
		},
		{
			name:   "imported symbol",
			source: "See [io.Writer].",
			want:   `<p>See <a href="https://pkg.go.dev/io#Writer"><code>io.Writer</code></a>.</p>` + "\n",
		},
		{
			name:   "imported method",
			source: "See [io.Writer.Write].",
			want:   `<p>See <a href="https://pkg.go.dev/io#Writer.Write"><code>io.Writer.Write</code></a>.</p>` + "\n",
		},
		{
			name:   "import path symbol",
			source: "See [net/http.Client].",
			want:   `<p>See <a href="https://pkg.go.dev/net/http#Client"><code>net/http.Client</code></a>.</p>` + "\n",
		},
		{
			name:   "import path package",
			source: "See [net/http].",
			want:   `<p>See <a href="https://pkg.go.dev/net/http"><code>net/http</code></a>.</p>` + "\n",
		},
		{
			name:   "package name",
			source: "See [shapes].",
			want:   `<p>See <a href="shapes/index.html"><code>shapes</code></a>.</p>` + "\n",
		},
		{
			name:   "unresolved stays literal",
			source: "See [Missing].",
			want:   "<p>See [Missing].</p>\n",
		},
		{
			name:   "inline link is untouched",
			source: "See [Circle](http://example.com).",
			want:   `<p>See <a href="http://example.com">Circle</a>.</p>` + "\n",
		},
		{
			name:   "reference link definition wins",
			source: "See [Circle].\n\n[Circle]: http://example.com\n",
			want:   `<p>See <a href="http://example.com">Circle</a>.</p>` + "\n",
		},
		{
			name:   "repeated links",
			source: "See [Circle] and [Circle].",
			want:   `<p>See <a href="Circle.html"><code>Circle</code></a> and <a href="Circle.html"><code>Circle</code></a>.</p>` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			out := render(t, scope, tc.source)

			// Assert
			if got, want := out, tc.want; !cmp.Equal(got, want) {
				t.Errorf("render(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestParser_WithoutScope_LeavesText(t *testing.T) {
	t.Parallel()

	// Arrange
	p := parser.New(parser.WithExtensions(doclink.Parser))
	r := html.New(html.WithExtensions(doclink.HTMLRenderer))
	source := []byte("See [Circle].")
	var buf bytes.Buffer

	// Act
	doc := p.Parse(source)
	err := r.Render(&buf, source, doc)
	out := buf.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Render(...) = %v, want nil", got)
	}
	if got, want := out, "<p>See [Circle].</p>\n"; !cmp.Equal(got, want) {
		t.Errorf("Render(...) = %q, want %q", got, want)
	}
}
