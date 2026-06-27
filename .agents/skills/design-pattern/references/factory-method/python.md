# Factory Method

Define an interface for creating an object, but let subclasses decide which class to instantiate.

## Intent

- Define an interface for creating an object, but let subclasses decide which class to instantiate
- Delegate instantiation to subclasses
- Useful when a class cannot anticipate the type of objects it needs to create

## Implementation

```python
from abc import ABC, abstractmethod


# Product interface
class Notification(ABC):
    @abstractmethod
    def send(self, message: str) -> None:
        pass


# Concrete products
class EmailNotification(Notification):
    def send(self, message: str) -> None:
        print(f"Sending email: {message}")


class SMSNotification(Notification):
    def send(self, message: str) -> None:
        print(f"Sending SMS: {message}")


# Creator abstract class
class NotificationFactory(ABC):
    @abstractmethod
    def create_notification(self) -> Notification:
        pass


# Concrete creators
class EmailFactory(NotificationFactory):
    def create_notification(self) -> Notification:
        return EmailNotification()


class SMSFactory(NotificationFactory):
    def create_notification(self) -> Notification:
        return SMSNotification()


# Client code - depends on abstraction
class NotificationService(object):
    def send_notification(self, factory: NotificationFactory, message: str) -> None:
        notification = factory.create_notification()
        notification.send(message)


# Usage
if __name__ == "__main__":
    service = NotificationService()
    
    # Use email factory
    email_factory = EmailFactory()
    service.send_notification(email_factory, "Hello via Email")
    
    # Use SMS factory
    sms_factory = SMSFactory()
    service.send_notification(sms_factory, "Hello via SMS")
```

## When to Use

- When a class cannot anticipate the type of objects it needs to create
- When subclasses should specify the objects they create
- When you want to delegate responsibility to helper subclasses
