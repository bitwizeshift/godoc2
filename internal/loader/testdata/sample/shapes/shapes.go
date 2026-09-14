// Package shapes declares interfaces satisfied by shapes.
package shapes

// Sizer is anything with a surface area.
type Sizer interface {
	Area() float64
}

// Nothing is an empty interface.
type Nothing any
