package markdown_test

import (
	"html/template"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/highlight"
	"github.com/bitwizeshift/godoc2/internal/markdown"
	"github.com/bitwizeshift/godoc2/internal/markdown/doclink/doclinktest"
)

const sampleDoc = "Package sample does *things* with [Circle].\n\n" +
	"Second paragraph.\n\n" +
	"# Usage\n\n" +
	"\tc := sample.NewCircle(2)\n\n" +
	"## The `Color` type\n\n" +
	"| a | b |\n|---|---|\n| 1 | 2 |\n"

func scope() doclinktest.MapScope {
	return doclinktest.MapScope{
		Symbols: map[string]string{"Circle": "Circle.html"},
	}
}

func TestDocument_HTML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		doc  string
		want template.HTML
	}{
		{
			name: "paragraph with doc link and emphasis",
			doc:  "Does *things* with [Circle].",
			want: `<p>Does <em>things</em> with <a href="Circle.html"><code>Circle</code></a>.</p>` + "\n",
		},
		{
			name: "heading gets an id",
			doc:  "# Usage Notes\n\nText.",
			want: `<h1 id="usage-notes">Usage Notes</h1>` + "\n<p>Text.</p>\n",
		},
		{
			name: "indented code block",
			doc:  "Example:\n\n\tc := New()\n",
			want: "<p>Example:</p>\n<pre><code>c := New()\n</code></pre>\n",
		},
		{
			name: "fenced code block without a language",
			doc:  "```\nc := New()\n```\n",
			want: "<pre><code>c := New()\n</code></pre>\n",
		},
		{
			name: "fenced code block with a language",
			doc:  "```go\nvar x int\n```\n",
			want: `<pre class="chroma"><code><span class="kd">var</span> <span class="nx">x</span> <span class="kt">int</span>` + "\n</code></pre>\n",
		},
		{
			name: "fenced code block with an unknown language",
			doc:  "```no-such-language\nx\n```\n",
			want: `<pre><code class="language-no-such-language">x` + "\n</code></pre>\n",
		},
		{
			name: "gfm table",
			doc:  "| a | b |\n|---|---|\n| 1 | 2 |\n",
			want: "<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			name: "raw html is escaped",
			doc:  "Use <b>bold</b>.",
			want: "<p>Use <!-- raw HTML omitted -->bold<!-- raw HTML omitted -->.</p>\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := markdown.New(highlight.New()).Parse(tc.doc, scope())

			// Act
			out, err := sut.HTML()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Document.HTML() = %v, want nil", got)
			}
			if got, want := out, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Document.HTML() mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestRenderer_ParseMarkdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		text string
		want template.HTML
	}{
		{
			name: "doc link stays literal",
			text: "See [Circle].",
			want: "<p>See [Circle].</p>\n",
		},
		{
			name: "raw html is kept",
			text: "<p align=\"center\">Centred</p>\n\nUse <b>bold</b>.",
			want: "<p align=\"center\">Centred</p>\n<p>Use <b>bold</b>.</p>\n",
		},
		{
			name: "heading gets an id",
			text: "# Usage Notes\n\nText.",
			want: `<h1 id="usage-notes">Usage Notes</h1>` + "\n<p>Text.</p>\n",
		},
		{
			name: "gfm autolink",
			text: "Visit https://example.com now.",
			want: `<p>Visit <a href="https://example.com">https://example.com</a> now.</p>` + "\n",
		},
		{
			name: "fenced code block with a language",
			text: "```go\nvar x int\n```\n",
			want: `<pre class="chroma"><code><span class="kd">var</span> <span class="nx">x</span> <span class="kt">int</span>` + "\n</code></pre>\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := markdown.New(highlight.New()).ParseMarkdown(tc.text)

			// Act
			out, err := sut.HTML()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Document.HTML() = %v, want nil", got)
			}
			if got, want := out, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Document.HTML() mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestDocument_RewriteLinks(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := markdown.New(highlight.New()).ParseMarkdown("[a](docs/a.md) and ![b](img/b.png \"B\") and [c](https://example.com).")
	sut.RewriteLinks(func(dest string) string {
		switch dest {
		case "docs/a.md":
			return "../docs/a.md"
		case "img/b.png":
			return "../img/b.png"
		default:
			return dest
		}
	})
	want := template.HTML(`<p><a href="../docs/a.md">a</a> and <img src="../img/b.png" alt="b" title="B"> and <a href="https://example.com">c</a>.</p>` + "\n")

	// Act
	out, err := sut.HTML()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Document.HTML() = %v, want nil", got)
	}
	if got, want := out, want; !cmp.Equal(got, want) {
		t.Errorf("Document.HTML() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestDocument_Summary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		doc  string
		want template.HTML
	}{
		{
			name: "first paragraph only",
			doc:  sampleDoc,
			want: `<p>Package sample does <em>things</em> with <a href="Circle.html"><code>Circle</code></a>.</p>` + "\n",
		},
		{
			name: "heading before paragraph",
			doc:  "# Title\n\nBody text.\n",
			want: "<p>Body text.</p>\n",
		},
		{
			name: "no paragraph",
			doc:  "# Title\n",
			want: "",
		},
		{
			name: "empty",
			doc:  "",
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := markdown.New(highlight.New()).Parse(tc.doc, scope())

			// Act
			out, err := sut.Summary()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Document.Summary() = %v, want nil", got)
			}
			if got, want := out, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Document.Summary() mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestDocument_Headings_ListsHeadingsInOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := markdown.New(highlight.New()).Parse(sampleDoc, scope())
	want := []markdown.Heading{
		{Level: 1, ID: "usage", Text: "Usage"},
		{Level: 2, ID: "the-color-type", Text: "The Color type"},
	}

	// Act
	headings := sut.Headings()

	// Assert
	if got, want := headings, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Document.Headings() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}
