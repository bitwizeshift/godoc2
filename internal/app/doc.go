// Package app binds the godoc2 command line to the generator.
//
// The flag components ([OutputFlags], [ProgressFlags], and [PatternArgs])
// each register their arguments and expose the object built from them. The
// [Builder] collects the components and builds the [Runner] that executes
// the generation.
package app
