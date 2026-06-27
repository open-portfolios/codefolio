# Iterator

Provide a way to access elements of a collection sequentially without exposing its underlying representation.

## Intent

- Access elements of a collection sequentially
- Decouple iteration from the collection
- Support multiple traversals simultaneously

## Implementation

```python
from abc import ABC, abstractmethod


class Iterator(ABC):
    @abstractmethod
    def has_next(self) -> bool:
        pass
    
    @abstractmethod
    def next(self) -> object:
        pass


class ArrayIterator(Iterator):
    array: list
    position: int
    
    def __init__(self, array: list) -> None:
        self.array = array
        self.position = 0
    
    def has_next(self) -> bool:
        return self.position < len(self.array)
    
    def next(self) -> object:
        item = self.array[self.position]
        self.position += 1
        return item


class ListIterator(Iterator):
    list_data: list
    position: int
    
    def __init__(self, list_data: list) -> None:
        self.list_data = list_data
        self.position = 0
    
    def has_next(self) -> bool:
        return self.position < len(self.list_data)
    
    def next(self) -> object:
        item = self.list_data[self.position]
        self.position += 1
        return item


if __name__ == "__main__":
    array = ["A", "B", "C"]
    array_iterator = ArrayIterator(array)
    
    while array_iterator.has_next():
        print(array_iterator.next())
    
    list_data = ["X", "Y", "Z"]
    list_iterator = ListIterator(list_data)
    
    while list_iterator.has_next():
        print(list_iterator.next())
```

## When to Use

- When you want to access a collection's elements without exposing its internal structure
- When you want to provide multiple traversal methods for the same collection
- When you want to decouple collection logic from iteration logic
