// Package inner is a second package of the alpha module.
package inner

import "example.com/multi/alpha"

// Inner is a value with a fixed name.
type Inner struct{}

// Name returns the name of the value.
func (Inner) Name() string {
	return "inner"
}

// NewUnit returns a zero unit.
func NewUnit() alpha.Unit {
	return alpha.Unit{}
}
