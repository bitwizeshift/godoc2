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

// Renderer writes the pages of one module.
type Renderer struct {
	module    *model.Module
	resolver  *link.Resolver
	index     *relate.Index
	docfiles  *docfile.Resolver
	markdown  *markdown.Renderer
	highlight *highlight.Highlighter
	tmpl      *template.Template
}

// New returns a [Renderer] for mod. docfiles resolves the links of Markdown
// documentation files. It returns an error if the embedded templates do not
// parse.
func New(mod *model.Module, resolver *link.Resolver, index *relate.Index, docfiles *docfile.Resolver) (*Renderer, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return &Renderer{
		module:    mod,
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

// Module writes the module page. root is the root package of the module, or
// nil when the module root holds no package.
func (r *Renderer) Module(w io.Writer, root *model.Package) error {
	b := r.builder(pathmap.Module(r.module.Path), root)
	return r.execute(w, b.modulePage(root))
}

// Package writes the page of p.
func (r *Renderer) Package(w io.Writer, p *model.Package) error {
	b := r.builder(pathmap.Package(r.module.Path, p.RelPath), p)
	return r.execute(w, b.packagePage(p))
}

// Type writes the page of t.
func (r *Renderer) Type(w io.Writer, t *model.Type) error {
	b := r.builder(pathmap.Symbol(r.module.Path, t.Pkg.RelPath, t.Name), t.Pkg)
	return r.execute(w, b.typePage(t))
}

// Func writes the page of a function or method.
func (r *Renderer) Func(w io.Writer, f *model.Func) error {
	b := r.builder(funcPath(r.module.Path, f), f.Pkg)
	return r.execute(w, b.funcPage(f))
}

// Value writes the page of a constant or variable.
func (r *Renderer) Value(w io.Writer, v *model.Value) error {
	b := r.builder(pathmap.Symbol(r.module.Path, v.Pkg.RelPath, v.Name), v.Pkg)
	return r.execute(w, b.valuePage(v))
}

// Source writes the highlighted page of a source file of p.
func (r *Renderer) Source(w io.Writer, p *model.Package, file *model.File, src []byte) error {
	b := r.builder(pathmap.Source(r.module.Path, p.RelPath, file.Name), p)
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

// funcPath returns the output path of a function or method page.
func funcPath(modulePath string, f *model.Func) string {
	if f.Recv != nil {
		return pathmap.Method(modulePath, f.Pkg.RelPath, f.Recv.Name, f.Name)
	}
	return pathmap.Symbol(modulePath, f.Pkg.RelPath, f.Name)
}
