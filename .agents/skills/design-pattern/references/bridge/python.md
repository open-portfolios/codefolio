# Bridge

Separate an abstraction from its implementation so that the two can vary independently.

## Intent

- Separate an abstraction from its implementation so that the two can vary independently
- Avoid permanent binding between an abstraction and its implementation
- Both can be extended independently

## Implementation

```python
from abc import ABC, abstractmethod


# Implementor
class Renderer(ABC):
    @abstractmethod
    def render_circle(self, radius: float) -> None:
        pass
    
    @abstractmethod
    def render_square(self, side: float) -> None:
        pass


# Concrete implementors
class VectorRenderer(Renderer):
    def render_circle(self, radius: float) -> None:
        print(f"Drawing circle with radius {radius}")
    
    def render_square(self, side: float) -> None:
        print(f"Drawing square with side {side}")


class RasterRenderer(Renderer):
    def render_circle(self, radius: float) -> None:
        print(f"Drawing pixels: circle with radius {radius}")
    
    def render_square(self, side: float) -> None:
        print(f"Drawing pixels: square with side {side}")


# Abstraction
class Shape(ABC):
    renderer: Renderer
    
    def __init__(self, renderer: Renderer) -> None:
        self.renderer = renderer
    
    @abstractmethod
    def draw(self) -> None:
        pass
    
    @abstractmethod
    def resize(self, factor: float) -> None:
        pass


# Refined abstraction
class Circle(Shape):
    radius: float
    
    def __init__(self, renderer: Renderer, radius: float) -> None:
        super().__init__(renderer)
        self.radius = radius
    
    def draw(self) -> None:
        self.renderer.render_circle(self.radius)
    
    def resize(self, factor: float) -> None:
        self.radius *= factor


if __name__ == "__main__":
    # Vector rendering
    circle = Circle(VectorRenderer(), 5)
    circle.draw()
    
    # Raster rendering
    circle2 = Circle(RasterRenderer(), 5)
    circle2.draw()
```

## When to Use

- When you want to avoid permanent binding between an abstraction and its implementation
- When both the abstractions and implementations should be extensible
- When changes in implementation should not affect clients
