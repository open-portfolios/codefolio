# Visitor

Represent an operation to be performed on elements of an object structure. Visitor lets you define a new operation without changing the classes of the elements.

## Intent

- Define a new operation without changing the classes of the elements
- Separate algorithm from object structure
- Add new operations to existing objects

## Implementation

```go
package main

import "fmt"

type Visitor interface {
	VisitCircle(circle *Circle)
	VisitRectangle(rectangle *Rectangle)
}

type Shape interface {
	Accept(visitor Visitor)
}

type Circle struct {
	Radius float64
}

func NewCircle(radius float64) *Circle {
	return &Circle{Radius: radius}
}

func (c *Circle) Accept(visitor Visitor) {
	visitor.VisitCircle(c)
}

type Rectangle struct {
	Width  float64
	Height float64
}

func NewRectangle(width, height float64) *Rectangle {
	return &Rectangle{Width: width, Height: height}
}

func (r *Rectangle) Accept(visitor Visitor) {
	visitor.VisitRectangle(r)
}

type AreaCalculator struct{}

func NewAreaCalculator() *AreaCalculator {
	return &AreaCalculator{}
}

func (v *AreaCalculator) VisitCircle(circle *Circle) {
	area := 3.14 * circle.Radius * circle.Radius
	fmt.Println("Circle area:", area)
}

func (v *AreaCalculator) VisitRectangle(rectangle *Rectangle) {
	area := rectangle.Width * rectangle.Height
	fmt.Println("Rectangle area:", area)
}

type PerimeterCalculator struct{}

func NewPerimeterCalculator() *PerimeterCalculator {
	return &PerimeterCalculator{}
}

func (v *PerimeterCalculator) VisitCircle(circle *Circle) {
	perimeter := 2 * 3.14 * circle.Radius
	fmt.Println("Circle perimeter:", perimeter)
}

func (v *PerimeterCalculator) VisitRectangle(rectangle *Rectangle) {
	perimeter := 2 * (rectangle.Width + rectangle.Height)
	fmt.Println("Rectangle perimeter:", perimeter)
}

func main() {
	shapes := []Shape{
		NewCircle(5),
		NewRectangle(4, 6),
	}
	
	areaCalculator := NewAreaCalculator()
	perimeterCalculator := NewPerimeterCalculator()
	
	for _, shape := range shapes {
		shape.Accept(areaCalculator)
		shape.Accept(perimeterCalculator)
	}
}
```

## When to Use

- When you have a structure of objects and want to perform operations on them
- When you want to add new operations without changing the object classes
- When you need to perform operations across different types of objects
