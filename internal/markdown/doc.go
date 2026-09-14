// Package markdown renders Go doc comments as HTML.
//
// Doc comments are parsed as GitHub Flavored Markdown with the doclink
// extension, so that bracketed Go identifiers link to their documentation.
// A parsed [Document] yields the full HTML, the first paragraph as a summary,
// and the list of headings for navigation.
package markdown
