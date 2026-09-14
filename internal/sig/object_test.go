package sig_test

import (
	"go/types"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/sig"
)

func importedType(t testing.TB, p *model.Package, path, name string) *types.TypeName {
	t.Helper()
	for _, imp := range p.TypesPkg.Imports() {
		if imp.Path() != path {
			continue
		}
		obj, ok := imp.Scope().Lookup(name).(*types.TypeName)
		if !ok {
			t.Fatalf("importedType(%q, %q): not a type", path, name)
		}
		return obj
	}
	t.Fatalf("importedType(%q, %q): package not imported", path, name)
	return nil
}

func TestPrinter_Object(t *testing.T) {
	t.Parallel()

	root := loadertest.Package(t, "")

	testCases := []struct {
		name string
		obj  *types.TypeName
		want sig.Rendered
	}{
		{
			name: "standard library interface",
			obj:  importedType(t, root, "io", "Writer"),
			want: sig.Rendered{
				Text: "type Writer interface {\n" +
					"\tWrite(p []byte) (n int, err error)\n" +
					"}",
				Links: []sig.Span{
					{Start: 35, End: 39, URL: builtin + "byte"},
					{Start: 44, End: 47, URL: builtin + "int"},
					{Start: 53, End: 58, URL: builtin + "error"},
				},
			},
		},
		{
			name: "local struct from type information",
			obj:  loadertest.Type(t, "", "Circle").Obj,
			want: sig.Rendered{
				Text: "type Circle struct {\n" +
					"\tRadius float64\n" +
					"\n" +
					"\tColor Color\n" +
					"\n" +
					"\t// contains unexported fields\n" +
					"}",
				Links: []sig.Span{
					{Start: 29, End: 36, URL: builtin + "float64"},
					{Start: 45, End: 50, URL: "Color.html"},
				},
			},
		},
		{
			name: "sibling package interface qualifies foreign types",
			obj:  loadertest.Type(t, "shapes", "Sizer").Obj,
			want: sig.Rendered{
				Text: "type Sizer interface {\n" +
					"\tArea() float64\n" +
					"}",
				Links: []sig.Span{
					{Start: 31, End: 38, URL: builtin + "float64"},
				},
			},
		},
		{
			name: "defined basic type",
			obj:  loadertest.Type(t, "", "Color").Obj,
			want: sig.Rendered{
				Text: "type Color int",
				Links: []sig.Span{
					{Start: 11, End: 14, URL: builtin + "int"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newPrinter(t)

			// Act
			rendered := sut.Object(tc.obj)

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Printer.Object(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
