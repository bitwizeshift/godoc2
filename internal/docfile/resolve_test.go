package docfile_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
)

func readmeFile(t testing.TB) *model.DocFile {
	t.Helper()
	return loadertest.Package(t, "readme").DocFile
}

func TestResolver_Resolve(t *testing.T) {
	t.Parallel()

	const from = "example.com/sample/readme/index.html"

	testCases := []struct {
		name string
		from string
		dest string
		want string
	}{
		{
			name: "package directory",
			from: from,
			dest: "../shapes",
			want: "../shapes/index.html",
		},
		{
			name: "module root",
			from: from,
			dest: "..",
			want: "../index.html",
		},
		{
			name: "own directory",
			from: from,
			dest: ".",
			want: "index.html",
		},
		{
			name: "Go file with line",
			from: from,
			dest: "readme.go#L4",
			want: "readme.go.html#L4",
		},
		{
			name: "copied file",
			from: from,
			dest: "docs/guide.md",
			want: "docs/guide.md",
		},
		{
			name: "copied file from another page",
			from: "example.com/sample/index.html",
			dest: "docs/guide.md",
			want: "readme/docs/guide.md",
		},
		{
			name: "missing file",
			from: from,
			dest: "docs/missing.md",
			want: "docs/missing.md",
		},
		{
			name: "outside the module",
			from: from,
			dest: "../../outside.md",
			want: "../../outside.md",
		},
		{
			name: "directory without package",
			from: from,
			dest: "docs",
			want: "docs",
		},
		{
			name: "absolute path",
			from: from,
			dest: "/absolute",
			want: "/absolute",
		},
		{
			name: "external URL",
			from: from,
			dest: "https://example.com/x",
			want: "https://example.com/x",
		},
		{
			name: "fragment only",
			from: from,
			dest: "#links",
			want: "#links",
		},
		{
			name: "mailto",
			from: from,
			dest: "mailto:someone@example.com",
			want: "mailto:someone@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := docfile.NewResolver(loadertest.Sample(t))
			f := readmeFile(t)

			// Act
			href := sut.Resolve(f, tc.from, tc.dest)

			// Assert
			if got, want := href, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Resolver.Resolve(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestResolver_Assets(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := docfile.NewResolver(loadertest.Sample(t))
	f := readmeFile(t)
	const from = "example.com/sample/readme/index.html"
	dir := filepath.Dir(f.Path)
	want := []docfile.Asset{
		{Source: filepath.Join(dir, "docs", "diagram.svg"), Output: "example.com/sample/readme/docs/diagram.svg"},
		{Source: filepath.Join(dir, "docs", "guide.md"), Output: "example.com/sample/readme/docs/guide.md"},
	}
	sut.Resolve(f, from, "docs/guide.md")
	sut.Resolve(f, from, "docs/diagram.svg")
	sut.Resolve(f, from, "docs/guide.md")
	sut.Resolve(f, from, "../shapes")
	sut.Resolve(f, from, "docs/missing.md")

	// Act
	assets := sut.Assets()

	// Assert
	if got, want := assets, want; !cmp.Equal(got, want) {
		t.Errorf("Resolver.Assets() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}
