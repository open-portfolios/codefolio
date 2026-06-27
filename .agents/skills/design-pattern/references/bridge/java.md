# Bridge

Separate an abstraction from its implementation so that the two can vary independently.

## Intent

- Separate an abstraction from its implementation so that the two can vary independently
- Avoid permanent binding between an abstraction and its implementation
- Both can be extended independently

## Structure

```
Abstraction
    │
    └── implementation ──> Implementor
                               │
                               └── ConcreteImplementor
```

## Implementation

```java
// Implementor
public interface Renderer {
    void renderCircle(float radius);
    void renderSquare(float side);
}

// Concrete implementors
public class VectorRenderer implements Renderer {
    @Override
    public void renderCircle(float radius) {
        System.out.println("Drawing circle with radius " + radius);
    }
    
    @Override
    public void renderSquare(float side) {
        System.out.println("Drawing square with side " + side);
    }
}

public class RasterRenderer implements Renderer {
    @Override
    public void renderCircle(float radius) {
        System.out.println("Drawing pixels: circle with radius " + radius);
    }
    
    @Override
    public void renderSquare(float side) {
        System.out.println("Drawing pixels: square with side " + side);
    }
}

// Abstraction
public abstract class Shape {
    protected Renderer renderer;
    
    protected Shape(Renderer renderer) {
        this.renderer = renderer;
    }
    
    public abstract void draw();
    public abstract void resize(float factor);
}

// Refined abstractions
public class Circle extends Shape {
    private float radius;
    
    public Circle(Renderer renderer, float radius) {
        super(renderer);
        this.radius = radius;
    }
    
    @Override
    public void draw() {
        renderer.renderCircle(radius);
    }
    
    @Override
    public void resize(float factor) {
        radius *= factor;
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        // Vector rendering
        Shape circle = new Circle(new VectorRenderer(), 5);
        circle.draw();
        
        // Raster rendering
        Shape circle2 = new Circle(new RasterRenderer(), 5);
        circle2.draw();
    }
}
```

## When to Use

- When you want to avoid permanent binding between an abstraction and its implementation
- When both the abstractions and implementations should be extensible
- When changes in implementation should not affect clients
- When you have many classes that differ only in their implementation
