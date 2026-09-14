package render

import (
	"html/template"
	"path"
	"strings"

	"github.com/bitwizeshift/godoc2/internal/model"
	"github.com/bitwizeshift/godoc2/internal/pathmap"
	"github.com/bitwizeshift/godoc2/internal/props"
	"github.com/bitwizeshift/godoc2/internal/relate"
)

const noExportedMessage = "This package has no exported identifiers."

// modulePage builds the module page, which is the page of the root package
// with the Tools and Packages tables.
func (b *builder) modulePage(root *model.Package) *page {
	pg := b.newPage(b.r.module.Path)
	pg.Breadcrumb = b.breadcrumb(root)
	pg.Heading = heading{Kind: "module", Name: b.r.module.Path}
	if root != nil {
		pg.Heading = heading{Kind: "package", Name: root.DisplayName()}
		pg.Title = root.ImportPath
	}
	var doc string
	if root != nil {
		doc = root.Doc
	}
	tools := b.toolsSection()
	packages := b.packagesSection("")
	pg.Sections = appendSection(pg.Sections, b.docSection(doc))
	pg.Sections = appendSection(pg.Sections, tools)
	pg.Sections = appendSection(pg.Sections, packages)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(tools))
	pg.Sidebar = appendSidebar(pg.Sidebar, b.treeSidebar(packages, root))
	if root != nil {
		b.addPackageMembers(pg, root, tools != nil || packages != nil)
	}
	return pg
}

// packagePage builds the page of a package below the module root.
func (b *builder) packagePage(p *model.Package) *page {
	pg := b.newPage(p.ImportPath)
	pg.Breadcrumb = b.breadcrumb(p)
	pg.Heading = heading{Kind: "package", Name: p.DisplayName(), Internal: p.Internal()}
	if p.Tool() {
		pg.Heading.Kind = "binary"
	}
	packages := b.packagesSection(p.RelPath)
	pg.Sections = appendSection(pg.Sections, b.docSection(p.Doc))
	pg.Sections = appendSection(pg.Sections, packages)
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(p.Doc)))
	pg.Sidebar = appendSidebar(pg.Sidebar, b.treeSidebar(packages, p))
	b.addPackageMembers(pg, p, packages != nil)
	return pg
}

// addPackageMembers appends the Examples, Constants, Variables, Types, and
// Functions sections of p, or the no-exported-identifiers message when the
// package declares nothing and has no tables.
func (b *builder) addPackageMembers(pg *page, p *model.Package, hasTables bool) {
	examples := b.examplesSection(p.Examples)
	consts := itemsSection("constants", "Constants", b.valueItems("const", p.Consts))
	vars := itemsSection("variables", "Variables", b.valueItems("var", p.Vars))
	types := itemsSection("types", "Types", b.typeItems(p.Types))
	funcs := itemsSection("functions", "Functions", b.funcItems("func", p.Funcs))
	for _, s := range []*section{examples, consts, vars, types, funcs} {
		pg.Sections = appendSection(pg.Sections, s)
		pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(s))
	}
	if consts == nil && vars == nil && types == nil && funcs == nil && !hasTables {
		pg.Sections = append(pg.Sections, section{ID: "empty", Kind: kindMessage, Message: noExportedMessage})
	}
}

// toolsSection lists the main packages of the module.
func (b *builder) toolsSection() *section {
	var rows []tableRow
	for _, p := range b.r.module.Packages {
		if p.Tool() {
			rows = append(rows, b.packageRow(p, p.DisplayName()))
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return &section{ID: "tools", Title: "Tools", Kind: kindTable, Rows: rows}
}

// packagesSection lists the non-main packages below rel.
func (b *builder) packagesSection(rel string) *section {
	prefix := ""
	if rel != "" {
		prefix = rel + "/"
	}
	var rows []tableRow
	for _, p := range b.r.module.Packages {
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
	summary, _ := b.withPackage(p).summaryAndFull(p.Doc)
	return tableRow{
		Name:     name,
		Href:     b.rel(pathmap.Package(b.r.module.Path, p.RelPath)),
		Internal: p.Internal(),
		Summary:  summary,
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
	pg := b.newPage(t.Pkg.ImportPath + "." + t.Name)
	pg.Breadcrumb = b.breadcrumb(t.Pkg, crumb{Text: t.Name, Href: ""})
	pg.Heading = heading{Kind: t.Kind.String(), Name: t.Name, Internal: t.Pkg.Internal()}
	pg.SourceHref = b.sourceHref(t.Spec)
	pg.Definition = b.code(b.printer("").Type(t))

	idx := b.r.index
	sections := []*section{
		b.docSection(t.Doc),
		b.examplesSection(t.Examples),
		badgesSection(props.Badges(t)),
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
	pg.Sidebar = appendSidebar(pg.Sidebar, new(b.docSidebar(t.Doc)))
	for _, s := range sections {
		pg.Sections = appendSection(pg.Sections, s)
		if s != nil && s.ID != "documentation" {
			pg.Sidebar = appendSidebar(pg.Sidebar, sidebarOf(s))
		}
	}
	return pg
}

func badgesSection(badges []props.Badge) *section {
	if len(badges) == 0 {
		return nil
	}
	return &section{ID: "properties", Title: "Properties", Kind: kindBadges, Badges: badges}
}

// funcPage builds the page of a function or method.
func (b *builder) funcPage(f *model.Func) *page {
	var symbols []crumb
	title := f.Pkg.ImportPath + "." + f.Name
	if f.Recv != nil {
		symbols = append(symbols, crumb{
			Text: f.Recv.Name,
			Href: b.rel(pathmap.Symbol(b.r.module.Path, f.Pkg.RelPath, f.Recv.Name)),
		})
		title = f.Pkg.ImportPath + "." + f.Recv.Name + "." + f.Name
	}
	symbols = append(symbols, crumb{Text: f.Name})
	pg := b.newPage(title)
	pg.Breadcrumb = b.breadcrumb(f.Pkg, symbols...)
	pg.Heading = heading{Kind: "func", Name: f.Name, Internal: f.Pkg.Internal()}
	pg.SourceHref = b.sourceHref(f.Decl)
	pg.Definition = b.code(b.printer("").Func(f))
	b.addDocAndExamples(pg, f.Doc, f.Examples)
	return pg
}

// valuePage builds the page of a constant or variable.
func (b *builder) valuePage(v *model.Value) *page {
	pg := b.newPage(v.Pkg.ImportPath + "." + v.Name)
	pg.Breadcrumb = b.breadcrumb(v.Pkg, crumb{Text: v.Name})
	pg.Heading = heading{Kind: v.Kind.String(), Name: v.Name, Internal: v.Pkg.Internal()}
	if v.Spec != nil {
		pg.SourceHref = b.sourceHref(v.Spec)
	}
	pg.Definition = b.code(b.printer("").Value(v))
	b.addDocAndExamples(pg, v.Doc, v.Examples)
	return pg
}

func (b *builder) addDocAndExamples(pg *page, doc string, exs []*model.Example) {
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
	pg := b.newPage(p.ImportPath + "/" + file.Name)
	pg.Breadcrumb = b.breadcrumb(p, crumb{Text: file.Name, Separator: "/"})
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
			Href: b.rel(pathmap.Source(b.r.module.Path, p.RelPath, f.Name)),
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
func (b *builder) treeSidebar(s *section, p *model.Package) *sidebarSection {
	if s == nil {
		return nil
	}
	root := &treeNode{Text: path.Base(b.r.module.Path), Href: b.rel(pathmap.Module(b.r.module.Path))}
	if p != nil {
		root.Text = p.DisplayName()
		root.Href = b.rel(pathmap.Package(b.r.module.Path, p.RelPath))
		root.Internal = p.Internal()
	}
	return &sidebarSection{Title: s.Title, Tree: packageTree(root, s.Rows)}
}
