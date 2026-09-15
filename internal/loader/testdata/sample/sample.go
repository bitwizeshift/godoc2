package sample

import (
	"fmt"
	"io"
	"sync"
)

// Shape is a sealed interface for geometric shapes.
type Shape interface {
	// Area returns the surface area.
	Area() float64

	sealed()
}

// Named is anything with a name.
type Named interface {
	Name() string
}

// Color is a colour in the RGB palette.
type Color int

// Colours of the palette.
const (
	Red Color = iota
	Green
	Blue
)

// Version is the fixture version.
const Version = "1.0"

// DefaultColor is the colour used when none is given.
var DefaultColor = Green

// ErrNegative is returned for negative sizes.
var ErrNegative = fmt.Errorf("negative size")

// Circle is a round shape.
//
// It implements [Shape] and [Named] through pointer receivers.
type Circle struct {
	// Radius is the circle radius.
	Radius float64

	// Color is the fill colour.
	Color Color

	hidden int
}

// NewCircle returns a circle with the given radius.
func NewCircle(radius float64) *Circle {
	return &Circle{Radius: radius}
}

// Area returns the surface area.
func (c *Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

// Name returns the shape name.
func (c *Circle) Name() string {
	return "circle"
}

func (c *Circle) sealed() {}

// Describe writes a description of c to w. It returns any error from w.
func Describe(w io.Writer, c *Circle) error {
	_, err := fmt.Fprintf(w, "%s %v", c.Name(), c.Radius)
	return err
}

// Square is a four-sided shape with value receivers.
type Square struct {
	// Side is the side length.
	Side float64
}

// Name returns the shape name.
func (s Square) Name() string {
	return "square"
}

// String returns the side length.
func (s Square) String() string {
	return fmt.Sprint(s.Side)
}

// ID is an alias for string identifiers.
type ID = string

// Stack is a generic last-in first-out container.
type Stack[T any] struct {
	items []T
}

// Push adds v to the top of the stack.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top of the stack. It returns false when the
// stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Counter is a mutex-protected count.
type Counter struct {
	mu sync.Mutex
	n  int
}

// Add adds n to the count.
func (c *Counter) Add(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += n
}

// Grid holds counters.
type Grid struct {
	// Cells are the counters.
	Cells [4]Counter
}

// Big is a large value type.
type Big struct {
	// Data is the payload.
	Data [128]byte
}

// Configure applies many settings and returns a circle. It has a long
// signature.
func Configure(radius float64, color Color, name string, writer io.Writer, extra map[string][]int) (*Circle, error) {
	return &Circle{Radius: radius, Color: color}, nil
}

func unexported() {}

// Legacy is the version before [Version] existed.
//
// Deprecated: Use [Version] instead. Legacy is kept only so that old
// callers still compile.
const Legacy = "0.9"

// Label is text placed next to a shape.
type Label struct {
	io.Reader

	// Text is the label text.
	//
	// It is drawn in [Color] Red.
	Text string

	Width, Height int // Width and Height are the size of the label box.

	Font string `json:"font"`

	// Legacy is the old label text.
	//
	// Deprecated: Use Text instead.
	Legacy string

	hidden bool
}

// ZeroCircle returns a circle with radius 0 by value.
func ZeroCircle() Circle {
	return Circle{}
}
