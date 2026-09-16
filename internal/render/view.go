package render

import (
	"html/template"

	"github.com/bitwizeshift/godoc2/internal/props"
)

// page is the view model of every page.
type page struct {
	Title string

	// HomeHref and HomeTitle are the target and title of the sidebar logo:
	// the module page, or the root page of a site with several modules.
	HomeHref  string
	HomeTitle string

	CSS        string
	JS         string
	IndexJS    string
	Root       string
	Breadcrumb []crumb
	Sidebar    []sidebarSection
	Heading    heading
	SourceHref string

	// Badges are the properties of a type, shown between the heading and the
	// definition.
	Badges     []props.Badge
	Definition template.HTML
	Sections   []section
}

// redirect is the view model of the root page of a single-module site.
type redirect struct {
	Title string
	Href  string
}

// crumb is one element of the breadcrumb path.
type crumb struct {
	Text string
	Href string

	// Badge marks the leading workspace and module badges.
	Badge bool

	// Title is the hover text of a badge. It is empty for a plain element.
	Title string

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
	Text       string
	Href       string
	Internal   bool
	Unexported bool
	Deprecated string
	Open       bool
	Children   []*treeNode
}

// sidebarItem is one link of a sidebar section. Deprecated holds the
// deprecation message of a deprecated target, shown when the badge is
// hovered.
type sidebarItem struct {
	Text       string
	Href       string
	Internal   bool
	Unexported bool
	Deprecated string
}

// heading is the page title. Kind is the keyword before the name, or empty
// for a plain title.
type heading struct {
	Kind       string
	Name       string
	Internal   bool
	Unexported bool
	Deprecated string
}

// Section kinds.
const (
	kindDoc       = "doc"
	kindTable     = "table"
	kindItems     = "items"
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
	Examples []example
	Groups   []relationGroup
	Message  string
}

type tableRow struct {
	Name       string
	Href       string
	Internal   bool
	Unexported bool
	Deprecated string
	Summary    template.HTML
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
	Unexported bool
	Deprecated string

	// Receiver is the form of the implementing type, T or *T, for relation
	// entries.
	Receiver string

	// Embedded marks a field entry declared by its type alone.
	Embedded bool
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
