package render

import (
	"go/ast"
	"html/template"
	"path"
	"path/filepath"
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

// newPage returns a page with the chrome that every page shares. mod is the
// module of the page, or nil for the root page.
func (b *builder) newPage(title string, mod *model.Module) *page {
	pg := &page{
		Title:     title,
		HomeHref:  b.rel(pathmap.Root()),
		HomeTitle: rootTitle,
		Favicon:   b.rel(pathmap.Static(FaviconFile)),
		CSS:       b.rel(pathmap.Static(CSSFile)),
		JS:        b.rel(pathmap.Static(JSFile)),
		IndexJS:   b.rel(pathmap.Static("search-index.js")),
		Root:      b.rel(""),
	}
	if mod != nil && !b.multiModule() {
		pg.HomeHref = b.rel(pathmap.Module(mod.Path))
		pg.HomeTitle = mod.Path
	}
	return pg
}

// multiModule reports whether the site has a root page that lists modules.
func (b *builder) multiModule() bool {
	return len(b.r.site.Modules) > 1
}

// printer returns a signature printer for the page. self is the URL the
// declared name links to.
func (b *builder) printer(self string) *sig.Printer {
	return &sig.Printer{Resolver: b.r.resolver, Pkg: b.pkg, From: b.from, Self: self, Unexported: b.r.site.Unexported}
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
	pos := b.r.site.Fset.Position(n.Pos())
	if !pos.IsValid() {
		return ""
	}
	file := filepath.Base(pos.Filename)
	return b.rel(pathmap.Source(b.pkg.Module.Path, b.pkg.RelPath, file)) + "#L" + strconv.Itoa(pos.Line)
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

// breadcrumb returns the path from the module badge to the package p of mod,
// followed by the given symbol crumbs. p is nil for the module page. In a
// site with several modules, the workspace badge comes first. On the pages
// below a module with a root package, the crumb of the root package follows
// the module badge.
func (b *builder) breadcrumb(mod *model.Module, p *model.Package, symbols ...crumb) []crumb {
	var crumbs []crumb
	module := crumb{Text: "module", Href: b.rel(pathmap.Module(mod.Path)), Badge: true, Title: mod.Path}
	if b.multiModule() {
		module.Separator = "/"
		crumbs = append(crumbs, b.rootCrumb())
	}
	crumbs = append(crumbs, module)
	if root := mod.Root(); p != nil && root != nil {
		crumbs = append(crumbs, crumb{Text: root.DisplayName(), Href: b.rel(pathmap.Package(mod.Path, "")), Root: true})
	}
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

// rootCrumb returns the badge that links to the root page of the site.
func (b *builder) rootCrumb() crumb {
	return crumb{Text: "workspace", Href: b.rel(pathmap.Root()), Badge: true}
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
	entry := &sidebarSection{Title: s.Title, Href: "#" + s.ID}
	for _, it := range s.Items {
		entry.Items = append(entry.Items, sidebarItem{Text: it.Name, Href: "#" + it.ID, Internal: it.Internal, Unexported: it.Unexported, Deprecated: it.Deprecated})
	}
	for _, row := range s.Rows {
		entry.Items = append(entry.Items, sidebarItem{Text: row.Name, Href: row.Href, Internal: row.Internal, Deprecated: row.Deprecated})
	}
	for _, ex := range s.Examples {
		entry.Items = append(entry.Items, sidebarItem{Text: ex.Title, Href: "#" + ex.ID})
	}
	for _, g := range s.Groups {
		for _, it := range g.Items {
			entry.Items = append(entry.Items, sidebarItem{Text: it.Name, Href: "#" + it.ID, Unexported: it.Unexported})
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
