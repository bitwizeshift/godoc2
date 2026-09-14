package loader_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/loader"
	"github.com/bitwizeshift/godoc2/internal/model"
)

// packageSummary is the comparable projection of a [model.Package].
type packageSummary struct {
	ImportPath string
	Name       string
	RelPath    string
	Internal   bool
	Tool       bool
	Summary    string
	DocFile    string
	Consts     []string
	Vars       []string
	Types      map[string][]string
	Funcs      []string
	Examples   []string
	Files      []string
}

func summarize(pkgs []*model.Package) []packageSummary {
	var result []packageSummary
	for _, p := range pkgs {
		s := packageSummary{
			ImportPath: p.ImportPath,
			Name:       p.Name,
			RelPath:    p.RelPath,
			Internal:   p.Internal(),
			Tool:       p.Tool(),
			Summary:    p.Summary(),
			Types:      map[string][]string{},
		}
		if p.DocFile != nil {
			s.DocFile = filepath.Base(p.DocFile.Path)
		}
		for _, v := range p.Consts {
			s.Consts = append(s.Consts, v.Name)
		}
		for _, v := range p.Vars {
			s.Vars = append(s.Vars, v.Name)
		}
		for _, t := range p.Types {
			var methods []string
			for _, m := range t.Methods {
				methods = append(methods, m.Name)
			}
			s.Types[t.Kind.String()+" "+t.Name] = methods
		}
		for _, f := range p.Funcs {
			s.Funcs = append(s.Funcs, f.Name)
		}
		for _, e := range p.Examples {
			s.Examples = append(s.Examples, e.Name)
		}
		for _, f := range p.Files {
			s.Files = append(s.Files, f.Name)
		}
		result = append(result, s)
	}
	return result
}

func fixtureDir(t testing.TB) string {
	t.Helper()
	return testdataDir(t, "sample")
}

func testdataDir(t testing.TB, name string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("Abs(...) = %v, want nil", err)
	}
	return dir
}

func TestLoad_WithFixtureModule_ReturnsModule(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	cfg := loader.Config{Dir: fixtureDir(t), Patterns: []string{"./..."}}
	want := []packageSummary{
		{
			ImportPath: "example.com/sample",
			Name:       "sample",
			RelPath:    "",
			Summary:    "Package sample is a fixture module for godoc2 tests.",
			Consts:     []string{"Blue", "Green", "Red", "Version"},
			Vars:       []string{"DefaultColor", "ErrNegative"},
			Types: map[string][]string{
				"struct Big":      nil,
				"struct Circle":   {"Area", "Name"},
				"type Color":      nil,
				"struct Counter":  {"Add"},
				"struct Grid":     nil,
				"alias ID":        nil,
				"interface Named": nil,
				"interface Shape": nil,
				"struct Square":   {"Name", "String"},
				"struct Stack":    {"Pop", "Push"},
			},
			Funcs:    []string{"Configure", "Describe", "NewCircle"},
			Examples: []string{""},
			Files:    []string{"doc.go", "sample.go", "sample_test.go"},
		},
		{
			ImportPath: "example.com/sample/cmd/tool",
			Name:       "main",
			RelPath:    "cmd/tool",
			Tool:       true,
			Summary:    "Tool prints the fixture version.",
			Types:      map[string][]string{},
			Files:      []string{"main.go"},
		},
		{
			ImportPath: "example.com/sample/empty",
			Name:       "empty",
			RelPath:    "empty",
			Summary:    "Package empty has no exported identifiers.",
			Types:      map[string][]string{},
			Files:      []string{"empty.go"},
		},
		{
			ImportPath: "example.com/sample/indexed",
			Name:       "indexed",
			RelPath:    "indexed",
			DocFile:    "index.md",
			Types:      map[string][]string{"struct Page": nil},
			Files:      []string{"indexed.go"},
		},
		{
			ImportPath: "example.com/sample/internal/secret",
			Name:       "secret",
			RelPath:    "internal/secret",
			Internal:   true,
			Summary:    "Package secret holds internal types.",
			Types:      map[string][]string{"struct Token": nil},
			Files:      []string{"secret.go"},
		},
		{
			ImportPath: "example.com/sample/readme",
			Name:       "readme",
			RelPath:    "readme",
			DocFile:    "README.md",
			Types:      map[string][]string{"struct Note": nil},
			Files:      []string{"readme.go"},
		},
		{
			ImportPath: "example.com/sample/shapes",
			Name:       "shapes",
			RelPath:    "shapes",
			Summary:    "Package shapes declares interfaces satisfied by shapes.",
			Types: map[string][]string{
				"interface Nothing": nil,
				"interface Sizer":   nil,
			},
			Funcs: []string{"NewUnit", "Perimeter"},
			Files: []string{"factory.go", "shapes.go"},
		},
	}

	// Act
	mod, err := loader.Load(ctx, cfg)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Load(...) = %v, want nil", got)
	}
	if got, want := mod.Path, "example.com/sample"; !cmp.Equal(got, want) {
		t.Errorf("Load(...) Path = %q, want %q", got, want)
	}
	if got, want := summarize(mod.Packages), want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Load(...) packages mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := mod.DocFile, (*model.DocFile)(nil); !cmp.Equal(got, want) {
		t.Errorf("Load(...) DocFile = %v, want nil", got)
	}
}

func TestLoad_WithBareModule_AttachesRootDocFile(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	dir := testdataDir(t, "bare")
	cfg := loader.Config{Dir: dir, Patterns: []string{"./..."}}
	want := &model.DocFile{
		Path: filepath.Join(dir, "README.md"),
		Text: "# Bare\n\nModule bare has no package in its root directory.\n\nSee [lib](lib) and [its source](lib/lib.go).\n",
	}

	// Act
	mod, err := loader.Load(ctx, cfg)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Load(...) = %v, want nil", got)
	}
	if got, want := mod.DocFile, want; !cmp.Equal(got, want) {
		t.Errorf("Load(...) DocFile mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestLoad_WithFixtureModule_AssignsExamples(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	cfg := loader.Config{Dir: fixtureDir(t), Patterns: []string{"."}}
	want := map[string][]model.Example{
		"NewCircle": {{
			Name:   "NewCircle",
			Code:   "c := sample.NewCircle(1)\nfmt.Println(c.Name())\n",
			Output: "circle\n",
		}},
		"Circle.Area": {{
			Name:   "Circle_Area",
			Doc:    "ExampleCircle_Area shows the area of a unit circle.\n",
			Code:   "c := sample.NewCircle(1)\nfmt.Printf(\"%.2f\\n\", c.Area())\n",
			Output: "3.14\n",
		}},
		"Square.String": {{
			Name:   "Square_String_zero",
			Suffix: "zero",
			Code:   "fmt.Println(sample.Square{})\n",
			Output: "0\n",
		}},
	}

	// Act
	mod, err := loader.Load(ctx, cfg)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Load(...) = %v, want nil", got)
	}
	if got, want := examplesOf(mod.Packages[0]), want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Load(...) examples mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

// examplesOf collects the examples attached to functions and methods, keyed
// by "Func" or "Type.Method".
func examplesOf(p *model.Package) map[string][]model.Example {
	result := map[string][]model.Example{}
	add := func(key string, exs []*model.Example) {
		for _, ex := range exs {
			result[key] = append(result[key], *ex)
		}
	}
	for _, f := range p.Funcs {
		add(f.Name, f.Examples)
	}
	for _, t := range p.Types {
		for _, m := range t.Methods {
			add(t.Name+"."+m.Name, m.Examples)
		}
	}
	return result
}

func TestLoad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		patterns []string
		wantErr  error
	}{
		{
			name:     "no match",
			patterns: []string{"./nosuchdir/..."},
			wantErr:  loader.ErrLoad,
		},
		{
			name:     "standard library only",
			patterns: []string{"fmt"},
			wantErr:  loader.ErrNoModule,
		},
		{
			name:     "missing package",
			patterns: []string{"example.com/sample/missing"},
			wantErr:  loader.ErrLoad,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			cfg := loader.Config{Dir: fixtureDir(t), Patterns: tc.patterns}

			// Act
			mod, err := loader.Load(ctx, cfg)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Load(...) = %v, want %v", got, want)
			}
			if got, want := mod, (*model.Module)(nil); !cmp.Equal(got, want) {
				t.Errorf("Load(...) module = %v, want nil", got)
			}
		})
	}
}
