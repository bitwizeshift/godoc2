// Package doclink is a goldmark extension for Go doc links.
//
// A doc link is a bracketed identifier such as [Name], [Name.Method],
// [pkg.Name], [pkg.Name.Method], [*Name], or [net/http]. The parser resolves
// each link through the [Scope] stored in the parse context with
// [NewContext]. Links that do not resolve stay as literal text, and links
// that have a Markdown link definition are left to the standard link parser.
package doclink
