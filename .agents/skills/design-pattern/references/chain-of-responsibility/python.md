# Chain of Responsibility

Pass a request along a chain of handlers. Each handler decides to process the request or pass it to the next handler.

## Intent

- Pass a request along a chain of handlers
- Each handler handles the request or forwards it to the next handler
- Decouple sender and receiver of a request

## Implementation

```python
from abc import ABC, abstractmethod


class Handler(ABC):
    _next_handler: Handler | None
    
    def __init__(self) -> None:
        self._next_handler = None
    
    def set_next(self, handler: 'Handler') -> 'Handler':
        self._next_handler = handler
        return handler
    
    def handle(self, request: str) -> None:
        if self._process(request):
            return
        if self._next_handler is not None:
            self._next_handler.handle(request)
    
    @abstractmethod
    def _process(self, request: str) -> bool:
        pass


class AuthHandler(Handler):
    def _process(self, request: str) -> bool:
        if request.startswith("auth:"):
            print("Authentication handled")
            return True
        return False


class ValidationHandler(Handler):
    def _process(self, request: str) -> bool:
        if request.startswith("valid:"):
            print("Validation handled")
            return True
        return False


if __name__ == "__main__":
    auth_handler = AuthHandler()
    validation_handler = ValidationHandler()
    
    auth_handler.set_next(validation_handler)
    
    auth_handler.handle("auth:login")
    auth_handler.handle("valid:data")
    auth_handler.handle("other:request")
```

## When to Use

- When more than one object may handle a request
- When you want to issue a request to one of several objects without specifying the receiver explicitly
- When you want to set up a chain of handlers dynamically
