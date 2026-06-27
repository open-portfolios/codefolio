# Factory Method

Define an interface for creating an object, but let subclasses decide which class to instantiate.

## Intent

- Define an interface for creating an object, but let subclasses decide which class to instantiate
- Delegate instantiation to subclasses
- Useful when a class cannot anticipate the type of objects it needs to create

## Structure

```
Creator (abstract)
    │
    └── factoryMethod() ──> Product (interface)
                              │
                              └── ConcreteProduct
```

## Implementation

```java
// Product interface
public interface Notification {
    void send(String message);
}

// Concrete products
public class EmailNotification implements Notification {
    @Override
    public void send(String message) {
        System.out.println("Sending email: " + message);
    }
}

public class SMSNotification implements Notification {
    @Override
    public void send(String message) {
        System.out.println("Sending SMS: " + message);
    }
}

// Creator abstract class
public abstract class NotificationFactory {
    protected abstract Notification createNotification();
}

// Concrete creators
public class EmailFactory extends NotificationFactory {
    @Override
    protected Notification createNotification() {
        return new EmailNotification();
    }
}

public class SMSFactory extends NotificationFactory {
    @Override
    protected Notification createNotification() {
        return new SMSNotification();
    }
}

// Client code - depends on abstraction
public class NotificationService {
    public void sendNotification(NotificationFactory factory, String message) {
        Notification notification = factory.createNotification();
        notification.send(message);
    }
}

// Usage
public class Main {
    public static void main(String[] args) {
        NotificationService service = new NotificationService();
        
        // Use email factory
        NotificationFactory emailFactory = new EmailFactory();
        service.sendNotification(emailFactory, "Hello via Email");
        
        // Use SMS factory
        NotificationFactory smsFactory = new SMSFactory();
        service.sendNotification(smsFactory, "Hello via SMS");
    }
}
```

## When to Use

- When a class cannot anticipate the type of objects it needs to create
- When subclasses should specify the objects they create
- When you want to delegate responsibility to helper subclasses
