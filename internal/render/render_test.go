package render_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/loader/loadertest"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/relate"
	"github.com/bitwizeshift/godoc2/internal/render"
)

func newRenderer(t testing.TB) *render.Renderer {
	t.Helper()
	return newRendererFor(t, loadertest.Sample(t))
}

func newRendererFor(t testing.TB, site *model.Site) *render.Renderer {
	t.Helper()
	r, err := render.New(site, link.New(site), relate.New(site), docfile.NewResolver(site))
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
	mod := loadertest.SampleModule(t)
	var out strings.Builder
	wantSections := []string{"documentation", "tools", "packages"}
	wantFragments := fragments{
		`<title>example.com/sample (module)</title>`,
		`<link rel="icon" type="image/svg+xml" href="../static/favicon.svg">`,
		`<h1><span class="kind">module</span> example.com/sample</h1>`,
		`<a class="badge badge-module" href="sample.html" title="example.com/sample">module</a>
    </nav>`,
		`<td><a href="sample/cmd/tool/index.html">tool</a></td>`,
		`<td><a href="sample/internal/secret/index.html">internal/secret</a> <span class="badge badge-internal" title="Importable only by packages rooted at the parent of the internal directory.">internal</span></td>`,
		`<td><a href="sample/shapes/index.html">shapes</a></td>`,
		`<td><a href="sample/empty/index.html">empty</a> <span class="badge badge-deprecated" title="Nothing lives here.">deprecated</span></td>`,
		`<td><a href="sample/readme/index.html">readme</a></td>`,
		`<td class="summary"><p>Package readme is documented by its README file.</p>`,
		`<td class="summary"><p>Package indexed is documented by index.md.</p>`,
		`<summary><a href="sample/index.html">sample</a></summary>`,
		`<summary><span class="dir">internal</span></summary>`,
		`<li><a href="sample/internal/secret/index.html">secret <span class="badge badge-internal" title="Importable only by packages rooted at the parent of the internal directory.">internal</span></a>
  </li>`,
		`<h1 id="usage">Usage</h1>`,
		`<li><a href="#usage">Usage</a></li>`,
		`<a href="sample/Circle.html"><code>Circle</code></a>`,
		`<a href="sample/shapes/Sizer.html"><code>shapes.Sizer</code></a>`,
	}

	// Act
	err := sut.Module(&out, mod)
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

func TestRenderer_Module_WithDocFile_ShowsDocFile(t *testing.T) {
	t.Parallel()

	// Arrange
	bare := loadertest.Bare(t)
	sut := newRendererFor(t, bare)
	var out strings.Builder
	wantSections := []string{"documentation", "packages"}
	wantFragments := fragments{
		`<h1><span class="kind">module</span> example.com/bare</h1>`,
		`<h1 id="bare">Bare</h1>`,
		`<summary><span class="dir">bare</span></summary>`,
		`<a href="bare/lib/index.html">lib</a>`,
		`<a href="bare/lib/lib.go.html">its source</a>`,
		`<td><a href="bare/lib/index.html">lib</a></td>`,
	}

	// Act
	err := sut.Module(&out, bare.Modules[0])
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

func TestRenderer_Root_WithOneModule_Redirects(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRenderer(t)
	var out strings.Builder
	wantFragments := fragments{
		`<meta http-equiv="refresh" content="0; url=example.com/sample.html">`,
		`<link rel="canonical" href="example.com/sample.html">`,
		`<title>example.com/sample</title>`,
		`<a href="example.com/sample.html">example.com/sample</a>`,
	}

	// Act
	err := sut.Root(&out)
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Root(...) = %v, want nil", got)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Root(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderer_Root_WithSeveralModules_ListsModules(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRendererFor(t, loadertest.Multi(t))
	var out strings.Builder
	wantSections := []string{"documentation", "modules"}
	wantFragments := fragments{
		`<title>Modules</title>`,
		`<a class="sidebar-logo" href="index.html" title="Modules" aria-label="Modules">`,
		`<a class="badge badge-module" href="index.html">workspace</a>`,
		`<h1>Modules</h1>`,
		`<h1 id="multi">Multi</h1>`,
		`<a href="example.com/multi/alpha/index.html">alpha</a>`,
		`<a href="example.com/multi/beta.html">beta</a>`,
		`<a href="example.com/multi/beta/lib/lib.go.html#L7">the wrapper</a>`,
		`<a href="example.com/multi/docs/notes.md">the notes</a>`,
		`<td><a href="example.com/multi.html">example.com/multi</a></td>
          <td class="summary"><p>Two modules live in this directory.</p>`,
		`<td><a href="example.com/multi/alpha.html">example.com/multi/alpha</a></td>
          <td class="summary"><p>Package alpha is the first module of the multi fixture.</p>`,
		`<td><a href="example.com/multi/beta.html">example.com/multi/beta</a></td>
          <td class="summary"><p>Module beta has no package in its root directory.</p>`,
		`<li><a href="example.com/multi/beta.html">example.com/multi/beta</a></li>`,
	}

	// Act
	err := sut.Root(&out)
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Root(...) = %v, want nil", got)
	}
	if got, want := sectionIDs(html), wantSections; !cmp.Equal(got, want) {
		t.Errorf("Renderer.Root(...) sections = %v, want %v", got, want)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Root(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderer_Package_WithSeveralModules_LinksToRoot(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRendererFor(t, loadertest.Multi(t))
	var out strings.Builder
	wantFragments := fragments{
		`<a class="sidebar-logo" href="../../../../index.html" title="Modules" aria-label="Modules">`,
		`<a class="badge badge-module" href="../../../../index.html">workspace</a>`,
		`<span class="sep">/</span><a class="badge badge-module" href="../../beta.html" title="example.com/multi/beta">module</a><span class="sep">/</span><a href="index.html">lib</a>`,
	}

	// Act
	err := sut.Package(&out, loadertest.MultiPackage(t, loadertest.MultiBetaPath, "lib"))
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Package(...) = %v, want nil", got)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Package(...) missing fragments:\n%s", strings.Join(got, "\n"))
	}
}

func TestRenderer_Type_WithTypeFromOtherModule_LinksToItsPage(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := newRendererFor(t, loadertest.Multi(t))
	var out strings.Builder
	wantFragments := fragments{
		`<a href="../../alpha/Unit.html"><span class="nx">alpha</span><span class="p">.</span><span class="nx">Unit</span></a>`,
	}

	// Act
	err := sut.Type(&out, loadertest.MultiType(t, loadertest.MultiBetaPath, "lib", "Wrapper"))
	html := out.String()

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Renderer.Type(...) = %v, want nil", got)
	}
	if got, want := missing(html, wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Renderer.Type(...) missing fragments:\n%s", strings.Join(got, "\n"))
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
			name:         "root package",
			pkg:          loadertest.Package(t, ""),
			wantSections: []string{"documentation", "packages", "examples", "constants", "errors", "variables", "interfaces", "types", "functions"},
			wantFragments: fragments{
				`<title>example.com/sample</title>`,
				`<h1><span class="kind">package</span> sample</h1>`,
				`<a class="badge badge-module" href="../sample.html" title="example.com/sample">module</a> <span class="crumb">(</span><a href="index.html">sample</a><span class="crumb">)</span>
    </nav>`,
				`<summary><a href="index.html">sample</a></summary>`,
				`<td><a href="shapes/index.html">shapes</a></td>`,
				`<td><a href="empty/index.html">empty</a> <span class="badge badge-deprecated" title="Nothing lives here.">deprecated</span></td>`,
				`<h3><a href="#errors">Sentinel Errors</a></h3>
    <ul>
      <li><a href="#var.ErrNegative">ErrNegative</a></li>
    </ul>`,
				`<h3><a href="#variables">Variables</a></h3>
    <ul>
      <li><a href="#var.DefaultColor">DefaultColor</a></li>
    </ul>`,
				`<h3><a href="#interfaces">Interfaces</a></h3>
    <ul>
      <li><a href="#type.Named">Named</a></li>
      <li><a href="#type.Namer">Namer</a></li>
      <li><a href="#type.Shape">Shape</a></li>
    </ul>`,
				`<h3><a href="#types">Types</a></h3>
    <ul>
      <li><a href="#type.Big">Big</a></li>
      <li><a href="#type.Circle">Circle</a></li>
      <li><a href="#type.Color">Color</a></li>
      <li><a href="#type.Counter">Counter</a></li>
      <li><a href="#type.Grid">Grid</a></li>
      <li><a href="#type.ID">ID</a></li>`,
				`<li><a href="#const.Legacy">Legacy <span class="badge badge-deprecated" title="Use [Version] instead. Legacy is kept only so that old callers still compile.">deprecated</span></a></li>`,
				`<span class="badge badge-deprecated" title="Use [Version] instead. Legacy is kept only so that old callers still compile.">deprecated</span> <a class="src" href="sample.go.html#L155">source</a>`,
				`<h1 id="usage">Usage</h1>`,
				`<li><a href="#usage">Usage</a></li>`,
				`<a href="Circle.html"><code>Circle</code></a>`,
				`<a href="shapes/Sizer.html"><code>shapes.Sizer</code></a>`,
				`<details class="item" id="type.Circle" open>`,
				`<a href="Circle.html"><span class="nx">Circle</span></a>`,
				`<a class="src" href="sample.go.html#L44">source</a>`,
				`<button type="button" class="read-more">Read more</button>`,
			},
		},
		{
			name:         "empty package",
			pkg:          loadertest.Package(t, "empty"),
			wantSections: []string{"documentation"},
			wantFragments: fragments{
				`<h1><span class="kind">package</span> empty <span class="badge badge-deprecated" title="Nothing lives here.">deprecated</span></h1>`,
				`<p class="message" id="empty">This package has no exported identifiers.</p>`,
				`<h3>Packages</h3>`,
				`<ul class="tree">
  <li><a href="index.html">empty <span class="badge badge-deprecated" title="Nothing lives here.">deprecated</span></a>
  </li>
</ul>`,
				`<span class="sep">/</span><a href="index.html">empty</a>`,
			},
		},
		{
			name:         "internal package",
			pkg:          loadertest.Package(t, "internal/secret"),
			wantSections: []string{"documentation", "types"},
			wantFragments: fragments{
				`<h1><span class="kind">package</span> secret <span class="badge badge-internal" title="Importable only by packages rooted at the parent of the internal directory.">internal</span></h1>`,
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
		{
			name:         "package documented by README",
			pkg:          loadertest.Package(t, "readme"),
			wantSections: []string{"documentation", "types"},
			wantFragments: fragments{
				`<h1 id="readme">Readme</h1>`,
				`<li><a href="#links">Links</a></li>`,
				`<a href="../shapes/index.html">Shapes</a>`,
				`<a href="readme.go.html#L4">The source</a>`,
				`<a href="docs/guide.md">The guide</a>`,
				`<a href="docs/missing.md">Missing</a>`,
				`<a href="../../outside.md">Outside</a>`,
				`<a href="../index.html">Root</a>`,
				`<a href="/absolute">Absolute</a>`,
				`<a href="https://example.com/x">external</a>`,
				`<a href="#links">Anchor</a>`,
				`<li>[Sizer] is not a doc link.</li>`,
				`<p align="center">Raw HTML stays.</p>`,
				`<img src="docs/diagram.svg" alt="Diagram">`,
			},
		},
		{
			name:         "package documented by index.md",
			pkg:          loadertest.Package(t, "indexed"),
			wantSections: []string{"documentation", "types"},
			wantFragments: fragments{
				`<h1 id="indexed">Indexed</h1>`,
				`<p>Package indexed is documented by index.md.</p>`,
			},
		},
		{
			name:         "doc comment wins over README",
			pkg:          loadertest.Package(t, "shapes"),
			wantSections: []string{"documentation", "interfaces", "functions"},
			wantFragments: fragments{
				`<p>Package shapes declares interfaces satisfied by shapes.</p>`,
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
			wantSections: []string{"documentation", "examples", "fields", "constructors", "methods", "utilities", "implements"},
			wantFragments: fragments{
				`<title>Circle - example.com/sample</title>`,
				`<h1><span class="kind">struct</span> Circle</h1>`,
				`<a class="src" href="sample.go.html#L44">source</a>`,
				`// contains unexported fields`,
				`<details class="item" id="field.Radius" open>`,
				`<span class="nx">Radius</span> <a href="https://pkg.go.dev/builtin#float64"><span class="kt">float64</span></a>`,
				`<a class="src" href="sample.go.html#L46">source</a>`,
				`<div class="docblock summary"><p>Radius is the circle radius.</p>`,
				`<details class="item" id="field.Color" open>`,
				`<li><a href="#field.Color">Color</a></li>`,
				`<li class="badge badge-prop" title="Size in bytes on a 64-bit system."><span class="badge-label">size</span> <span class="badge-value">24 bytes</span></li>`,
				`<li class="badge badge-prop" title="Values can be compared with == and !=, and can be map keys."><span class="badge-label">comparable</span></li>`,
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
				`<details class="item" id="example-Circle" open>
        <summary><h3 class="example-title">Example</h3></summary>`,
				`<p class="example-output-label">Output:</p>
          <pre class="example-output">circle`,
			},
		},
		{
			name:         "struct with embedded and shared fields",
			typ:          loadertest.Type(t, "", "Label"),
			wantSections: []string{"documentation", "fields", "implements"},
			wantFragments: fragments{
				`<details class="item" id="field.Reader" open>`,
				`<span class="badge badge-embedded" title="Embedded in the struct. Its fields and methods are promoted to the struct.">embedded</span>`,
				`<details class="item" id="field.Text" open>`,
				`<div class="docblock summary"><p>Text is the label text.</p>`,
				`<button type="button" class="read-more">Read more</button>`,
				`<details class="item" id="field.Width" open>`,
				`<details class="item" id="field.Height" open>`,
				`<div class="docblock summary"><p>Width and Height are the size of the label box.</p>`,
				`<details class="item" id="field.Font" open>`,
				`<span class="badge badge-deprecated" title="Use Text instead.">deprecated</span>`,
				`<li><a href="#field.Legacy">Legacy <span class="badge badge-deprecated" title="Use Text instead.">deprecated</span></a></li>`,
			},
		},
		{
			name:         "interface",
			typ:          loadertest.Type(t, "", "Named"),
			wantSections: []string{"documentation", "constructors", "methods", "implementations"},
			wantFragments: fragments{
				`<h1><span class="kind">interface</span> Named</h1>`,
				`<details class="item" id="method.Name" open>`,
				`<a href="Named.Name.html"><span class="nf">Name</span></a>`,
				`<details class="item" id="ctor.NewCircle" open>`,
				`<details class="item" id="ctor.NewUnit" open>`,
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
			wantSections: []string{"documentation", "constructors", "methods", "implementations"},
			wantFragments: fragments{
				`<span class="badge-label">sealed</span>`,
				`// contains unexported methods`,
			},
		},
		{
			name:         "enum with instances",
			typ:          loadertest.Type(t, "", "Color"),
			wantSections: []string{"documentation", "instances", "utilities"},
			wantFragments: fragments{
				`<li class="badge badge-prop" title="Values can be compared with &lt;, &lt;=, &gt;, and &gt;=."><span class="badge-label">ordered</span></li>`,
				`<details class="item" id="instance.Red" open>`,
				`<details class="item" id="instance.DefaultColor" open>`,
			},
		},
		{
			name:         "internal type",
			typ:          loadertest.Type(t, "internal/secret", "Token"),
			wantSections: []string{"documentation", "fields"},
			wantFragments: fragments{
				`<h1><span class="kind">struct</span> Token <span class="badge badge-internal" title="Importable only by packages rooted at the parent of the internal directory.">internal</span></h1>`,
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

func TestRenderer_WithUnexported_MarksUnexportedSymbols(t *testing.T) {
	t.Parallel()

	const badge = `<span class="badge badge-unexported" title="Not exported. It cannot be named outside its package.">unexported</span>`

	testCases := []struct {
		name          string
		render        func(*render.Renderer, *strings.Builder) error
		wantFragments fragments
	}{
		{
			name: "package page lists unexported symbols last",
			render: func(r *render.Renderer, out *strings.Builder) error {
				return r.Package(out, loadertest.UnexportedPackage(t, ""))
			},
			wantFragments: fragments{
				`<li><a href="#type.Stack">Stack</a></li>
      <li><a href="#type.point">point ` + badge + `</a></li>`,
				`<li><a href="#func.ZeroCircle">ZeroCircle</a></li>
      <li><a href="#func.newPoint">newPoint ` + badge + `</a></li>
      <li><a href="#func.unexported">unexported ` + badge + `</a></li>`,
				`<li><a href="#const.Version">Version</a></li>
      <li><a href="#const.maxPoints">maxPoints ` + badge + `</a></li>`,
				`<details class="item" id="type.point" open>`,
				`<a href="~point.html"><span class="nx">point</span></a>`,
				`<div class="item-links">` + badge + ` <a class="src" href="sample.go.html#L184">source</a></div>`,
			},
		},
		{
			name: "unexported type page",
			render: func(r *render.Renderer, out *strings.Builder) error {
				return r.Type(out, loadertest.UnexportedType(t, "", "point"))
			},
			wantFragments: fragments{
				`<h1><span class="kind">struct</span> point ` + badge + `</h1>`,
				`<details class="item" id="field.y" open>`,
				`<div class="item-links">` + badge + ` <a class="src" href="sample.go.html#L189">source</a></div>`,
				`<details class="item" id="ctor.newPoint" open>`,
				`<a href="~newPoint.html"><span class="nf">newPoint</span></a>`,
				`<details class="item" id="method.Name" open>`,
				`<a href="~point.Name.html"><span class="nf">Name</span></a>`,
				`<details class="item" id="method.shift" open>`,
				`<a href="~point.~shift.html"><span class="nf">shift</span></a>`,
				`<details class="item" id="instance.origin" open>`,
			},
		},
		{
			name: "interface page lists unexported implementation last",
			render: func(r *render.Renderer, out *strings.Builder) error {
				return r.Type(out, loadertest.UnexportedType(t, "", "Named"))
			},
			wantFragments: fragments{
				`<li><a href="#implementations.example.com/sample.Square">Square</a></li>
      <li><a href="#implementations.example.com/sample.point">point ` + badge + `</a></li>`,
				`<span class="badge badge-receiver" title="implemented by *point"><code>*point</code></span> ` + badge,
				`<details class="item" id="ctor.newPoint" open>`,
			},
		},
		{
			name: "exported type page shows unexported members",
			render: func(r *render.Renderer, out *strings.Builder) error {
				return r.Type(out, loadertest.UnexportedType(t, "", "Circle"))
			},
			wantFragments: fragments{
				`<span class="nx">hidden</span> <a href="https://pkg.go.dev/builtin#int"><span class="kt">int</span></a>`,
				`<details class="item" id="field.hidden" open>`,
				`<details class="item" id="method.sealed" open>`,
				`<a href="Circle.~sealed.html"><span class="nf">sealed</span></a>`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := newRendererFor(t, loadertest.Unexported(t))
			var out strings.Builder

			// Act
			err := tc.render(sut, &out)
			html := out.String()

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Renderer render = %v, want nil", got)
			}
			if got, want := missing(html, tc.wantFragments), []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Renderer missing fragments:\n%s", strings.Join(got, "\n"))
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
				`<details class="item" id="example-NewCircle" open>`,
				`<pre class="example-output">circle
</pre>`,
			},
		},
		{
			name:         "method",
			fn:           loadertest.Method(t, "", "Circle", "Area"),
			wantSections: []string{"documentation", "examples"},
			wantFragments: fragments{
				`<title>Circle.Area - example.com/sample</title>`,
				`<h1><span class="kind">func</span> Area</h1>`,
				`<span class="sep">.</span><a href="Circle.html">Circle</a><span class="sep">.</span><span class="crumb">Area</span>`,
			},
		},
		{
			name:         "interface method",
			fn:           loadertest.Method(t, "", "Shape", "Area"),
			wantSections: []string{"documentation"},
			wantFragments: fragments{
				`<h1><span class="kind">func</span> Area</h1>`,
				`<span class="sep">.</span><a href="Shape.html">Shape</a><span class="sep">.</span><span class="crumb">Area</span>`,
				`<a class="src" href="sample.go.html#L12">source</a>`,
				`<p>Area returns the surface area.</p>`,
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
