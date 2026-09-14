// Package link resolves Go objects to documentation URLs.
//
// Objects declared in a loaded module resolve to generated pages, relative
// to the page that references them. Objects from the standard library and
// from modules outside the site resolve to pkg.go.dev.
package link
