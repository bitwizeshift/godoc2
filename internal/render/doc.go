// Package render writes the HTML pages of the documentation site.
//
// A [Renderer] builds one view model per page from the documentation model
// and executes the embedded templates straight into the destination writer.
// The static assets that every page references are written with
// [Renderer.CSS] and [Renderer.JS].
package render
