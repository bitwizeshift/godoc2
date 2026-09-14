package pathmap_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/pathmap"
)

const module = "github.com/example/mod"

func TestModule(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.Module(module)

	// Assert
	if got, want := p, "github.com/example/mod/index.html"; !cmp.Equal(got, want) {
		t.Errorf("Module(...) = %q, want %q", got, want)
	}
}

func TestPackage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		rel  string
		want string
	}{
		{
			name: "root package",
			rel:  "",
			want: "github.com/example/mod/index.html",
		},
		{
			name: "nested package",
			rel:  "internal/args",
			want: "github.com/example/mod/internal/args/index.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			p := pathmap.Package(module, tc.rel)

			// Assert
			if got, want := p, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Package(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestSymbol(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.Symbol(module, "args", "CommandLine")

	// Assert
	if got, want := p, "github.com/example/mod/args/CommandLine.html"; !cmp.Equal(got, want) {
		t.Errorf("Symbol(...) = %q, want %q", got, want)
	}
}

func TestMethod(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.Method(module, "args", "Widget", "Execute")

	// Assert
	if got, want := p, "github.com/example/mod/args/Widget.Execute.html"; !cmp.Equal(got, want) {
		t.Errorf("Method(...) = %q, want %q", got, want)
	}
}

func TestSource(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.Source(module, "args", "commandline.go")

	// Assert
	if got, want := p, "github.com/example/mod/args/commandline.go.html"; !cmp.Equal(got, want) {
		t.Errorf("Source(...) = %q, want %q", got, want)
	}
}

func TestFile(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.File(module, "docs/images/logo.png")

	// Assert
	if got, want := p, "github.com/example/mod/docs/images/logo.png"; !cmp.Equal(got, want) {
		t.Errorf("File(...) = %q, want %q", got, want)
	}
}

func TestStatic(t *testing.T) {
	t.Parallel()

	// Act
	p := pathmap.Static("godoc2.css")

	// Assert
	if got, want := p, "static/godoc2.css"; !cmp.Equal(got, want) {
		t.Errorf("Static(...) = %q, want %q", got, want)
	}
}

func TestRel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		from string
		to   string
		want string
	}{
		{
			name: "same directory",
			from: "a/b/index.html",
			to:   "a/b/Foo.html",
			want: "Foo.html",
		},
		{
			name: "child directory",
			from: "a/index.html",
			to:   "a/b/index.html",
			want: "b/index.html",
		},
		{
			name: "parent directory",
			from: "a/b/index.html",
			to:   "a/index.html",
			want: "../index.html",
		},
		{
			name: "sibling directory",
			from: "a/b/index.html",
			to:   "a/c/Foo.html",
			want: "../c/Foo.html",
		},
		{
			name: "static asset",
			from: "github.com/example/mod/args/index.html",
			to:   "static/godoc2.css",
			want: "../../../../static/godoc2.css",
		},
		{
			name: "from root",
			from: "index.html",
			to:   "a/b/index.html",
			want: "a/b/index.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			href := pathmap.Rel(tc.from, tc.to)

			// Assert
			if got, want := href, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Rel(...) = %q, want %q", got, want)
			}
		})
	}
}
