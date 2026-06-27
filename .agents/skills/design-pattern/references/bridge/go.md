# Bridge

Separate an abstraction from its implementation so that the two can vary independently.

## Intent

- Separate an abstraction from its implementation so that the two can vary independently
- Avoid permanent binding between an abstraction and its implementation
- Both can be extended independently

## Implementation

```go
package main

import "fmt"

// Implementor
type Renderer interface {
	RenderCircle(radius float64)
	RenderSquare(side float64)
}

// Concrete implementors
type VectorRenderer struct{}

func NewVectorRenderer() *VectorRenderer {
	return &VectorRenderer{}
}

func (r *VectorRenderer) RenderCircle(radius float64) {
	fmt.Println("Drawing circle with radius", radius)
}

func (r *VectorRenderer) RenderSquare(side float64) {
	fmt.Println("Drawing square with side", side)
}

type RasterRenderer struct{}

func NewRasterRenderer() *RasterRenderer {
	return &RasterRenderer{}
}

func (r *RasterRenderer) RenderCircle(radius float64) {
	fmt.Println("Drawing pixels: circle with radius", radius)
}

func (r *RasterRenderer) RenderSquare(side float64) {
	fmt.Println("Drawing pixels: square with side", side)
}

// Abstraction
type Shape interface {
	Draw()
	Resize(factor float64)
}

type Circle struct {
	renderer Renderer
	radius   float64
}

func NewCircle(renderer Renderer, radius float64) *Circle {
	return &Circle{renderer: renderer, radius: radius}
}

func (c *Circle) Draw() {
	c.renderer.RenderCircle(c.radius)
}

func (c *Circle) Resize(factor float64) {
	c.radius *= factor
}

// Client code
func main() {
	// Vector rendering
	circle := NewCircle(NewVectorRenderer(), 5)
	circle.Draw()
	
	// Raster rendering
	circle2 := NewCircle(NewRasterRenderer(), 5)
	circle2.Draw()
}
```

## When to Use

- When you want to avoid permanent binding between an abstraction and its implementation
- When both the abstractions and implementations should be extensible
- When changes in implementation should not affect clients
