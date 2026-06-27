# Flyweight

Use sharing to support large numbers of fine-grained objects efficiently.

## Intent

- Use sharing to support large numbers of fine-grained objects efficiently
- Reduce memory consumption by sharing common state
- Separate intrinsic state (shared) from extrinsic state (unique)

## Implementation

```python
class TreeType(object):
    name: str
    color: str
    texture: str
    
    def __init__(self, name: str, color: str, texture: str) -> None:
        self.name = name
        self.color = color
        self.texture = texture
    
    def draw(self, x: int, y: int) -> None:
        print(f"Drawing {self.name} tree at ({x}, {y})")


class TreeFactory(object):
    _instance: 'TreeFactory' | None
    _tree_types: dict[str, TreeType]
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance
    
    def get_tree_type(self, name: str, color: str, texture: str) -> TreeType:
        key = f"{name}_{color}_{texture}"
        if key not in self._tree_types:
            self._tree_types[key] = TreeType(name, color, texture)
        return self._tree_types[key]


class Tree(object):
    x: int
    y: int
    type: TreeType
    
    def __init__(self, x: int, y: int, tree_type: TreeType) -> None:
        self.x = x
        self.y = y
        self.type = tree_type
    
    def draw(self) -> None:
        self.type.draw(self.x, self.y)


if __name__ == "__main__":
    factory = TreeFactory()
    
    oak = factory.get_tree_type("Oak", "Green", "rough")
    pine = factory.get_tree_type("Pine", "DarkGreen", "smooth")
    
    tree1 = Tree(10, 20, oak)
    tree2 = Tree(30, 40, oak)
    tree3 = Tree(50, 60, pine)
    
    tree1.draw()
    tree2.draw()
    tree3.draw()
```

## When to Use

- When your application uses a large number of objects with significant memory cost
- When most object state can be made extrinsic
- When you need to maintain many objects in memory
