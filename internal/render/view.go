package render

import (
	"html/template"

	"github.com/bitwizeshift/godoc2/internal/props"
)

// page is the view model of every page.
type page struct {
	Title      string
	ModulePath string
	ModuleHref string
	CSS        string
	JS         string
	IndexJS    string
	Root       string
	Breadcrumb []crumb
	Sidebar    []sidebarSection
	Heading    heading
	SourceHref string
	Definition template.HTML
	Sections   []section
}

// crumb is one element of the breadcrumb path.
type crumb struct {
	Text string
	Href string

	// Module marks the leading module badge.
	Module bool

	// Separator precedes the element: "/" between directories, "." between
	// symbols, or empty for the first element.
	Separator string
}

type sidebarSection struct {
	Title string
	Items []sidebarItem

	// Tree replaces Items with a collapsible tree when it is not empty.
	Tree []*treeNode
}

// treeNode is one node of the sidebar package tree. A node without an Href
// is a directory that holds no package.
type treeNode struct {
	Text     string
	Href     string
	Internal bool
	Open     bool
	Children []*treeNode
}

type sidebarItem struct {
	Text     string
	Href     string
	Internal bool
}

type heading struct {
	Kind     string
	Name     string
	Internal bool
}

// Section kinds.
const (
	kindDoc       = "doc"
	kindTable     = "table"
	kindItems     = "items"
	kindBadges    = "badges"
	kindExamples  = "examples"
	kindRelations = "relations"
	kindMessage   = "message"
	kindRaw       = "raw"
)

// section is one titled block of the main content.
type section struct {
	ID    string
	Title string
	Kind  string

	HTML     template.HTML
	Rows     []tableRow
	Items    []item
	Badges   []props.Badge
	Examples []example
	Groups   []relationGroup
	Message  string
}

type tableRow struct {
	Name     string
	Href     string
	Internal bool
	Summary  template.HTML
}

// item is one code-referencing entry.
type item struct {
	ID         string
	Name       string
	Href       string
	Code       template.HTML
	SourceHref string
	Summary    template.HTML
	Full       template.HTML
	Internal   bool

	// Receiver is the form of the implementing type, T or *T, for relation
	// entries.
	Receiver string
}

type example struct {
	ID     string
	Title  string
	Doc    template.HTML
	Code   template.HTML
	Output string
}

type relationGroup struct {
	Path  string
	Items []item
}
