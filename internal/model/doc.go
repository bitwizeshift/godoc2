// Package model holds the documentation model of a Go module.
//
// The model is produced by the loader and consumed by the page renderers. It
// keeps references to the [go/doc], [go/ast], and [go/types] objects of the
// loaded packages instead of copies, so that a single loaded module can be
// documented without duplicating its data.
package model
