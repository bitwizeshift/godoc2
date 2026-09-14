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

// Name returns the name of the wrapped unit.
func (w *Wrapper) Name() string {
	return w.Unit.Name()
}

// NewUnit returns a zero unit.
func NewUnit() alpha.Unit {
	return alpha.Unit{}
}
