package relate_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/relate"
)

// implSummary is the comparable projection of a [relate.Impl].
type implSummary struct {
	Name    string
	Pkg     string
	Pointer bool
	Local   bool
}

func summarize(impls []relate.Impl) []implSummary {
	var result []implSummary
	for _, impl := range impls {
		result = append(result, implSummary{
			Name:    impl.Obj.Name(),
			Pkg:     impl.Obj.Pkg().Path(),
			Pointer: impl.Pointer,
			Local:   impl.Local != nil,
		})
	}
	return result
}

func valueNames(vals []*model.Value) []string {
	var names []string
	for _, v := range vals {
		names = append(names, v.Name)
	}
	return names
}

func funcNames(fns []*model.Func) []string {
	var names []string
	for _, f := range fns {
		names = append(names, f.Name)
	}
	return names
}

func TestIndex_Instances(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []string
	}{
		{
			name: "enum values",
			typ:  loadertest.Type(t, "", "Color"),
			want: []string{"Blue", "Green", "Red", "DefaultColor"},
		},
		{
			name: "no instances",
			typ:  loadertest.Type(t, "", "Circle"),
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := relate.New(loadertest.Sample(t))

			// Act
			instances := sut.Instances(tc.typ)

			// Assert
			if got, want := valueNames(instances), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Index.Instances(...) = %v, want %v", got, want)
			}
		})
	}
}

// funcGroupSummary is the comparable projection of a [relate.FuncGroup].
type funcGroupSummary struct {
	Path  string
	Funcs []string
}

func summarizeFuncGroups(groups []relate.FuncGroup) []funcGroupSummary {
	var result []funcGroupSummary
	for _, g := range groups {
		result = append(result, funcGroupSummary{Path: g.Path, Funcs: funcNames(g.Funcs)})
	}
	return result
}

func TestIndex_Constructors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []funcGroupSummary
	}{
		{
			name: "same package first then other packages",
			typ:  loadertest.Type(t, "", "Circle"),
			want: []funcGroupSummary{
				{Path: "", Funcs: []string{"Configure", "NewCircle", "ZeroCircle"}},
				{Path: "example.com/sample/shapes", Funcs: []string{"NewUnit"}},
			},
		},
		{
			name: "interface includes functions that return an implementation",
			typ:  loadertest.Type(t, "", "Named"),
			want: []funcGroupSummary{
				{Path: "", Funcs: []string{"Configure", "NewCircle"}},
				{Path: "example.com/sample/shapes", Funcs: []string{"NewUnit"}},
			},
		},
		{
			name: "none",
			typ:  loadertest.Type(t, "", "Square"),
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := relate.New(loadertest.Sample(t))

			// Act
			ctors := sut.Constructors(tc.typ)

			// Assert
			if got, want := summarizeFuncGroups(ctors), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Index.Constructors(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestIndex_Constructors_WithSeveralModules_OrdersByDistance(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := relate.New(loadertest.Multi(t))
	unit := loadertest.MultiType(t, loadertest.MultiAlphaPath, "", "Unit")
	want := []funcGroupSummary{
		{Path: "example.com/multi/alpha/inner", Funcs: []string{"NewUnit"}},
		{Path: "example.com/multi", Funcs: []string{"NewUnit"}},
		{Path: "example.com/multi/beta/lib", Funcs: []string{"NewUnit"}},
	}

	// Act
	ctors := sut.Constructors(unit)

	// Assert
	if got, want := summarizeFuncGroups(ctors), want; !cmp.Equal(got, want) {
		t.Errorf("Index.Constructors(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestIndex_Utilities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []funcGroupSummary
	}{
		{
			name: "same package first then other packages",
			typ:  loadertest.Type(t, "", "Circle"),
			want: []funcGroupSummary{
				{Path: "", Funcs: []string{"Describe"}},
				{Path: "example.com/sample/shapes", Funcs: []string{"Perimeter"}},
			},
		},
		{
			name: "parameter of constructor is not a utility",
			typ:  loadertest.Type(t, "", "Color"),
			want: []funcGroupSummary{
				{Path: "", Funcs: []string{"Configure"}},
			},
		},
		{
			name: "none",
			typ:  loadertest.Type(t, "", "Big"),
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := relate.New(loadertest.Sample(t))

			// Act
			utils := sut.Utilities(tc.typ)

			// Assert
			if got, want := summarizeFuncGroups(utils), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Index.Utilities(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestIndex_Implements(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []implSummary
	}{
		{
			name: "pointer receivers",
			typ:  loadertest.Type(t, "", "Circle"),
			want: []implSummary{
				{Name: "Named", Pkg: "example.com/sample", Pointer: true, Local: true},
				{Name: "Shape", Pkg: "example.com/sample", Pointer: true, Local: true},
				{Name: "Sizer", Pkg: "example.com/sample/shapes", Pointer: true, Local: true},
			},
		},
		{
			name: "value receivers",
			typ:  loadertest.Type(t, "", "Square"),
			want: []implSummary{
				{Name: "Named", Pkg: "example.com/sample", Pointer: false, Local: true},
				{Name: "Stringer", Pkg: "fmt", Pointer: false, Local: false},
			},
		},
		{
			name: "interface is not checked",
			typ:  loadertest.Type(t, "", "Shape"),
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := relate.New(loadertest.Sample(t))

			// Act
			impls := sut.Implements(tc.typ)

			// Assert
			if got, want := summarize(impls), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty(), sortImpls()) {
				t.Errorf("Index.Implements(...) mismatch (-want +got):\n%s", cmp.Diff(want, got, sortImpls()))
			}
		})
	}
}

func TestIndex_Implementations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  *model.Type
		want []implSummary
	}{
		{
			name: "local interface",
			typ:  loadertest.Type(t, "", "Named"),
			want: []implSummary{
				{Name: "Circle", Pkg: "example.com/sample", Pointer: true, Local: true},
				{Name: "Square", Pkg: "example.com/sample", Pointer: false, Local: true},
				{Name: "File", Pkg: "os", Pointer: true, Local: false},
			},
		},
		{
			name: "sibling package interface",
			typ:  loadertest.Type(t, "shapes", "Sizer"),
			want: []implSummary{
				{Name: "Circle", Pkg: "example.com/sample", Pointer: true, Local: true},
			},
		},
		{
			name: "empty interface",
			typ:  loadertest.Type(t, "shapes", "Nothing"),
			want: nil,
		},
		{
			name: "struct is not an interface",
			typ:  loadertest.Type(t, "", "Circle"),
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := relate.New(loadertest.Sample(t))

			// Act
			impls := sut.Implementations(tc.typ)

			// Assert
			if got, want := summarize(impls), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty(), sortImpls()) {
				t.Errorf("Index.Implementations(...) mismatch (-want +got):\n%s", cmp.Diff(want, got, sortImpls()))
			}
		})
	}
}

func TestGroups_PutsSamePackageFirst(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := relate.New(loadertest.Sample(t))
	circle := loadertest.Type(t, "", "Circle")
	impls := sut.Implements(circle)
	want := []groupSummary{
		{Path: "", Names: []string{"Named", "Shape"}},
		{Path: "example.com/sample/shapes", Names: []string{"Sizer"}},
	}

	// Act
	groups := relate.Groups(impls, circle)

	// Assert
	if got, want := summarizeGroups(groups), want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Groups(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestGroups_WithSeveralModules_OrdersByDistance(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := relate.New(loadertest.Multi(t))
	namer := loadertest.MultiType(t, loadertest.MultiAlphaPath, "", "Namer")
	impls := sut.Implementations(namer)
	want := []groupSummary{
		{Path: "", Names: []string{"Unit"}},
		{Path: "example.com/multi/alpha/inner", Names: []string{"Inner"}},
		{Path: "example.com/multi", Names: []string{"Root"}},
		{Path: "example.com/multi/beta/lib", Names: []string{"Wrapper"}},
	}

	// Act
	groups := relate.Groups(impls, namer)

	// Assert
	if got, want := summarizeGroups(groups), want; !cmp.Equal(got, want) {
		t.Errorf("Groups(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

type groupSummary struct {
	Path  string
	Names []string
}

func summarizeGroups(groups []relate.Group) []groupSummary {
	var result []groupSummary
	for _, g := range groups {
		s := groupSummary{Path: g.Path}
		for _, impl := range g.Impls {
			s.Names = append(s.Names, impl.Obj.Name())
		}
		result = append(result, s)
	}
	return result
}

func sortImpls() cmp.Option {
	return cmpopts.SortSlices(func(lhs, rhs implSummary) bool {
		if lhs.Pkg != rhs.Pkg {
			return lhs.Pkg < rhs.Pkg
		}
		return lhs.Name < rhs.Name
	})
}
