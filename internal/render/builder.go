package render

import (
	"go/ast"
	"html/template"
	"path"
	"strconv"
	"strings"

	"github.com/bitwizeshift/godoc2/internal/markdown"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/sig"
)

// builder assembles the view model of one page.
type builder struct {
	r    *Renderer
	from string
	pkg  *model.Package
}

func (r *Renderer) builder(from string, pkg *model.Package) *builder {
	return &builder{r: r, from: from, pkg: pkg}
}

// rel returns the href from the page to an output path.
func (b *builder) rel(to string) string {
	return pathmap.Rel(b.from, to)
}

// newPage returns a page with the chrome that every page shares.
func (b *builder) newPage(title string) *page {
	mod := b.r.module
	return &page{
		Title:      title,
		ModulePath: mod.Path,
		ModuleHref: b.rel(pathmap.Module(mod.Path)),
		CSS:        b.rel(pathmap.Static(CSSFile)),
		JS:         b.rel(pathmap.Static(JSFile)),
		IndexJS:    b.rel(pathmap.Static("search-index.js")),
		Root:       b.rel(""),
	}
}

// printer returns a signature printer for the page. self is the URL the
// declared name links to.
func (b *builder) printer(self string) *sig.Printer {
	return &sig.Printer{Resolver: b.r.resolver, Pkg: b.pkg, From: b.from, Self: self}
}

// doc parses a doc comment in the scope of the page package. It returns nil
// for a blank comment.
func (b *builder) doc(text string) *markdown.Document {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	scope := packageScope{resolver: b.r.resolver, pkg: b.pkg, from: b.from}
	return b.r.markdown.Parse(text, scope)
}

// docFile parses a Markdown documentation file. Its relative links are
// rewritten to hrefs from the page.
func (b *builder) docFile(f *model.DocFile) *markdown.Document {
	doc := b.r.markdown.ParseMarkdown(f.Text)
	doc.RewriteLinks(func(dest string) string {
		return b.r.docfiles.Resolve(f, b.from, dest)
	})
	return doc
}

// packageDoc parses the documentation of p: its doc comment, or else its
// Markdown file. It returns nil when p has neither.
func (b *builder) packageDoc(p *model.Package) *markdown.Document {
	if doc := b.doc(p.Doc); doc != nil {
		return doc
	}
	if p.DocFile != nil {
		return b.docFile(p.DocFile)
	}
	return nil
}

// docSection returns the Documentation section, or nil for a nil document.
func (b *builder) docSection(doc *markdown.Document) *section {
	if doc == nil {
		return nil
	}
	html, err := doc.HTML()
	if err != nil {
		return nil
	}
	return &section{ID: "documentation", Title: "Documentation", Kind: kindDoc, HTML: html}
}

// docSidebar returns the Sections sidebar entry for a documented page.
func (b *builder) docSidebar(doc *markdown.Document) sidebarSection {
	items := []sidebarItem{{Text: "Documentation", Href: "#documentation"}}
	if doc != nil {
		for _, h := range doc.Headings() {
			items = append(items, sidebarItem{Text: h.Text, Href: "#" + h.ID})
		}
	}
	return sidebarSection{Title: "Sections", Items: items}
}

// summaryAndFull renders the first paragraph of doc and, when the document
// has more than that paragraph, the full document.
func (b *builder) summaryAndFull(doc *markdown.Document) (template.HTML, template.HTML) {
	if doc == nil {
		return "", ""
	}
	summary, err := doc.Summary()
	if err != nil {
		return "", ""
	}
	full, err := doc.HTML()
	if err != nil || full == summary {
		return summary, ""
	}
	return summary, full
}

// code highlights a rendered declaration.
func (b *builder) code(r sig.Rendered) template.HTML {
	html, err := b.r.highlight.Code(r)
	if err != nil {
		return template.HTML(template.HTMLEscapeString(r.Text))
	}
	return html
}

// sourceHref returns the href to the source line of a node.
func (b *builder) sourceHref(n ast.Node) string {
	if n == nil {
		return ""
	}
	pos := b.r.module.Fset.Position(n.Pos())
	if !pos.IsValid() {
		return ""
	}
	file := path.Base(pos.Filename)
	return b.rel(pathmap.Source(b.r.module.Path, b.pkg.RelPath, file)) + "#L" + strconv.Itoa(pos.Line)
}

// examplesSection returns the Examples section, or nil when there are none.
func (b *builder) examplesSection(exs []*model.Example) *section {
	if len(exs) == 0 {
		return nil
	}
	s := &section{ID: "examples", Title: "Examples", Kind: kindExamples}
	for _, ex := range exs {
		code, err := b.r.highlight.Snippet(ex.Code)
		if err != nil {
			code = template.HTML(template.HTMLEscapeString(ex.Code))
		}
		var doc template.HTML
		if parsed := b.doc(ex.Doc); parsed != nil {
			doc, _ = parsed.HTML()
		}
		s.Examples = append(s.Examples, example{
			ID:     "example-" + ex.Name,
			Title:  exampleTitle(ex),
			Doc:    doc,
			Code:   code,
			Output: ex.Output,
		})
	}
	return s
}

func exampleTitle(ex *model.Example) string {
	if ex.Suffix != "" {
		return "Example (" + ex.Suffix + ")"
	}
	return "Example"
}

// breadcrumb returns the path from the module badge to the package, followed
// by the given symbol crumbs.
func (b *builder) breadcrumb(p *model.Package, symbols ...crumb) []crumb {
	mod := b.r.module
	crumbs := []crumb{{Text: "module", Href: b.rel(pathmap.Module(mod.Path)), Module: true}}
	if p != nil && p.RelPath != "" {
		elems := strings.Split(p.RelPath, "/")
		for i, elem := range elems {
			rel := strings.Join(elems[:i+1], "/")
			c := crumb{Text: elem, Separator: "/"}
			if b.r.resolver.Local(path.Join(mod.Path, rel)) {
				c.Href = b.rel(pathmap.Package(mod.Path, rel))
			}
			crumbs = append(crumbs, c)
		}
	}
	for _, c := range symbols {
		if c.Separator == "" {
			c.Separator = "."
		}
		crumbs = append(crumbs, c)
	}
	return crumbs
}

// appendSection adds s to the page when it is not nil.
func appendSection(sections []section, s *section) []section {
	if s == nil {
		return sections
	}
	return append(sections, *s)
}

// itemsSection returns a section of items, or nil when items is empty.
func itemsSection(id, title string, items []item) *section {
	if len(items) == 0 {
		return nil
	}
	return &section{ID: id, Title: title, Kind: kindItems, Items: items}
}

// sidebarOf returns a sidebar entry that links to every item of a section.
func sidebarOf(s *section) *sidebarSection {
	if s == nil {
		return nil
	}
	entry := &sidebarSection{Title: s.Title}
	for _, it := range s.Items {
		entry.Items = append(entry.Items, sidebarItem{Text: it.Name, Href: "#" + it.ID, Internal: it.Internal})
	}
	for _, row := range s.Rows {
		entry.Items = append(entry.Items, sidebarItem{Text: row.Name, Href: row.Href, Internal: row.Internal})
	}
	for _, ex := range s.Examples {
		entry.Items = append(entry.Items, sidebarItem{Text: ex.Title, Href: "#" + ex.ID})
	}
	for _, g := range s.Groups {
		for _, it := range g.Items {
			entry.Items = append(entry.Items, sidebarItem{Text: it.Name, Href: "#" + it.ID})
		}
	}
	if len(entry.Items) == 0 {
		entry.Items = []sidebarItem{{Text: s.Title, Href: "#" + s.ID}}
	}
	return entry
}

func appendSidebar(sidebar []sidebarSection, s *sidebarSection) []sidebarSection {
	if s == nil {
		return sidebar
	}
	return append(sidebar, *s)
}
