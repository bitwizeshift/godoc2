// Package alpha is the first module of the multi fixture.
//
// It declares one type that the beta module wraps.
package alpha

// Unit is a value with a fixed name.
type Unit struct{}

// Name returns the name of the unit.
func (Unit) Name() string {
	return "unit"
}

// Namer is a value with a name.
type Namer interface {
	Name() string
}
