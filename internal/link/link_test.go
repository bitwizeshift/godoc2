package link_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
)

const rootIndex = "example.com/sample/index.html"

func TestResolver_URL(t *testing.T) {
	t.Parallel()

	site := loadertest.Sample(t)
	root := loadertest.Package(t, "")
	circle := loadertest.Type(t, "", "Circle")
	area := loadertest.Method(t, "", "Circle", "Area")
	fmtPkg := importOf(t, root, "fmt")

	testCases := []struct {
		name   string
		from   string
		obj    types.Object
		want   string
		wantOK bool
	}{
		{
			name:   "local type from root",
			from:   rootIndex,
			obj:    circle.Obj,
			want:   "Circle.html",
			wantOK: true,
		},
		{
			name:   "local type from nested package",
			from:   "example.com/sample/shapes/index.html",
			obj:    circle.Obj,
			want:   "../Circle.html",
			wantOK: true,
		},
		{
			name:   "local method",
			from:   rootIndex,
			obj:    area.Obj,
			want:   "Circle.Area.html",
			wantOK: true,
		},
		{
			name:   "local function",
			from:   rootIndex,
			obj:    loadertest.Func(t, "", "NewCircle").Obj,
			want:   "NewCircle.html",
			wantOK: true,
		},
		{
			name:   "local constant",
			from:   rootIndex,
			obj:    loadertest.Value(t, "", "Red").Obj,
			want:   "Red.html",
			wantOK: true,
		},
		{
			name:   "standard library type",
			from:   rootIndex,
			obj:    fmtPkg.Scope().Lookup("Stringer"),
			want:   "https://pkg.go.dev/fmt#Stringer",
			wantOK: true,
		},
		{
			name:   "universe type",
			from:   rootIndex,
			obj:    types.Universe.Lookup("error"),
			want:   "https://pkg.go.dev/builtin#error",
			wantOK: true,
		},
		{
			name:   "struct field",
			from:   rootIndex,
			obj:    fieldOf(t, circle, "Radius"),
			want:   "",
			wantOK: false,
		},
		{
			name:   "nil object",
			from:   rootIndex,
			obj:    nil,
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := link.New(site)

			// Act
			url, ok := sut.URL(tc.from, tc.obj)

			// Assert
			if got, want := ok, tc.wantOK; !cmp.Equal(got, want) {
				t.Fatalf("Resolver.URL(...) ok = %v, want %v", got, want)
			}
			if got, want := url, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Resolver.URL(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestResolver_URL_WithThirdPartyObject_PinsVersion(t *testing.T) {
	t.Parallel()

	// Arrange
	depPkg := types.NewPackage("github.com/acme/widgets", "widgets")
	obj := types.NewTypeName(token.NoPos, depPkg, "Widget", nil)
	depPkg.Scope().Insert(obj)
	site := &model.Site{
		Modules: []*model.Module{{Path: "example.com/mod"}},
		Deps: map[string]model.Dependency{
			"github.com/acme/widgets": {Path: "github.com/acme/widgets", Version: "v1.2.3"},
		},
	}
	sut := link.New(site)

	// Act
	url, ok := sut.URL("example.com/mod/index.html", obj)

	// Assert
	if got, want := ok, true; !cmp.Equal(got, want) {
		t.Fatalf("Resolver.URL(...) ok = %v, want %v", got, want)
	}
	if got, want := url, "https://pkg.go.dev/github.com/acme/widgets@v1.2.3#Widget"; !cmp.Equal(got, want) {
		t.Errorf("Resolver.URL(...) = %q, want %q", got, want)
	}
}

func TestResolver_Lookup(t *testing.T) {
	t.Parallel()

	site := loadertest.Sample(t)
	root := loadertest.Package(t, "")

	testCases := []struct {
		name   string
		pkg    *types.Package
		symbol string
		method string
		want   string
		wantOK bool
	}{
		{
			name:   "type",
			pkg:    root.TypesPkg,
			symbol: "Circle",
			method: "",
			want:   "Circle.html",
			wantOK: true,
		},
		{
			name:   "method",
			pkg:    root.TypesPkg,
			symbol: "Circle",
			method: "Area",
			want:   "Circle.Area.html",
			wantOK: true,
		},
		{
			name:   "imported package symbol",
			pkg:    importOf(t, root, "io"),
			symbol: "Writer",
			method: "",
			want:   "https://pkg.go.dev/io#Writer",
			wantOK: true,
		},
		{
			name:   "unknown symbol",
			pkg:    root.TypesPkg,
			symbol: "Missing",
			method: "",
			want:   "",
			wantOK: false,
		},
		{
			name:   "unknown method",
			pkg:    root.TypesPkg,
			symbol: "Circle",
			method: "Missing",
			want:   "",
			wantOK: false,
		},
		{
			name:   "method on non-type",
			pkg:    root.TypesPkg,
			symbol: "NewCircle",
			method: "Area",
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := link.New(site)

			// Act
			url, ok := sut.Lookup(rootIndex, tc.pkg, tc.symbol, tc.method)

			// Assert
			if got, want := ok, tc.wantOK; !cmp.Equal(got, want) {
				t.Fatalf("Resolver.Lookup(...) ok = %v, want %v", got, want)
			}
			if got, want := url, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Resolver.Lookup(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestResolver_PackageURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		importPath string
		want       string
	}{
		{
			name:       "local package",
			importPath: "example.com/sample/shapes",
			want:       "shapes/index.html",
		},
		{
			name:       "standard library package",
			importPath: "net/http",
			want:       "https://pkg.go.dev/net/http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := link.New(loadertest.Sample(t))

			// Act
			url := sut.PackageURL(rootIndex, tc.importPath)

			// Assert
			if got, want := url, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Resolver.PackageURL(...) = %q, want %q", got, want)
			}
		})
	}
}

func importOf(t testing.TB, p *model.Package, path string) *types.Package {
	t.Helper()
	for _, imp := range p.TypesPkg.Imports() {
		if imp.Path() == path {
			return imp
		}
	}
	t.Fatalf("importOf(%q): not imported by %s", path, p.ImportPath)
	return nil
}

func fieldOf(t testing.TB, typ *model.Type, name string) *types.Var {
	t.Helper()
	st, ok := typ.Obj.Type().Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("fieldOf(%q): %s is not a struct", name, typ.Name)
	}
	for f := range st.Fields() {
		if f.Name() == name {
			return f
		}
	}
	t.Fatalf("fieldOf(%q): not found on %s", name, typ.Name)
	return nil
}
