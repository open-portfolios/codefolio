# Prototype

Create new objects by cloning an existing object.

## Intent

- Create new objects by cloning an existing object
- Avoid costly object creation
- Produce objects that are independent of their creation process

## Structure

```
Prototype (interface)
    └── clone() ──> Prototype

ConcretePrototype
    └── implements Prototype
```

## Implementation

### Using Cloneable Interface

```java
public class Document implements Cloneable {
    private String title;
    private String content;
    
    public Document(String title, String content) {
        this.title = title;
        this.content = content;
    }
    
    public void setTitle(String title) {
        this.title = title;
    }
    
    @Override
    public Document clone() {
        try {
            return (Document) super.clone();
        } catch (CloneNotSupportedException e) {
            throw new RuntimeException(e);
        }
    }
}
```

### Using Copy Constructor

```java
public class Document {
    private String title;
    private String content;
    
    public Document(String title, String content) {
        this.title = title;
        this.content = content;
    }
    
    // Copy constructor
    public Document(Document other) {
        this.title = other.title;
        this.content = other.content;
    }
    
    public Document copy() {
        return new Document(this);
    }
}
```

### Using Builder Pattern

```java
public class Document {
    private String title;
    private String content;
    
    public Document(String title, String content) {
        this.title = title;
        this.content = content;
    }
    
    public Document copy() {
        return new Document.Builder()
            .title(this.title)
            .content(this.content)
            .build();
    }
    
    public static class Builder {
        private String title;
        private String content;
        
        public Builder title(String title) {
            this.title = title;
            return this;
        }
        
        public Builder content(String content) {
            this.content = content;
            return this;
        }
        
        public Document build() {
            return new Document(title, content);
        }
    }
}
```

## When to Use

- When creating new objects is expensive (database calls, network requests)
- When you need to avoid subclassing to create objects
- When instances can have one of few possible states
