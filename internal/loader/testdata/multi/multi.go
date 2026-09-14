// Package multi is the root module of the multi fixture.
//
// The alpha and beta modules sit below it in the same directory.
package multi

import "example.com/multi/alpha"

// Name is the name of the fixture.
const Name = "multi"

// Root is the value the root module names.
type Root struct{}

// Name returns the name of the fixture.
func (Root) Name() string {
	return Name
}

// NewUnit returns a zero unit.
func NewUnit() alpha.Unit {
	return alpha.Unit{}
}
