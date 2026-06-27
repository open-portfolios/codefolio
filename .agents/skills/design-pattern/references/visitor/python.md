# Visitor

Represent an operation to be performed on elements of an object structure. Visitor lets you define a new operation without changing the classes of the elements.

## Intent

- Define a new operation without changing the classes of the elements
- Separate algorithm from object structure
- Add new operations to existing objects

## Implementation

```python
from abc import ABC, abstractmethod


class Visitor(ABC):
    @abstractmethod
    def visit_circle(self, circle: 'Circle') -> None:
        pass
    
    @abstractmethod
    def visit_rectangle(self, rectangle: 'Rectangle') -> None:
        pass


class Shape(ABC):
    @abstractmethod
    def accept(self, visitor: Visitor) -> None:
        pass


class Circle(Shape):
    radius: float
    
    def __init__(self, radius: float) -> None:
        self.radius = radius
    
    def accept(self, visitor: Visitor) -> None:
        visitor.visit_circle(self)


class Rectangle(Shape):
    width: float
    height: float
    
    def __init__(self, width: float, height: float) -> None:
        self.width = width
        self.height = height
    
    def accept(self, visitor: Visitor) -> None:
        visitor.visit_rectangle(self)


class AreaCalculator(Visitor):
    def visit_circle(self, circle: Circle) -> None:
        area = 3.14 * circle.radius * circle.radius
        print(f"Circle area: {area}")
    
    def visit_rectangle(self, rectangle: Rectangle) -> None:
        area = rectangle.width * rectangle.height
        print(f"Rectangle area: {area}")


class PerimeterCalculator(Visitor):
    def visit_circle(self, circle: Circle) -> None:
        perimeter = 2 * 3.14 * circle.radius
        print(f"Circle perimeter: {perimeter}")
    
    def visit_rectangle(self, rectangle: Rectangle) -> None:
        perimeter = 2 * (rectangle.width + rectangle.height)
        print(f"Rectangle perimeter: {perimeter}")


if __name__ == "__main__":
    shapes = [
        Circle(5),
        Rectangle(4, 6)
    ]
    
    area_calculator = AreaCalculator()
    perimeter_calculator = PerimeterCalculator()
    
    for shape in shapes:
        shape.accept(area_calculator)
        shape.accept(perimeter_calculator)
```

## When to Use

- When you have a structure of objects and want to perform operations on them
- When you want to add new operations without changing the object classes
- When you need to perform operations across different types of objects
