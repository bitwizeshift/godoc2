// Package docfile finds the Markdown files that document a directory in
// place of a package doc comment, and maps the relative links of those files
// to the generated site.
//
// [Find] locates the file of one directory. The loader calls it for every
// package without a doc comment, for a module root without a package, and
// for the root of a site with several modules.
//
// [Resolver] rewrites the link destinations of a found file. Links to module
// directories, package directories, and Go source files point at the
// generated pages. Links to other files inside a module or the site directory
// point at a copy of the file, which the resolver records as an [Asset] for
// the generator to write.
package docfile
