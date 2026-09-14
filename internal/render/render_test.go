package render_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/relate"
	"github.com/bitwizeshift/godoc2/internal/render"
)

func newRenderer(t testing.TB) *render.Renderer {
	t.Helper()
	mod := loadertest.Sample(t)
	r, err := render.New(mod, link.New(mod), relate.New(mod))
	if err != nil {
		t.Fatalf("render.New(...) = %v, want nil", err)
	}
	return r
}

var sectionPattern = regexp.MustCompile(`<section class="section" id="([a-z]+)">`)

// sectionIDs lists the ids of the content sections in document order.
func sectionIDs(html string) []string {
	var ids []string
	for _, m := range sectionPattern.FindAllStringSubmatch(html, -1) {
		ids = append(ids, m[1])
	}
	return ids
}

// fragments is the list of exact HTML fragments a page must contain.
type fragments []string

func missing(html string, want fragments) []string {
	var result []string
	for _, f := range want {
		if !strings.Contains(html, f) {
			result = append(result, f)
		}
	}
	return result
}

func TestRenderer_Module(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRenderer(t)
	root := loadertest.Package(t, "")
	var out strings.Builder
	wantSections := []string{"documentation", "tools", "packages", "examples", "constants", "variables", "types", "functions"}
	wantFragments := fragments{
		`<h1><span class="kind">package</span> sample</h1>`,
		`<a class="badge badge-module" href="index.html">module</a>`,
		`<td><a href="cmd/tool/index.html">tool</a></td>`,
		`<td><a href="internal/secret/index.html">internal/secret</a> <span class="badge badge-internal">internal</span></td>`,
		`<td><a href="shapes/index.html">shapes</a></td>`,
		`<summary><span class="dir">internal</span></summary>`,
		`<h1 id="usage">Usage</h1>`,
		`<li><a href="#usage">Usage</a></li>`,
		`<a href="Circle.html"><code>Circle</code></a>`,
		`<a href="shapes/Sizer.html"><code>shapes.Sizer</code></a>`,
		`<details class="item" id="type.Circle" open>`,
		`<a href="Circle.html"><span class="nx">Circle</span></a>`,
		`<a class="src" href="sample.go.html#L44">source</a>`,
		`<button type="button" class="read-more">Read more</button>`,
	}

	// Act
	err := sut.Module(&out, root)
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Module(...) = %v, want nil", got)
	}
	if got, want := sectionIDs(html), wantSections; !cmp.Equal(got, want) {
		t.Errorf("Renderer.Module(...) sections = %v, want %v", got, want)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Module(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderer_Package(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		pkg           *model.Package
		wantSections  []string
		wantFragments fragments
	}{
		{
			name:         "empty package",
			pkg:          loadertest.Package(t, "empty"),
			wantSections: []string{"documentation"},
			wantFragments: fragments{
				`<h1><span class="kind">package</span> empty</h1>`,
				`<p class="message" id="empty">This package has no exported identifiers.</p>`,
				`<span class="sep">/</span><a href="index.html">empty</a>`,
			},
		},
		{
			name:         "internal package",
			pkg:          loadertest.Package(t, "internal/secret"),
			wantSections: []string{"documentation", "types"},
			wantFragments: fragments{
				`<h1><span class="kind">package</span> secret <span class="badge badge-internal">internal</span></h1>`,
				`<span class="sep">/</span><span class="crumb">internal</span><span class="sep">/</span><a href="index.html">secret</a>`,
			},
		},
		{
			name:         "tool package",
			pkg:          loadertest.Package(t, "cmd/tool"),
			wantSections: []string{"documentation"},
			wantFragments: fragments{
				`<h1><span class="kind">binary</span> tool</h1>`,
				`<span class="crumb">cmd</span><span class="sep">/</span><a href="index.html">tool</a>`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newRenderer(t)
			var out strings.Builder

			// Act
			err := sut.Package(&out, tc.pkg)
			html := out.String()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Renderer.Package(...) = %v, want nil", got)
			}
			if got, want := sectionIDs(html), tc.wantSections; !cmp.Equal(got, want) {
				t.Errorf("Renderer.Package(...) sections = %v, want %v", got, want)
			}
			if got, want := missing(html, tc.wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Renderer.Package(...) missing fragments:\n%s", strings.Join(got, "\n"))
			}
		})
	}
}

func TestRenderer_Type(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typ           *model.Type
		wantSections  []string
		wantFragments fragments
	}{
		{
			name:         "struct",
			typ:          loadertest.Type(t, "", "Circle"),
			wantSections: []string{"documentation", "examples", "properties", "constructors", "methods", "utilities", "implements"},
			wantFragments: fragments{
				`<h1><span class="kind">struct</span> Circle</h1>`,
				`<a class="src" href="sample.go.html#L44">source</a>`,
				`// contains unexported fields`,
				`<span class="badge-label">size</span> <span class="badge-value">24 bytes</span>`,
				`<details class="item" id="ctor.NewCircle" open>`,
				`<summary><h3 class="relation-path">example.com/sample/shapes</h3></summary>
<details class="item" id="ctor.NewUnit" open>`,
				`<a href="shapes/NewUnit.html"><span class="nf">NewUnit</span></a>`,
				`<a class="src" href="shapes/factory.go.html#L6">source</a>`,
				`<details class="item" id="method.Area" open>`,
				`<details class="item" id="util.Describe" open>`,
				`<summary><h3 class="relation-path">example.com/sample/shapes</h3></summary>
<details class="item" id="util.Perimeter" open>`,
				`<a href="shapes/Perimeter.html"><span class="nf">Perimeter</span></a>`,
				`<li><a href="#implements.example.com/sample.Named">Named</a></li>`,
				`<span class="badge badge-receiver" title="implemented by *Circle"><code>*Circle</code></span>`,
				`<details class="group" open>
        <summary><h3 class="relation-path">example.com/sample/shapes</h3></summary>
<details class="item" id="implements.example.com/sample/shapes.Sizer" open>`,
				`<a href="shapes/Sizer.html"><span class="nx">Sizer</span></a>`,
				`<details class="example" id="example-Circle">`,
			},
		},
		{
			name:         "interface",
			typ:          loadertest.Type(t, "", "Named"),
			wantSections: []string{"documentation", "properties", "implementations"},
			wantFragments: fragments{
				`<h1><span class="kind">interface</span> Named</h1>`,
				`<li><a href="#implementations.example.com/sample.Circle">Circle</a></li>`,
				`<span class="badge badge-receiver" title="implemented by *Circle"><code>*Circle</code></span>`,
				`<span class="badge badge-receiver" title="implemented by Square"><code>Square</code></span>`,
				`<li><a href="#implementations.example.com/sample.Square">Square</a></li>`,
				`<summary><h3 class="relation-path">os</h3></summary>`,
				`<a href="https://pkg.go.dev/os#File"><span class="nx">File</span></a> <span class="kd">struct</span>`,
				`<span class="badge badge-receiver" title="implemented by *File"><code>*File</code></span>`,
			},
		},
		{
			name:         "sealed interface",
			typ:          loadertest.Type(t, "", "Shape"),
			wantSections: []string{"documentation", "properties", "implementations"},
			wantFragments: fragments{
				`<span class="badge-label">sealed</span>`,
				`// contains unexported methods`,
			},
		},
		{
			name:         "enum with instances",
			typ:          loadertest.Type(t, "", "Color"),
			wantSections: []string{"documentation", "properties", "instances", "utilities"},
			wantFragments: fragments{
				`<details class="item" id="instance.Red" open>`,
				`<details class="item" id="instance.DefaultColor" open>`,
			},
		},
		{
			name:         "internal type",
			typ:          loadertest.Type(t, "internal/secret", "Token"),
			wantSections: []string{"documentation", "properties"},
			wantFragments: fragments{
				`<h1><span class="kind">struct</span> Token <span class="badge badge-internal">internal</span></h1>`,
				`<span class="badge-label">internal</span>`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newRenderer(t)
			var out strings.Builder

			// Act
			err := sut.Type(&out, tc.typ)
			html := out.String()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Renderer.Type(...) = %v, want nil", got)
			}
			if got, want := sectionIDs(html), tc.wantSections; !cmp.Equal(got, want) {
				t.Errorf("Renderer.Type(...) sections = %v, want %v", got, want)
			}
			if got, want := missing(html, tc.wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Renderer.Type(...) missing fragments:\n%s", strings.Join(got, "\n"))
			}
		})
	}
}

func TestRenderer_Type_ExternalImplementationHasNoSourceLink(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRenderer(t)
	var out strings.Builder
	itemPattern := regexp.MustCompile(`(?s)<details class="item" id="implementations\.os\.File".*?</details>`)

	// Act
	err := sut.Type(&out, loadertest.Type(t, "", "Named"))
	item := itemPattern.FindString(out.String())

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Type(...) = %v, want nil", got)
	}
	if got, want := strings.Contains(item, `class="src"`), false; !cmp.Equal(got, want) {
		t.Errorf("Renderer.Type(...) external item has source link = %v, want %v", got, want)
	}
}

func TestRenderer_Func(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		fn            *model.Func
		wantSections  []string
		wantFragments fragments
	}{
		{
			name:         "function",
			fn:           loadertest.Func(t, "", "NewCircle"),
			wantSections: []string{"documentation", "examples"},
			wantFragments: fragments{
				`<h1><span class="kind">func</span> NewCircle</h1>`,
				`<span class="sep">.</span><span class="crumb">NewCircle</span>`,
				`<a class="src" href="sample.go.html#L55">source</a>`,
				`<details class="example" id="example-NewCircle">`,
				`<pre class="example-output">circle
</pre>`,
			},
		},
		{
			name:         "method",
			fn:           loadertest.Method(t, "", "Circle", "Area"),
			wantSections: []string{"documentation", "examples"},
			wantFragments: fragments{
				`<h1><span class="kind">func</span> Area</h1>`,
				`<span class="sep">.</span><a href="Circle.html">Circle</a><span class="sep">.</span><span class="crumb">Area</span>`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newRenderer(t)
			var out strings.Builder

			// Act
			err := sut.Func(&out, tc.fn)
			html := out.String()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Renderer.Func(...) = %v, want nil", got)
			}
			if got, want := sectionIDs(html), tc.wantSections; !cmp.Equal(got, want) {
				t.Errorf("Renderer.Func(...) sections = %v, want %v", got, want)
			}
			if got, want := missing(html, tc.wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Renderer.Func(...) missing fragments:\n%s", strings.Join(got, "\n"))
			}
		})
	}
}

func TestRenderer_Value_WritesDefinitionAndDoc(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRenderer(t)
	var out strings.Builder
	wantFragments := fragments{
		`<h1><span class="kind">const</span> Green</h1>`,
		`<span class="mi">1</span>`,
		`<a class="src" href="sample.go.html#L28">source</a>`,
		`<p>Colours of the palette.</p>`,
	}

	// Act
	err := sut.Value(&out, loadertest.Value(t, "", "Green"))
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Value(...) = %v, want nil", got)
	}
	if got, want := sectionIDs(html), []string{"documentation"}; !cmp.Equal(got, want) {
		t.Errorf("Renderer.Value(...) sections = %v, want %v", got, want)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Value(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderer_Source_WritesLinkableLines(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRenderer(t)
	pkg := loadertest.Package(t, "shapes")
	file := pkg.Files[1]
	src, readErr := os.ReadFile(filepath.Join(loadertest.SampleDir(), "shapes", "shapes.go"))
	if readErr != nil {
		t.Fatalf("ReadFile(...) = %v, want nil", readErr)
	}
	var out strings.Builder
	wantFragments := fragments{
		`<h1><span class="kind">file</span> shapes.go</h1>`,
		`<a href="index.html">shapes</a><span class="sep">/</span><span class="crumb">shapes.go</span>`,
		`id="L4"`,
		`<li><a href="shapes.go.html">shapes.go</a></li>`,
	}

	// Act
	err := sut.Source(&out, pkg, file, src)
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Source(...) = %v, want nil", got)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Source(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}
