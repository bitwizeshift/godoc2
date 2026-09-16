package render

import (
	"html/template"
	"path"
	"slices"
	"strings"

	"github.com/bitwizeshift/godoc2/internal/markdown"
	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/props"
	"github.com/bitwizeshift/godoc2/internal/relate"
)

const noExportedMessage = "This package has no exported identifiers."

// rootTitle is the heading of the root page of a site with several modules.
const rootTitle = "Modules"

// rootPage builds the root page of a site with several modules: the
// Markdown file of the site directory, then the Modules table.
func (b *builder) rootPage() *page {
	pg := b.newPage(rootTitle, nil)
	pg.Breadcrumb = []crumb{b.rootCrumb()}
	pg.Heading = heading{Name: rootTitle}
	var doc *markdown.Document
	if b.r.site.DocFile != nil {
		doc = b.docFile(b.r.site.DocFile)
	}
	modules := b.modulesSection()
	pg.Sections = appendSection(pg.Sections, b.docSection(doc))
	pg.Sections = appendSection(pg.Sections, modules)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(modules))
	return pg
}

// modulesSection lists every module of the site with the summary of its
// documentation.
func (b *builder) modulesSection() *section {
	var rows []tableRow
	for _, mod := range b.r.site.Modules {
		summary, _ := b.summaryAndFull(b.moduleDoc(mod, mod.Root()))
		rows = append(rows, tableRow{
			Name:       mod.Path,
			Href:       b.rel(pathmap.Module(mod.Path)),
			Deprecated: mod.Deprecated(),
			Summary:    summary,
		})
	}
	return &section{ID: "modules", Title: rootTitle, Kind: kindTable, Rows: rows}
}

// modulePage builds the page of mod, which is the page of its root package
// with the Tools and Packages tables. root is nil for a module without a
// root package.
func (b *builder) modulePage(mod *model.Module, root *model.Package) *page {
	pg := b.newPage(mod.Path, mod)
	pg.Breadcrumb = b.breadcrumb(mod, root)
	pg.Heading = heading{Kind: "module", Name: mod.Path}
	if root != nil {
		pg.Heading = heading{Kind: "package", Name: root.DisplayName()}
		pg.Title = root.ImportPath
	}
	doc := b.moduleDoc(mod, root)
	tools := b.toolsSection(mod)
	packages := b.packagesSection(mod, "")
	pg.Sections = appendSection(pg.Sections, b.docSection(doc))
	pg.Sections = appendSection(pg.Sections, tools)
	pg.Sections = appendSection(pg.Sections, packages)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(tools))
	pg.Sidebar = appendSidebar(pg.Sidebar, b.treeSidebar(packages, mod, root))
	if root != nil {
		b.addPackageMembers(pg, root, tools != nil || packages != nil)
	}
	return pg
}

// moduleDoc parses the documentation of the module page: that of the root
// package, or else the Markdown file of the module root.
func (b *builder) moduleDoc(mod *model.Module, root *model.Package) *markdown.Document {
	if root != nil {
		return b.withPackage(root).packageDoc(root)
	}
	if mod.DocFile != nil {
		return b.docFile(mod.DocFile)
	}
	return nil
}

// packagePage builds the page of a package below the module root.
func (b *builder) packagePage(p *model.Package) *page {
	pg := b.newPage(p.ImportPath, p.Module)
	pg.Breadcrumb = b.breadcrumb(p.Module, p)
	pg.Heading = heading{Kind: "package", Name: p.DisplayName(), Internal: p.Internal(), Deprecated: p.Deprecated()}
	if p.Tool() {
		pg.Heading.Kind = "binary"
	}
	doc := b.packageDoc(p)
	packages := b.packagesSection(p.Module, p.RelPath)
	pg.Sections = appendSection(pg.Sections, b.docSection(doc))
	pg.Sections = appendSection(pg.Sections, packages)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, b.treeSidebar(packages, p.Module, p))
	b.addPackageMembers(pg, p, packages != nil)
	return pg
}

// addPackageMembers appends the Examples, Constants, Sentinel Errors,
// Variables, Interfaces, Types, and Functions sections of p, or the
// no-exported-identifiers message when the package declares nothing and has
// no tables. The Sentinel Errors section holds the variables of type error,
// and the Variables section holds the other variables. The Interfaces
// section holds the types whose underlying type is an interface, aliases
// included, and the Types section holds the other types.
func (b *builder) addPackageMembers(pg *page, p *model.Package, hasTables bool) {
	errs, otherVars := partition(p.Vars, (*model.Value).SentinelError)
	ifaces, otherTypes := partition(p.Types, (*model.Type).Interface)
	examples := b.examplesSection(p.Examples)
	consts := itemsSection("constants", "Constants", b.valueItems("const", p.Consts))
	sentinels := itemsSection("errors", "Sentinel Errors", b.valueItems("var", errs))
	vars := itemsSection("variables", "Variables", b.valueItems("var", otherVars))
	interfaces := itemsSection("interfaces", "Interfaces", b.typeItems(ifaces))
	types := itemsSection("types", "Types", b.typeItems(otherTypes))
	funcs := itemsSection("functions", "Functions", b.funcItems("func", p.Funcs))
	members := []*section{consts, sentinels, vars, interfaces, types, funcs}
	for _, s := range append([]*section{examples}, members...) {
		pg.Sections = appendSection(pg.Sections, s)
		pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(s))
	}
	if !hasTables && !slices.ContainsFunc(members, func(s *section) bool { return s != nil }) {
		pg.Sections = append(pg.Sections, section{ID: "empty", Kind: kindMessage, Message: noExportedMessage})
	}
}

// partition divides items into those accepted by match and the rest, each
// in the given order.
func partition[T any](items []T, match func(T) bool) (matched, rest []T) {
	for _, it := range items {
		if match(it) {
			matched = append(matched, it)
		} else {
			rest = append(rest, it)
		}
	}
	return matched, rest
}

// toolsSection lists the main packages of mod.
func (b *builder) toolsSection(mod *model.Module) *section {
	var rows []tableRow
	for _, p := range mod.Packages {
		if p.Tool() {
			rows = append(rows, b.packageRow(p, p.DisplayName()))
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return &section{ID: "tools", Title: "Tools", Kind: kindTable, Rows: rows}
}

// packagesSection lists the non-main packages of mod below rel.
func (b *builder) packagesSection(mod *model.Module, rel string) *section {
	prefix := ""
	if rel != "" {
		prefix = rel + "/"
	}
	var rows []tableRow
	for _, p := range mod.Packages {
		if p.Tool() || p.RelPath == "" || !strings.HasPrefix(p.RelPath, prefix) {
			continue
		}
		rows = append(rows, b.packageRow(p, strings.TrimPrefix(p.RelPath, prefix)))
	}
	if len(rows) == 0 {
		return nil
	}
	return &section{ID: "packages", Title: "Packages", Kind: kindTable, Rows: rows}
}

func (b *builder) packageRow(p *model.Package, name string) tableRow {
	scoped := b.withPackage(p)
	summary, _ := scoped.summaryAndFull(scoped.packageDoc(p))
	return tableRow{
		Name:       name,
		Href:       b.rel(pathmap.Package(p.Module.Path, p.RelPath)),
		Internal:   p.Internal(),
		Deprecated: p.Deprecated(),
		Summary:    summary,
	}
}

// withPackage returns a copy of the builder that resolves doc links in the
// scope of p.
func (b *builder) withPackage(p *model.Package) *builder {
	c := *b
	c.pkg = p
	return &c
}

// typePage builds the page of a type.
func (b *builder) typePage(t *model.Type) *page {
	pg := b.newPage(t.Pkg.ImportPath+"."+t.Name, t.Pkg.Module)
	pg.Breadcrumb = b.breadcrumb(t.Pkg.Module, t.Pkg, crumb{Text: t.Name, Href: ""})
	pg.Heading = heading{Kind: t.Kind.String(), Name: t.Name, Internal: t.Pkg.Internal(), Unexported: !t.Exported(), Deprecated: t.Deprecated()}
	pg.SourceHref = b.sourceHref(t.Spec)
	pg.Badges = props.Badges(t)
	pg.Definition = b.code(b.printer("").Type(t))

	idx := b.r.index
	sections := []*section{
		b.docSection(b.doc(t.Doc)),
		b.examplesSection(t.Examples),
		itemsSection("fields", "Fields", b.fieldItems(t.Fields)),
		itemsSection("instances", "Instances", b.valueItems("instance", idx.Instances(t))),
		b.funcGroupsSection("constructors", "Constructors", "ctor", idx.Constructors(t)),
	}
	if t.Kind != model.KindInterface {
		sections = append(sections, itemsSection("methods", "Methods", b.funcItems("method", t.Methods)))
	}
	sections = append(sections,
		b.funcGroupsSection("utilities", "Utilities", "util", idx.Utilities(t)),
		b.relationSection("implements", "Implements", relate.Groups(idx.Implements(t), t), t),
		b.relationSection("implementations", "Implementations", relate.Groups(idx.Implementations(t), t), t),
	)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(b.doc(t.Doc))))
	for _, s := range sections {
		pg.Sections = appendSection(pg.Sections, s)
		if s != nil && s.ID != "documentation" {
			pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(s))
		}
	}
	return pg
}

// funcPage builds the page of a function or method.
func (b *builder) funcPage(f *model.Func) *page {
	var symbols []crumb
	title := f.Pkg.ImportPath + "." + f.Name
	if f.Recv != nil {
		symbols = append(symbols, crumb{
			Text: f.Recv.Name,
			Href: b.rel(pathmap.Symbol(f.Pkg.Module.Path, f.Pkg.RelPath, f.Recv.Name)),
		})
		title = f.Pkg.ImportPath + "." + f.Recv.Name + "." + f.Name
	}
	symbols = append(symbols, crumb{Text: f.Name})
	pg := b.newPage(title, f.Pkg.Module)
	pg.Breadcrumb = b.breadcrumb(f.Pkg.Module, f.Pkg, symbols...)
	pg.Heading = heading{Kind: "func", Name: f.Name, Internal: f.Pkg.Internal(), Unexported: !f.Exported(), Deprecated: f.Deprecated()}
	pg.SourceHref = b.sourceHref(f.Decl)
	pg.Definition = b.code(b.printer("").Func(f))
	b.addDocAndExamples(pg, f.Doc, f.Examples)
	return pg
}

// valuePage builds the page of a constant or variable.
func (b *builder) valuePage(v *model.Value) *page {
	pg := b.newPage(v.Pkg.ImportPath+"."+v.Name, v.Pkg.Module)
	pg.Breadcrumb = b.breadcrumb(v.Pkg.Module, v.Pkg, crumb{Text: v.Name})
	pg.Heading = heading{Kind: v.Kind.String(), Name: v.Name, Internal: v.Pkg.Internal(), Unexported: !v.Exported(), Deprecated: v.Deprecated()}
	if v.Spec != nil {
		pg.SourceHref = b.sourceHref(v.Spec)
	}
	pg.Definition = b.code(b.printer("").Value(v))
	b.addDocAndExamples(pg, v.Doc, v.Examples)
	return pg
}

func (b *builder) addDocAndExamples(pg *page, text string, exs []*model.Example) {
	doc := b.doc(text)
	examples := b.examplesSection(exs)
	pg.Sections = appendSection(pg.Sections, b.docSection(doc))
	pg.Sections = appendSection(pg.Sections, examples)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(examples))
}

// sourcePage builds the page of a source file. It returns any error from
// the highlighter.
func (b *builder) sourcePage(p *model.Package, file *model.File, src []byte) (*page, error) {
	var out strings.Builder
	if err := b.r.highlight.Source(&out, src); err != nil {
		return nil, err
	}
	pg := b.newPage(p.ImportPath+"/"+file.Name, p.Module)
	pg.Breadcrumb = b.breadcrumb(p.Module, p, crumb{Text: file.Name, Separator: "/"})
	pg.Heading = heading{Kind: "file", Name: file.Name, Internal: p.Internal()}
	pg.Sections = []section{{ID: "source", Kind: kindRaw, HTML: template.HTML(out.String())}}
	pg.Sidebar = []sidebarSection{{Title: "Files", Items: b.fileItems(p)}}
	return pg, nil
}

func (b *builder) fileItems(p *model.Package) []sidebarItem {
	var items []sidebarItem
	for _, f := range p.Files {
		items = append(items, sidebarItem{
			Text: f.Name,
			Href: b.rel(pathmap.Source(p.Module.Path, p.RelPath, f.Name)),
		})
	}
	return items
}

// packageTree turns the package rows into a tree under root. The root is
// open and every package node is collapsed.
func packageTree(root *treeNode, rows []tableRow) []*treeNode {
	root.Open = true
	for _, row := range rows {
		level := &root.Children
		var node *treeNode
		for elem := range strings.SplitSeq(row.Name, "/") {
			node = findNode(*level, elem)
			if node == nil {
				node = &treeNode{Text: elem}
				*level = append(*level, node)
			}
			level = &node.Children
		}
		node.Href = row.Href
		node.Internal = row.Internal
		node.Deprecated = row.Deprecated
	}
	return []*treeNode{root}
}

func findNode(nodes []*treeNode, text string) *treeNode {
	for _, n := range nodes {
		if n.Text == text {
			return n
		}
	}
	return nil
}

// treeSidebar returns the sidebar entry of a package table as a tree rooted
// at the page package. p is nil for a module without a root package.
func (b *builder) treeSidebar(s *section, mod *model.Module, p *model.Package) *sidebarSection {
	if s == nil {
		return nil
	}
	root := &treeNode{Text: path.Base(mod.Path), Href: b.rel(pathmap.Module(mod.Path))}
	if p != nil {
		root.Text = p.DisplayName()
		root.Href = b.rel(pathmap.Package(mod.Path, p.RelPath))
		root.Internal = p.Internal()
		root.Deprecated = p.Deprecated()
	}
	return &sidebarSection{Title: s.Title, Tree: packageTree(root, s.Rows)}
}
