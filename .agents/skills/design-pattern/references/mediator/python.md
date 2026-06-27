# Mediator

Define an object that encapsulates how a set of objects interact. Mediator promotes loose coupling by keeping objects from referring to each other explicitly.

## Intent

- Encapsulate interaction between objects
- Decouple objects from knowing about each other
- Centralize control logic

## Implementation

```python
from abc import ABC, abstractmethod


class ChatMediator(ABC):
    @abstractmethod
    def send_message(self, message: str, sender: 'User') -> None:
        pass
    
    @abstractmethod
    def add_user(self, user: 'User') -> None:
        pass


class ChatRoom(ChatMediator):
    users: list
    
    def __init__(self) -> None:
        self.users = []
    
    def send_message(self, message: str, sender: 'User') -> None:
        for user in self.users:
            if user != sender:
                user.receive_message(message, sender.name)
    
    def add_user(self, user: 'User') -> None:
        self.users.append(user)


class User(object):
    name: str
    mediator: ChatMediator
    
    def __init__(self, name: str, mediator: ChatMediator) -> None:
        self.name = name
        self.mediator = mediator
    
    def send(self, message: str) -> None:
        print(f"{self.name} sends: {message}")
        self.mediator.send_message(message, self)
    
    def receive_message(self, message: str, sender: str) -> None:
        print(f"{self.name} received from {sender}: {message}")


if __name__ == "__main__":
    mediator = ChatRoom()
    
    user1 = User("Alice", mediator)
    user2 = User("Bob", mediator)
    user3 = User("Charlie", mediator)
    
    mediator.add_user(user1)
    mediator.add_user(user2)
    mediator.add_user(user3)
    
    user1.send("Hello everyone!")
```

## When to Use

- When you have a set of objects with complex communication
- When you want to reduce coupling between objects
- When you want to centralize complex interaction logic
