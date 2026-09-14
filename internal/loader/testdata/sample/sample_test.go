package sample_test

import (
	"fmt"

	"example.com/sample"
)

func Example() {
	fmt.Println(sample.Version)
	// Output: 1.0
}

func ExampleCircle() {
	c := sample.Circle{Radius: 2}
	fmt.Println(c.Name())
	// Output: circle
}

func ExampleNewCircle() {
	c := sample.NewCircle(1)
	fmt.Println(c.Name())
	// Output: circle
}

// ExampleCircle_Area shows the area of a unit circle.
func ExampleCircle_Area() {
	c := sample.NewCircle(1)
	fmt.Printf("%.2f\n", c.Area())
	// Output: 3.14
}

func ExampleSquare_String_zero() {
	fmt.Println(sample.Square{})
	// Output: 0
}
