package generate_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/emit"
	"github.com/bitwizeshift/godoc2/internal/emit/emittest"
	"github.com/bitwizeshift/godoc2/internal/generate"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/progress"
)

// siteLoader is a [generate.Loader] that returns a fixed site.
type siteLoader struct {
	site *model.Site
}

func (l siteLoader) Load(context.Context) (*model.Site, error) {
	return l.site, nil
}

// errLoader is a [generate.Loader] that always fails.
type errLoader struct {
	err error
}

func (l errLoader) Load(context.Context) (*model.Site, error) {
	return nil, l.err
}

// stageRecorder is a [progress.Reporter] that records the stage names.
type stageRecorder struct {
	stages []string
	files  []string
}

func (r *stageRecorder) Stage(name string) { r.stages = append(r.stages, name) }
func (r *stageRecorder) File(path string)  { r.files = append(r.files, path) }

var _ progress.Reporter = (*stageRecorder)(nil)

func TestGenerator_Generate_WithFixture_WritesEveryPage(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sink := emittest.NewMemSink()
	reporter := &stageRecorder{}
	sut := &generate.Generator{
		Loader:   siteLoader{site: loadertest.Sample(t)},
		Sink:     sink,
		Reporter: reporter,
	}
	wantPaths := []string{
		"example.com/sample.html",
		"example.com/sample/Big.html",
		"example.com/sample/Blue.html",
		"example.com/sample/Circle.Area.html",
		"example.com/sample/Circle.Name.html",
		"example.com/sample/Circle.html",
		"example.com/sample/Color.html",
		"example.com/sample/Configure.html",
		"example.com/sample/Counter.Add.html",
		"example.com/sample/Counter.html",
		"example.com/sample/DefaultColor.html",
		"example.com/sample/Describe.html",
		"example.com/sample/ErrNegative.html",
		"example.com/sample/Green.html",
		"example.com/sample/Grid.html",
		"example.com/sample/ID.html",
		"example.com/sample/Label.html",
		"example.com/sample/Legacy.html",
		"example.com/sample/Named.html",
		"example.com/sample/Namer.html",
		"example.com/sample/NewCircle.html",
		"example.com/sample/Red.html",
		"example.com/sample/Shape.html",
		"example.com/sample/Square.Name.html",
		"example.com/sample/Square.String.html",
		"example.com/sample/Square.html",
		"example.com/sample/Stack.Pop.html",
		"example.com/sample/Stack.Push.html",
		"example.com/sample/Stack.html",
		"example.com/sample/Version.html",
		"example.com/sample/ZeroCircle.html",
		"example.com/sample/cmd/tool/index.html",
		"example.com/sample/cmd/tool/main.go.html",
		"example.com/sample/doc.go.html",
		"example.com/sample/empty/empty.go.html",
		"example.com/sample/empty/index.html",
		"example.com/sample/index.html",
		"example.com/sample/indexed/Page.html",
		"example.com/sample/indexed/index.html",
		"example.com/sample/indexed/indexed.go.html",
		"example.com/sample/internal/secret/Token.html",
		"example.com/sample/internal/secret/index.html",
		"example.com/sample/internal/secret/secret.go.html",
		"example.com/sample/readme/Note.html",
		"example.com/sample/readme/docs/diagram.svg",
		"example.com/sample/readme/docs/guide.md",
		"example.com/sample/readme/index.html",
		"example.com/sample/readme/readme.go.html",
		"example.com/sample/sample.go.html",
		"example.com/sample/sample_test.go.html",
		"example.com/sample/shapes/NewUnit.html",
		"example.com/sample/shapes/Nothing.html",
		"example.com/sample/shapes/Perimeter.html",
		"example.com/sample/shapes/Sizer.html",
		"example.com/sample/shapes/factory.go.html",
		"example.com/sample/shapes/index.html",
		"example.com/sample/shapes/shapes.go.html",
		"index.html",
		"static/godoc2.css",
		"static/godoc2.js",
		"static/search-index.js",
	}
	wantStages := []string{
		"Loading packages",
		"Indexing",
		"Writing static files",
		"Writing root page",
		"Writing module example.com/sample",
		"Writing package example.com/sample",
		"Writing package example.com/sample/cmd/tool",
		"Writing package example.com/sample/empty",
		"Writing package example.com/sample/indexed",
		"Writing package example.com/sample/internal/secret",
		"Writing package example.com/sample/readme",
		"Writing package example.com/sample/shapes",
		"Copying linked files",
		"Writing search index",
	}

	// Act
	err := sut.Generate(ctx)
	index, _ := sink.Content("static/search-index.js")
	guide, _ := sink.Content("example.com/sample/readme/docs/guide.md")

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Generator.Generate(...) = %v, want nil", got)
	}
	if got, want := sink.Paths(), wantPaths; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) paths mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := reporter.stages, wantStages; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) stages mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := reporter.files, wantPaths; !cmp.Equal(got, want, cmpopts.SortSlices(strings.Compare)) {
		t.Errorf("Generator.Generate(...) reported files mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpopts.SortSlices(strings.Compare)))
	}
	if got, want := strings.Contains(index, `{"name":"Circle.Area","kind":"method","package":"example.com/sample","path":"example.com/sample/Circle.Area.html"}`), true; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) search index has method entry = %v, want %v", got, want)
	}
	if got, want := guide, "# Guide\n\nThe guide is copied next to the page.\n"; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) copied file mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestGenerator_Generate_WithSeveralModules_WritesRootPage(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sink := emittest.NewMemSink()
	reporter := &stageRecorder{}
	sut := &generate.Generator{
		Loader:   siteLoader{site: loadertest.Multi(t)},
		Sink:     sink,
		Reporter: reporter,
	}
	wantPaths := []string{
		"example.com/multi.html",
		"example.com/multi/Name.html",
		"example.com/multi/NewUnit.html",
		"example.com/multi/Root.Name.html",
		"example.com/multi/Root.html",
		"example.com/multi/alpha.html",
		"example.com/multi/alpha/Namer.html",
		"example.com/multi/alpha/Unit.Name.html",
		"example.com/multi/alpha/Unit.html",
		"example.com/multi/alpha/alpha.go.html",
		"example.com/multi/alpha/index.html",
		"example.com/multi/alpha/inner/Inner.Name.html",
		"example.com/multi/alpha/inner/Inner.html",
		"example.com/multi/alpha/inner/NewUnit.html",
		"example.com/multi/alpha/inner/index.html",
		"example.com/multi/alpha/inner/inner.go.html",
		"example.com/multi/beta.html",
		"example.com/multi/beta/lib/New.html",
		"example.com/multi/beta/lib/NewUnit.html",
		"example.com/multi/beta/lib/Wrapper.Name.html",
		"example.com/multi/beta/lib/Wrapper.html",
		"example.com/multi/beta/lib/index.html",
		"example.com/multi/beta/lib/lib.go.html",
		"example.com/multi/docs/notes.md",
		"example.com/multi/index.html",
		"example.com/multi/multi.go.html",
		"index.html",
		"static/godoc2.css",
		"static/godoc2.js",
		"static/search-index.js",
	}
	wantStages := []string{
		"Loading packages",
		"Indexing",
		"Writing static files",
		"Writing root page",
		"Writing module example.com/multi",
		"Writing package example.com/multi",
		"Writing module example.com/multi/alpha",
		"Writing package example.com/multi/alpha",
		"Writing package example.com/multi/alpha/inner",
		"Writing module example.com/multi/beta",
		"Writing package example.com/multi/beta/lib",
		"Copying linked files",
		"Writing search index",
	}

	// Act
	err := sut.Generate(ctx)
	root, _ := sink.Content("index.html")
	notes, _ := sink.Content("example.com/multi/docs/notes.md")

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Generator.Generate(...) = %v, want nil", got)
	}
	if got, want := sink.Paths(), wantPaths; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) paths mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := reporter.stages, wantStages; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) stages mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	if got, want := strings.Contains(root, `<h1>Modules</h1>`), true; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) root page lists modules = %v, want %v", got, want)
	}
	if got, want := notes, "# Notes\n\nThe notes are copied next to the root page.\n"; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) copied file mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestGenerator_Generate_WithOneModule_RedirectsRootPage(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sink := emittest.NewMemSink()
	sut := &generate.Generator{
		Loader: siteLoader{site: loadertest.Sample(t)},
		Sink:   sink,
	}

	// Act
	err := sut.Generate(ctx)
	root, _ := sink.Content("index.html")

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Generator.Generate(...) = %v, want nil", got)
	}
	if got, want := strings.Contains(root, `<meta http-equiv="refresh" content="0; url=example.com/sample.html">`), true; !cmp.Equal(got, want) {
		t.Errorf("Generator.Generate(...) root page redirects = %v, want %v", got, want)
	}
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	testErr := errors.New("boom")

	testCases := []struct {
		name    string
		loader  generate.Loader
		sink    emit.Sink
		wantErr error
	}{
		{
			name:    "loader fails",
			loader:  errLoader{err: testErr},
			sink:    emittest.NewMemSink(),
			wantErr: testErr,
		},
		{
			name:    "sink fails",
			loader:  siteLoader{site: loadertest.Sample(t)},
			sink:    emittest.ErrSink(testErr),
			wantErr: testErr,
		},
		{
			name:    "wraps generate sentinel",
			loader:  errLoader{err: testErr},
			sink:    emittest.NewMemSink(),
			wantErr: generate.ErrGenerate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			sut := &generate.Generator{Loader: tc.loader, Sink: tc.sink}

			// Act
			err := sut.Generate(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Generator.Generate(...) = %v, want %v", got, want)
			}
		})
	}
}
