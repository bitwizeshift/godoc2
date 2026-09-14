// Package emit abstracts the destination of generated files.
//
// A [Sink] creates one writable file at a time. Page generators write each
// page straight into the file returned by [Sink.Create] and close it before
// they start the next page, which keeps the memory use of the tool small.
package emit
