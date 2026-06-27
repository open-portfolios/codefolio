# Composite

Compose objects into tree structures to represent part-whole hierarchies. Let clients treat individual objects and compositions uniformly.

## Intent

- Compose objects into tree structures to represent part-whole hierarchies
- Let clients treat individual objects and compositions uniformly
- Simplify client code by treating complex structures as single objects

## Implementation

```python
from abc import ABC, abstractmethod


class Component(ABC):
    @abstractmethod
    def execute(self) -> None:
        pass


class Leaf(Component):
    name: str
    
    def __init__(self, name: str) -> None:
        self.name = name
    
    def execute(self) -> None:
        print(f"Leaf {self.name} executed")


class Composite(Component):
    children: list[Component]
    
    def __init__(self) -> None:
        self.children = []
    
    def add(self, component: Component) -> None:
        self.children.append(component)
    
    def remove(self, component: Component) -> None:
        self.children.remove(component)
    
    def execute(self) -> None:
        for child in self.children:
            child.execute()


if __name__ == "__main__":
    tree = Composite()
    branch1 = Composite()
    
    branch1.add(Leaf("A"))
    branch1.add(Leaf("B"))
    
    tree.add(branch1)
    tree.add(Leaf("C"))
    
    tree.execute()
```

## When to Use

- When you want to represent part-whole hierarchies objects
- When you want clients to ignore the difference between compositions and individual objects
- When you want to treat primitive and complex objects uniformly
