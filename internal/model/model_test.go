package model_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/godoc2/internal/model"
)

func TestSite_Module(t *testing.T) {
	t.Parallel()

	alpha := &model.Module{Path: "example.com/alpha"}
	beta := &model.Module{Path: "example.com/beta"}

	testCases := []struct {
		name string
		path string
		want *model.Module
	}{
		{
			name: "first module",
			path: "example.com/alpha",
			want: alpha,
		},
		{
			name: "second module",
			path: "example.com/beta",
			want: beta,
		},
		{
			name: "unknown module",
			path: "example.com/gamma",
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &model.Site{Modules: []*model.Module{alpha, beta}}

			// Act
			mod := sut.Module(tc.path)

			// Assert
			if got, want := mod, tc.want; got != want {
				t.Errorf("Site.Module(%q) = %v, want %v", tc.path, got, want)
			}
		})
	}
}

func TestSite_Packages_ReturnsPackagesInModuleOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	a := &model.Package{ImportPath: "example.com/alpha"}
	b := &model.Package{ImportPath: "example.com/beta"}
	c := &model.Package{ImportPath: "example.com/beta/lib"}
	sut := &model.Site{Modules: []*model.Module{
		{Path: "example.com/alpha", Packages: []*model.Package{a}},
		{Path: "example.com/beta", Packages: []*model.Package{b, c}},
	}}

	// Act
	pkgs := sut.Packages()

	// Assert
	if got, want := pkgs, []*model.Package{a, b, c}; !cmp.Equal(got, want) {
		t.Errorf("Site.Packages() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestModule_Root(t *testing.T) {
	t.Parallel()

	root := &model.Package{ImportPath: "example.com/mod", RelPath: ""}
	nested := &model.Package{ImportPath: "example.com/mod/lib", RelPath: "lib"}

	testCases := []struct {
		name     string
		packages []*model.Package
		want     *model.Package
	}{
		{
			name:     "root package present",
			packages: []*model.Package{root, nested},
			want:     root,
		},
		{
			name:     "no root package",
			packages: []*model.Package{nested},
			want:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &model.Module{Path: "example.com/mod", Packages: tc.packages}

			// Act
			p := sut.Root()

			// Assert
			if got, want := p, tc.want; got != want {
				t.Errorf("Module.Root() = %v, want %v", got, want)
			}
		})
	}
}

func TestModule_Deprecated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		packages []*model.Package
		want     string
	}{
		{
			name:     "deprecated root package",
			packages: []*model.Package{{RelPath: "", Doc: "Package mod is old.\n\nDeprecated: Use next.\n"}},
			want:     "Use next.",
		},
		{
			name:     "root package in use",
			packages: []*model.Package{{RelPath: "", Doc: "Package mod is current.\n"}},
			want:     "",
		},
		{
			name:     "no root package",
			packages: []*model.Package{{RelPath: "lib", Doc: "Deprecated: Gone.\n"}},
			want:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &model.Module{Path: "example.com/mod", Packages: tc.packages}

			// Act
			msg := sut.Deprecated()

			// Assert
			if got, want := msg, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Module.Deprecated() = %q, want %q", got, want)
			}
		})
	}
}

func TestDeprecation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "not deprecated",
			doc:  "Circle is a round shape.\n",
			want: "",
		},
		{
			name: "own paragraph",
			doc:  "Legacy is the old version.\n\nDeprecated: Use Version instead.\n",
			want: "Use Version instead.",
		},
		{
			name: "multi-line message",
			doc:  "Legacy is the old version.\n\nDeprecated: Use Version instead. Legacy is kept\nonly for old callers.\n",
			want: "Use Version instead. Legacy is kept only for old callers.",
		},
		{
			name: "first paragraph",
			doc:  "Deprecated: Gone.\n\nMore text.\n",
			want: "Gone.",
		},
		{
			name: "prefix inside a paragraph",
			doc:  "Legacy is old. Deprecated: not really.\n",
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

			// Act
			msg := model.Deprecation(tc.doc)

			// Assert
			if got, want := msg, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Deprecation(...) = %q, want %q", got, want)
			}
		})
	}
}
