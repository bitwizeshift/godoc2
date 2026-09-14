// Package model holds the documentation model of a site of Go modules.
//
// The model is produced by the loader and consumed by the page renderers. It
// keeps references to the [go/doc], [go/ast], and [go/types] objects of the
// loaded packages instead of copies, so that the loaded modules can be
// documented without duplicating their data.
package model
