# Chain of Responsibility

Pass a request along a chain of handlers. Each handler decides to process the request or pass it to the next handler.

## Intent

- Pass a request along a chain of handlers
- Each handler handles the request or forwards it to the next handler
- Decouple sender and receiver of a request

## Implementation

```java
public abstract class Handler {
    private Handler next;
    
    public Handler setNext(Handler handler) {
        this.next = handler;
        return handler;
    }
    
    public final void handle(String request) {
        if (process(request)) {
            return;
        }
        if (next != null) {
            next.handle(request);
        }
    }
    
    protected abstract boolean process(String request);
}

public class AuthHandler extends Handler {
    @Override
    protected boolean process(String request) {
        if (request.startsWith("auth:")) {
            System.out.println("Authentication handled");
            return true;
        }
        return false;
    }
}

public class ValidationHandler extends Handler {
    @Override
    protected boolean process(String request) {
        if (request.startsWith("valid:")) {
            System.out.println("Validation handled");
            return true;
        }
        return false;
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        Handler authHandler = new AuthHandler();
        Handler validationHandler = new ValidationHandler();
        
        authHandler.setNext(validationHandler);
        
        authHandler.handle("auth:login");
        authHandler.handle("valid:data");
        authHandler.handle("other:request");
    }
}
```

## When to Use

- When more than one object may handle a request
- When you want to issue a request to one of several objects without specifying the receiver explicitly
- When you want to set up a chain of handlers dynamically
