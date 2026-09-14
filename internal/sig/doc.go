// Package sig prints cleaned-up declarations with link spans.
//
// A [Printer] renders a function, type, or value declaration as plain Go
// text and records the byte ranges of identifiers that resolve to a
// documentation URL. Long function signatures are broken so that each
// parameter sits on its own line. Struct fields and interface methods are
// separated by blank lines, and unexported members are replaced by a comment.
package sig
