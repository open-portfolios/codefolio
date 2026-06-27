# Visitor

Represent an operation to be performed on elements of an object structure. Visitor lets you define a new operation without changing the classes of the elements.

## Intent

- Define a new operation without changing the classes of the elements
- Separate algorithm from object structure
- Add new operations to existing objects

## Implementation

```java
public interface Visitor {
    void visitCircle(Circle circle);
    void visitRectangle(Rectangle rectangle);
}

public interface Shape {
    void accept(Visitor visitor);
}

public class Circle implements Shape {
    private double radius;
    
    public Circle(double radius) {
        this.radius = radius;
    }
    
    public double getRadius() {
        return radius;
    }
    
    @Override
    public void accept(Visitor visitor) {
        visitor.visitCircle(this);
    }
}

public class Rectangle implements Shape {
    private double width;
    private double height;
    
    public Rectangle(double width, double height) {
        this.width = width;
        this.height = height;
    }
    
    public double getWidth() {
        return width;
    }
    
    public double getHeight() {
        return height;
    }
    
    @Override
    public void accept(Visitor visitor) {
        visitor.visitRectangle(this);
    }
}

public class AreaCalculator implements Visitor {
    @Override
    public void visitCircle(Circle circle) {
        double area = Math.PI * circle.getRadius() * circle.getRadius();
        System.out.println("Circle area: " + area);
    }
    
    @Override
    public void visitRectangle(Rectangle rectangle) {
        double area = rectangle.getWidth() * rectangle.getHeight();
        System.out.println("Rectangle area: " + area);
    }
}

public class PerimeterCalculator implements Visitor {
    @Override
    public void visitCircle(Circle circle) {
        double perimeter = 2 * Math.PI * circle.getRadius();
        System.out.println("Circle perimeter: " + perimeter);
    }
    
    @Override
    public void visitRectangle(Rectangle rectangle) {
        double perimeter = 2 * (rectangle.getWidth() + rectangle.getHeight());
        System.out.println("Rectangle perimeter: " + perimeter);
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        List<Shape> shapes = Arrays.asList(
            new Circle(5),
            new Rectangle(4, 6)
        );
        
        Visitor areaCalculator = new AreaCalculator();
        Visitor perimeterCalculator = new PerimeterCalculator();
        
        for (Shape shape : shapes) {
            shape.accept(areaCalculator);
            shape.accept(perimeterCalculator);
        }
    }
}
```

## When to Use

- When you have a structure of objects and want to perform operations on them
- When you want to add new operations without changing the object classes
- When you need to perform operations across different types of objects
