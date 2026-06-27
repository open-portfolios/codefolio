# Memento

Without violating encapsulation, capture and externalize an object's internal state so that the object can be restored to this state later.

## Intent

- Capture and save an object's internal state
- Restore the object to a previous state
- Support undo operations

## Implementation

```python
class Memento(object):
    def __init__(self, state: str) -> None:
        self.state = state


class Originator(object):
    state: str
    
    def __init__(self) -> None:
        self.state = ""
    
    def set_state(self, state: str) -> None:
        self.state = state
    
    def get_state(self) -> str:
        return self.state
    
    def save(self) -> Memento:
        return Memento(self.state)
    
    def restore(self, memento: Memento) -> None:
        self.state = memento.state


class Caretaker(object):
    mementos: list
    originator: Originator
    
    def __init__(self, originator: Originator) -> None:
        self.mementos = []
        self.originator = originator
    
    def backup(self) -> None:
        self.mementos.append(self.originator.save())
    
    def undo(self) -> None:
        if not self.mementos:
            return
        memento = self.mementos.pop()
        self.originator.restore(memento)


if __name__ == "__main__":
    originator = Originator()
    caretaker = Caretaker(originator)
    
    originator.set_state("State 1")
    caretaker.backup()
    
    originator.set_state("State 2")
    caretaker.backup()
    
    originator.set_state("State 3")
    
    print(f"Current: {originator.get_state()}")
    
    caretaker.undo()
    print(f"After undo: {originator.get_state()}")
```

## When to Use

- When you need to save and restore an object's state
- When you want to implement undo functionality
- When you need to capture snapshots of an object's state
