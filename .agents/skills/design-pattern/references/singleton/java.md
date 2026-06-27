# Singleton

Ensure a class has only one instance and provide a global point of access to it.

## Intent

- Ensure a class has only one instance
- Provide a global point of access to that instance
- Control object creation at a single point

## Implementation

### Thread-Safe Double-Cchecked Locking

```java
public class Singleton {
    private static volatile Singleton instance;
    
    private Singleton() {
        // Private constructor prevents instantiation
    }
    
    public static Singleton getInstance() {
        if (instance == null) {
            synchronized (Singleton.class) {
                if (instance == null) {
                    instance = new Singleton();
                }
            }
        }
        return instance;
    }
}
```

### Bill Pugh Singleton (Initialization-on-demand holder)

```java
public class Singleton {
    private Singleton() {}
    
    private static class SingletonHolder {
        private static final Singleton INSTANCE = new Singleton();
    }
    
    public static Singleton getInstance() {
        return SingletonHolder.INSTANCE;
    }
}
```

### Enum Singleton

```java
public enum Singleton {
    INSTANCE;
    
    public void doSomething() {
        // implementation
    }
}
```

## When to Use

- When exactly one instance of a class is needed
- When you need global access to that instance
- Common use cases: configuration managers, logging, connection pools

## Considerations

- Consider dependency injection instead of Singleton when possible
- Singletons can make testing difficult
- Be careful with thread safety in multi-threaded environments
