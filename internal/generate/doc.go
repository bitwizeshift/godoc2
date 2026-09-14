// Package generate drives the documentation build from loading to the last
// written file.
//
// A [Generator] loads the module, builds the link and relation indexes, and
// then writes the static assets, the module page, every package with its
// symbol and source pages, and finally the search index. Each page is written
// and closed before the next one starts.
package generate
