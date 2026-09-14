package render

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/bitwizeshift/godoc2/internal/docfile"
	"github.com/bitwizeshift/godoc2/internal/highlight"
	"github.com/bitwizeshift/godoc2/internal/link"
	"github.com/bitwizeshift/godoc2/internal/markdown"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/relate"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/godoc2.css
var siteCSS string

//go:embed static/godoc2.js
var siteJS string

// Asset names under the static directory.
const (
	CSSFile = "godoc2.css"
	JSFile  = "godoc2.js"
)

// Renderer writes the pages of one site.
type Renderer struct {
	site      *model.Site
	resolver  *link.Resolver
	index     *relate.Index
	docfiles  *docfile.Resolver
	markdown  *markdown.Renderer
	highlight *highlight.Highlighter
	tmpl      *template.Template
}

// New returns a [Renderer] for site. docfiles resolves the links of Markdown
// documentation files. It returns an error if the embedded templates do not
// parse.
func New(site *model.Site, resolver *link.Resolver, index *relate.Index, docfiles *docfile.Resolver) (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return &Renderer{
		site:      site,
		resolver:  resolver,
		index:     index,
		docfiles:  docfiles,
		markdown:  markdown.New(),
		highlight: highlight.New(),
		tmpl:      tmpl,
	}, nil
}

// CSS writes the site stylesheet, including the highlighting rules.
func (r *Renderer) CSS(w io.Writer) error {
	if _, err := io.WriteString(w, siteCSS); err != nil {
		return err
	}
	return r.highlight.CSS(w)
}

// JS writes the site script.
func (r *Renderer) JS(w io.Writer) error {
	_, err := io.WriteString(w, siteJS)
	return err
}

// Root writes the root page of the site. For a site with one module it is a
// redirect to the module page. For several modules it lists them.
func (r *Renderer) Root(w io.Writer) error {
	if len(r.site.Modules) == 1 {
		return r.redirect(w, r.site.Modules[0])
	}
	b := r.builder(pathmap.Root(), nil)
	return r.execute(w, b.rootPage())
}

// Module writes the page of mod, which is also the page of its root package
// when the module has one.
func (r *Renderer) Module(w io.Writer, mod *model.Module) error {
	root := mod.Root()
	b := r.builder(pathmap.Module(mod.Path), root)
	return r.execute(w, b.modulePage(mod, root))
}

// Package writes the page of p.
func (r *Renderer) Package(w io.Writer, p *model.Package) error {
	b := r.builder(pathmap.Package(p.Module.Path, p.RelPath), p)
	return r.execute(w, b.packagePage(p))
}

// Type writes the page of t.
func (r *Renderer) Type(w io.Writer, t *model.Type) error {
	b := r.builder(pathmap.Symbol(t.Pkg.Module.Path, t.Pkg.RelPath, t.Name), t.Pkg)
	return r.execute(w, b.typePage(t))
}

// Func writes the page of a function or method.
func (r *Renderer) Func(w io.Writer, f *model.Func) error {
	b := r.builder(funcPath(f), f.Pkg)
	return r.execute(w, b.funcPage(f))
}

// Value writes the page of a constant or variable.
func (r *Renderer) Value(w io.Writer, v *model.Value) error {
	b := r.builder(pathmap.Symbol(v.Pkg.Module.Path, v.Pkg.RelPath, v.Name), v.Pkg)
	return r.execute(w, b.valuePage(v))
}

// Source writes the highlighted page of a source file of p.
func (r *Renderer) Source(w io.Writer, p *model.Package, file *model.File, src []byte) error {
	b := r.builder(pathmap.Source(p.Module.Path, p.RelPath, file.Name), p)
	pg, err := b.sourcePage(p, file, src)
	if err != nil {
		return err
	}
	return r.execute(w, pg)
}

func (r *Renderer) execute(w io.Writer, pg *page) error {
	if err := r.tmpl.ExecuteTemplate(w, "layout.html", pg); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}

// redirect writes a page that sends the browser to the page of mod.
func (r *Renderer) redirect(w io.Writer, mod *model.Module) error {
	pg := redirect{
		Title: mod.Path,
		Href:  pathmap.Rel(pathmap.Root(), pathmap.Module(mod.Path)),
	}
	if err := r.tmpl.ExecuteTemplate(w, "redirect.html", pg); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}

// funcPath returns the output path of a function or method page.
func funcPath(f *model.Func) string {
	if f.Recv != nil {
		return pathmap.Method(f.Pkg.Module.Path, f.Pkg.RelPath, f.Recv.Name, f.Name)
	}
	return pathmap.Symbol(f.Pkg.Module.Path, f.Pkg.RelPath, f.Name)
}
