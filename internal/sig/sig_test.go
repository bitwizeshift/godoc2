package sig_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/sig"
)

const (
	rootIndex = "example.com/sample/index.html"
	builtin   = "https://pkg.go.dev/builtin#"
)

func newPrinter(t testing.TB) *sig.Printer {
	t.Helper()
	site := loadertest.Sample(t)
	return &sig.Printer{
		Resolver: link.New(site),
		Pkg:      loadertest.Package(t, ""),
		From:     rootIndex,
	}
}

func TestPrinter_Func(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		fn   *model.Func
		want sig.Rendered
	}{
		{
			name: "constructor",
			fn:   loadertest.Func(t, "", "NewCircle"),
			want: sig.Rendered{
				Text: "func NewCircle(radius float64) *Circle",
				Links: []sig.Span{
					{Start: 22, End: 29, URL: builtin + "float64"},
					{Start: 32, End: 38, URL: "Circle.html"},
				},
			},
		},
		{
			name: "pointer method",
			fn:   loadertest.Method(t, "", "Circle", "Area"),
			want: sig.Rendered{
				Text: "func (c *Circle) Area() float64",
				Links: []sig.Span{
					{Start: 9, End: 15, URL: "Circle.html"},
					{Start: 24, End: 31, URL: builtin + "float64"},
				},
			},
		},
		{
			name: "interface method",
			fn:   loadertest.Method(t, "", "Shape", "Area"),
			want: sig.Rendered{
				Text: "func (Shape) Area() float64",
				Links: []sig.Span{
					{Start: 6, End: 11, URL: "Shape.html"},
					{Start: 20, End: 27, URL: builtin + "float64"},
				},
			},
		},
		{
			name: "external parameter and error result",
			fn:   loadertest.Func(t, "", "Describe"),
			want: sig.Rendered{
				Text: "func Describe(w io.Writer, c *Circle) error",
				Links: []sig.Span{
					{Start: 16, End: 25, URL: "https://pkg.go.dev/io#Writer"},
					{Start: 30, End: 36, URL: "Circle.html"},
					{Start: 38, End: 43, URL: builtin + "error"},
				},
			},
		},
		{
			name: "generic method with tuple result",
			fn:   loadertest.Method(t, "", "Stack", "Pop"),
			want: sig.Rendered{
				Text: "func (s *Stack[T]) Pop() (T, bool)",
				Links: []sig.Span{
					{Start: 9, End: 14, URL: "Stack.html"},
					{Start: 29, End: 33, URL: builtin + "bool"},
				},
			},
		},
		{
			name: "long signature",
			fn:   loadertest.Func(t, "", "Configure"),
			want: sig.Rendered{
				Text: "func Configure(\n" +
					"\tradius float64,\n" +
					"\tcolor Color,\n" +
					"\tname string,\n" +
					"\twriter io.Writer,\n" +
					"\textra map[string][]int,\n" +
					") (*Circle, error)",
				Links: []sig.Span{
					{Start: 24, End: 31, URL: builtin + "float64"},
					{Start: 40, End: 45, URL: "Color.html"},
					{Start: 53, End: 59, URL: builtin + "string"},
					{Start: 69, End: 78, URL: "https://pkg.go.dev/io#Writer"},
					{Start: 91, End: 97, URL: builtin + "string"},
					{Start: 100, End: 103, URL: builtin + "int"},
					{Start: 109, End: 115, URL: "Circle.html"},
					{Start: 117, End: 122, URL: builtin + "error"},
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
			rendered := sut.Func(tc.fn)

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Printer.Func(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestPrinter_Func_WithSelf_LinksName(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newPrinter(t)
	sut.Self = "NewCircle.html"
	want := sig.Rendered{
		Text: "func NewCircle(radius float64) *Circle",
		Links: []sig.Span{
			{Start: 5, End: 14, URL: "NewCircle.html"},
			{Start: 22, End: 29, URL: builtin + "float64"},
			{Start: 32, End: 38, URL: "Circle.html"},
		},
	}

	// Act
	rendered := sut.Func(loadertest.Func(t, "", "NewCircle"))

	// Assert
	if got, want := rendered, want; !cmp.Equal(got, want) {
		t.Errorf("Printer.Func(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestPrinter_Type(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want sig.Rendered
	}{
		{
			name: "struct with unexported field",
			typ:  loadertest.Type(t, "", "Circle"),
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
			name: "sealed interface",
			typ:  loadertest.Type(t, "", "Shape"),
			want: sig.Rendered{
				Text: "type Shape interface {\n" +
					"\tArea() float64\n" +
					"\n" +
					"\t// contains unexported methods\n" +
					"}",
				Links: []sig.Span{
					{Start: 24, End: 28, URL: "Shape.Area.html"},
					{Start: 31, End: 38, URL: builtin + "float64"},
				},
			},
		},
		{
			name: "open interface",
			typ:  loadertest.Type(t, "", "Named"),
			want: sig.Rendered{
				Text: "type Named interface {\n" +
					"\tName() string\n" +
					"}",
				Links: []sig.Span{
					{Start: 24, End: 28, URL: "Named.Name.html"},
					{Start: 31, End: 37, URL: builtin + "string"},
				},
			},
		},
		{
			name: "alias",
			typ:  loadertest.Type(t, "", "ID"),
			want: sig.Rendered{
				Text: "type ID = string",
				Links: []sig.Span{
					{Start: 10, End: 16, URL: builtin + "string"},
				},
			},
		},
		{
			name: "defined basic type",
			typ:  loadertest.Type(t, "", "Color"),
			want: sig.Rendered{
				Text: "type Color int",
				Links: []sig.Span{
					{Start: 11, End: 14, URL: builtin + "int"},
				},
			},
		},
		{
			name: "generic struct with only unexported fields",
			typ:  loadertest.Type(t, "", "Stack"),
			want: sig.Rendered{
				Text: "type Stack[T any] struct {\n" +
					"\t// contains unexported fields\n" +
					"}",
				Links: []sig.Span{
					{Start: 13, End: 16, URL: builtin + "any"},
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
			rendered := sut.Type(tc.typ)

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Printer.Type(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestPrinter_Value(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value *model.Value
		want  sig.Rendered
	}{
		{
			name:  "iota constant",
			value: loadertest.Value(t, "", "Red"),
			want: sig.Rendered{
				Text: "const Red Color = 0",
				Links: []sig.Span{
					{Start: 10, End: 15, URL: "Color.html"},
				},
			},
		},
		{
			name:  "implicit repetition",
			value: loadertest.Value(t, "", "Green"),
			want: sig.Rendered{
				Text: "const Green Color = 1",
				Links: []sig.Span{
					{Start: 12, End: 17, URL: "Color.html"},
				},
			},
		},
		{
			name:  "untyped string constant",
			value: loadertest.Value(t, "", "Version"),
			want: sig.Rendered{
				Text: `const Version = "1.0"`,
			},
		},
		{
			name:  "variable with inferred type",
			value: loadertest.Value(t, "", "DefaultColor"),
			want: sig.Rendered{
				Text: "var DefaultColor Color = Green",
				Links: []sig.Span{
					{Start: 17, End: 22, URL: "Color.html"},
					{Start: 25, End: 30, URL: "Green.html"},
				},
			},
		},
		{
			name:  "variable with call initializer",
			value: loadertest.Value(t, "", "ErrNegative"),
			want: sig.Rendered{
				Text: `var ErrNegative error = fmt.Errorf("negative size")`,
				Links: []sig.Span{
					{Start: 16, End: 21, URL: builtin + "error"},
					{Start: 24, End: 34, URL: "https://pkg.go.dev/fmt#Errorf"},
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
			rendered := sut.Value(tc.value)

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Printer.Value(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestPrinter_Field(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		field *model.Field
		want  sig.Rendered
	}{
		{
			name:  "embedded field",
			field: loadertest.Field(t, "", "Label", "Reader"),
			want: sig.Rendered{
				Text: "io.Reader",
				Links: []sig.Span{
					{Start: 0, End: 9, URL: "https://pkg.go.dev/io#Reader"},
				},
			},
		},
		{
			name:  "named field",
			field: loadertest.Field(t, "", "Label", "Text"),
			want: sig.Rendered{
				Text: "Text string",
				Links: []sig.Span{
					{Start: 5, End: 11, URL: builtin + "string"},
				},
			},
		},
		{
			name:  "second name of a shared declaration",
			field: loadertest.Field(t, "", "Label", "Height"),
			want: sig.Rendered{
				Text: "Height int",
				Links: []sig.Span{
					{Start: 7, End: 10, URL: builtin + "int"},
				},
			},
		},
		{
			name:  "field with tag",
			field: loadertest.Field(t, "", "Label", "Font"),
			want: sig.Rendered{
				Text: "Font string `json:\"font\"`",
				Links: []sig.Span{
					{Start: 5, End: 11, URL: builtin + "string"},
				},
			},
		},
		{
			name:  "field of local type",
			field: loadertest.Field(t, "", "Circle", "Color"),
			want: sig.Rendered{
				Text: "Color Color",
				Links: []sig.Span{
					{Start: 6, End: 11, URL: "Color.html"},
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
			rendered := sut.Field(tc.field)

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Printer.Field(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
