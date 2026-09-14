package shapes

import "example.com/sample"

// NewUnit returns a circle with radius 1.
func NewUnit() *sample.Circle {
	return sample.NewCircle(1)
}

// Perimeter returns the perimeter of c.
func Perimeter(c *sample.Circle) float64 {
	return 2 * 3.14159 * c.Radius
}
