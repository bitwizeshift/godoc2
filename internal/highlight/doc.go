// Package highlight renders Go source as syntax-highlighted HTML with chroma.
//
// [Highlighter.Code] renders a declaration from the sig package and wraps the
// linked identifier ranges in anchors. [Highlighter.Source] renders a whole
// file with linkable line numbers, and [Highlighter.CSS] writes the style
// rules that both outputs use, for the light and the dark theme.
package highlight
