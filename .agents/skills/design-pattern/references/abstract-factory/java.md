# Abstract Factory

Provide an interface for creating families of related objects without specifying concrete classes.

## Intent

- Provide an interface for creating families of related objects
- Ensure that created objects are compatible with each other
- Isolate concrete implementations from client code

## Structure

```
AbstractFactory
    ├── createProductA() ──> AbstractProductA
    └── createProductB() ──> AbstractProductB
           │                        │
           │                        │
    ConcreteFactory1         ConcreteProductA1
    ConcreteFactory2         ConcreteProductA2
```

## Implementation

```java
// Abstract products
public interface Button {
    void render();
}

public interface TextField {
    void render();
}

// Abstract factory
public interface UIFactory {
    Button createButton();
    TextField createTextField();
}

// Concrete products - Windows
public class WindowsButton implements Button {
    @Override
    public void render() {
        System.out.println("Rendering Windows button");
    }
}

public class WindowsTextField implements TextField {
    @Override
    public void render() {
        System.out.println("Rendering Windows text field");
    }
}

// Concrete products - Mac
public class MacButton implements Button {
    @Override
    public void render() {
        System.out.println("Rendering Mac button");
    }
}

public class MacTextField implements TextField {
    @Override
    public void render() {
        System.out.println("Rendering Mac text field");
    }
}

// Concrete factories
public class WindowsFactory implements UIFactory {
    @Override
    public Button createButton() {
        return new WindowsButton();
    }
    
    @Override
    public TextField createTextField() {
        return new WindowsTextField();
    }
}

public class MacFactory implements UIFactory {
    @Override
    public Button createButton() {
        return new MacButton();
    }
    
    @Override
    public TextField createTextField() {
        return new MacTextField();
    }
}

// Client code - depends on abstraction
public class UIRenderer {
    public void renderUI(UIFactory factory) {
        Button button = factory.createButton();
        TextField textField = factory.createTextField();
        
        button.render();
        textField.render();
    }
}

// Usage
public class Main {
    public static void main(String[] args) {
        UIRenderer renderer = new UIRenderer();
        
        // On Mac
        UIFactory macFactory = new MacFactory();
        renderer.renderUI(macFactory);
        
        // On Windows
        UIFactory windowsFactory = new WindowsFactory();
        renderer.renderUI(windowsFactory);
    }
}
```

## When to Use

- When your system needs to be independent of how its products are created
- When you need to work with various families of related products
- When you want to provide a library without exposing implementation details
