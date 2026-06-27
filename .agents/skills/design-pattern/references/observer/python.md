# Observer

Define a one-to-many dependency between objects so that when one object changes state, all its dependents are notified.

## Intent

- Define a one-to-many dependency between objects
- Notify dependents when an object changes state
- Loose coupling between subject and observers

## Implementation

```python
from abc import ABC, abstractmethod


class Observer(ABC):
    @abstractmethod
    def update(self, message: str) -> None:
        pass


class Subject(ABC):
    @abstractmethod
    def attach(self, observer: Observer) -> None:
        pass
    
    @abstractmethod
    def detach(self, observer: Observer) -> None:
        pass
    
    @abstractmethod
    def notify(self) -> None:
        pass


class NewsAgency(Subject):
    observers: list
    news: str
    
    def __init__(self) -> None:
        self.observers = []
        self.news = ""
    
    def attach(self, observer: Observer) -> None:
        self.observers.append(observer)
    
    def detach(self, observer: Observer) -> None:
        self.observers.remove(observer)
    
    def notify(self) -> None:
        for observer in self.observers:
            observer.update(self.news)
    
    def set_news(self, news: str) -> None:
        self.news = news
        self.notify()


class NewsChannel(Observer):
    def __init__(self, name: str) -> None:
        self.name = name
    
    def update(self, news: str) -> None:
        print(f"{self.name} received: {news}")


if __name__ == "__main__":
    agency = NewsAgency()
    
    channel1 = NewsChannel("Channel 1")
    channel2 = NewsChannel("Channel 2")
    
    agency.attach(channel1)
    agency.attach(channel2)
    
    agency.set_news("Breaking news!")
```

## When to Use

- When changes to one object require changing others, and you don't know how many objects need to change
- When an object should notify other objects without making assumptions about these objects
- When you need loose coupling between objects
