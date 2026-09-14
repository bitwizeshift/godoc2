package generate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/emit"
	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/progress"
	"github.com/bitwizeshift/godoc2/internal/relate"
	"github.com/bitwizeshift/godoc2/internal/render"
	"github.com/bitwizeshift/godoc2/internal/search"
)

// ErrGenerate wraps every error reported by [Generator.Generate].
var ErrGenerate = errors.New("generate")

// Loader loads the module to document.
type Loader interface {
	Load(ctx context.Context) (*model.Module, error)
}

// Generator builds the documentation site.
type Generator struct {
	// Loader loads the module.
	Loader Loader

	// Sink receives the generated files.
	Sink emit.Sink

	// Reporter receives progress. A nil Reporter discards progress.
	Reporter progress.Reporter
}

// Generate loads the module and writes every page. It returns an error
// wrapping [ErrGenerate] together with the failing stage and the cause.
func (g *Generator) Generate(ctx context.Context) error {
	reporter := g.Reporter
	if reporter == nil {
		reporter = progress.Discard()
	}
	run := &run{gen: g, reporter: reporter, search: &search.Index{}}
	return run.execute(ctx)
}

// run holds the state of one generation.
type run struct {
	gen      *Generator
	reporter progress.Reporter
	module   *model.Module
	renderer *render.Renderer
	docfiles *docfile.Resolver
	search   *search.Index
}

func (r *run) execute(ctx context.Context) error {
	r.reporter.Stage("Loading packages")
	mod, err := r.gen.Loader.Load(ctx)
	if err != nil {
		return r.fail("Loading packages", err)
	}
	r.module = mod

	r.reporter.Stage("Indexing")
	r.docfiles = docfile.NewResolver(mod)
	renderer, err := render.New(mod, link.New(mod), relate.New(mod), r.docfiles)
	if err != nil {
		return r.fail("Indexing", err)
	}
	r.renderer = renderer

	r.reporter.Stage("Writing static files")
	if err := r.staticFiles(); err != nil {
		return r.fail("Writing static files", err)
	}

	r.reporter.Stage("Writing module page")
	if err := r.modulePage(); err != nil {
		return r.fail("Writing module page", err)
	}

	for _, p := range mod.Packages {
		stage := "Writing package " + p.ImportPath
		r.reporter.Stage(stage)
		if err := r.packagePages(p); err != nil {
			return r.fail(stage, err)
		}
	}

	r.reporter.Stage("Writing search index")
	if err := r.write(pathmap.Static(search.FileName), r.search.Write); err != nil {
		return r.fail("Writing search index", err)
	}
	return nil
}

func (r *run) fail(stage string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrGenerate, stage, err)
}

// write creates the file at path, fills it with fn, and closes it.
func (r *run) write(path string, fn func(io.Writer) error) error {
	f, err := r.gen.Sink.Create(path)
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	r.reporter.File(path)
	return nil
}

func (r *run) staticFiles() error {
	if err := r.write(pathmap.Static(render.CSSFile), r.renderer.CSS); err != nil {
		return err
	}
	return r.write(pathmap.Static(render.JSFile), r.renderer.JS)
}

// modulePage writes the module page, which is also the page of the root
// package when the module has one.
func (r *run) modulePage() error {
	root := r.rootPackage()
	if root != nil {
		r.search.Add(search.Entry{
			Name:    root.ImportPath,
			Kind:    "package",
			Package: root.ImportPath,
			Path:    pathmap.Module(r.module.Path),
		})
	}
	return r.write(pathmap.Module(r.module.Path), func(w io.Writer) error {
		return r.renderer.Module(w, root)
	})
}

func (r *run) rootPackage() *model.Package {
	for _, p := range r.module.Packages {
		if p.RelPath == "" {
			return p
		}
	}
	return nil
}

// packagePages writes the package page, every symbol page, and every source
// page of p.
func (r *run) packagePages(p *model.Package) error {
	if p.RelPath != "" {
		r.search.Add(search.Entry{
			Name:    p.ImportPath,
			Kind:    "package",
			Package: p.ImportPath,
			Path:    pathmap.Package(r.module.Path, p.RelPath),
		})
		if err := r.write(pathmap.Package(r.module.Path, p.RelPath), func(w io.Writer) error {
			return r.renderer.Package(w, p)
		}); err != nil {
			return err
		}
	}
	for _, t := range p.Types {
		if err := r.typePages(t); err != nil {
			return err
		}
	}
	for _, f := range p.Funcs {
		if err := r.funcPage(f); err != nil {
			return err
		}
	}
	for _, v := range append(append([]*model.Value{}, p.Consts...), p.Vars...) {
		if err := r.valuePage(v); err != nil {
			return err
		}
	}
	for _, f := range p.Files {
		if err := r.sourcePage(p, f); err != nil {
			return err
		}
	}
	return nil
}

func (r *run) typePages(t *model.Type) error {
	path := pathmap.Symbol(r.module.Path, t.Pkg.RelPath, t.Name)
	r.search.Add(search.Entry{Name: t.Name, Kind: "type", Package: t.Pkg.ImportPath, Path: path})
	if err := r.write(path, func(w io.Writer) error {
		return r.renderer.Type(w, t)
	}); err != nil {
		return err
	}
	for _, m := range t.Methods {
		if err := r.funcPage(m); err != nil {
			return err
		}
	}
	return nil
}

func (r *run) funcPage(f *model.Func) error {
	entry := search.Entry{Name: f.Name, Kind: "func", Package: f.Pkg.ImportPath}
	if f.Recv != nil {
		entry.Name = f.Recv.Name + "." + f.Name
		entry.Kind = "method"
		entry.Path = pathmap.Method(r.module.Path, f.Pkg.RelPath, f.Recv.Name, f.Name)
	} else {
		entry.Path = pathmap.Symbol(r.module.Path, f.Pkg.RelPath, f.Name)
	}
	r.search.Add(entry)
	return r.write(entry.Path, func(w io.Writer) error {
		return r.renderer.Func(w, f)
	})
}

func (r *run) valuePage(v *model.Value) error {
	path := pathmap.Symbol(r.module.Path, v.Pkg.RelPath, v.Name)
	r.search.Add(search.Entry{Name: v.Name, Kind: v.Kind.String(), Package: v.Pkg.ImportPath, Path: path})
	return r.write(path, func(w io.Writer) error {
		return r.renderer.Value(w, v)
	})
}

func (r *run) sourcePage(p *model.Package, f *model.File) error {
	src, err := os.ReadFile(f.Path)
	if err != nil {
		return err
	}
	return r.write(pathmap.Source(r.module.Path, p.RelPath, f.Name), func(w io.Writer) error {
		return r.renderer.Source(w, p, f, src)
	})
}
