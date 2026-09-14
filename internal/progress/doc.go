// Package progress reports the generation stages to the user.
//
// A [Reporter] receives the name of each stage as it starts, and the path of
// each file as it is written. The [Writer] implementation prints stages
// always and prints files only when verbose output is enabled.
package progress
