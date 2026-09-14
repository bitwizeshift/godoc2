// Package lib wraps the unit of the alpha module.
package lib

import "example.com/multi/alpha"

// Wrapper holds a unit.
type Wrapper struct {
	Unit alpha.Unit
}

// New returns a wrapper around a zero unit.
func New() *Wrapper {
	return &Wrapper{}
}
