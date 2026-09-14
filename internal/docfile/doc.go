// Package docfile finds the Markdown files that document a directory in
// place of a package doc comment, and maps the relative links of those files
// to the generated site.
//
// [Find] locates the file of one directory. The loader calls it for every
// package without a doc comment and for a module root without a package.
//
// [Resolver] rewrites the link destinations of a found file. Links to package
// directories and Go source files point at the generated pages. Links to
// other files inside the module point at a copy of the file, which the
// resolver records as an [Asset] for the generator to write.
package docfile
