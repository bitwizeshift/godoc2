package markdown_test

import (
	"html/template"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
			sut := markdown.New().Parse(tc.doc, scope())

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
			sut := markdown.New().Parse(tc.doc, scope())

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
	sut := markdown.New().Parse(sampleDoc, scope())
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
