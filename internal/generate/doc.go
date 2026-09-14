// Package generate drives the documentation build from loading to the last
// written file.
//
// A [Generator] loads the modules, builds the link and relation indexes, and
// then writes the static assets, the root page, each module page with every
// package and its symbol and source pages, the files that documentation
// links to, and finally the search index. Each page is written and closed
// before the next one starts.
package generate
