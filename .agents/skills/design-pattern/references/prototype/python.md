# Prototype

Create new objects by cloning an existing object.

## Intent

- Create new objects by cloning an existing object
- Avoid costly object creation
- Produce objects that are independent of their creation process

## Implementation

### Using copy Module

```python
import copy


class Document(object):
    title: str
    content: str
    
    def __init__(self, title: str, content: str) -> None:
        self.title = title
        self.content = content
    
    def clone(self) -> 'Document':
        return copy.deepcopy(self)
    
    def __repr__(self) -> str:
        return f"Document(title={self.title}, content={self.content})"


# Usage
if __name__ == "__main__":
    original = Document("Original", "Original content")
    cloned = original.clone()
    
    print("Original:", original)
    print("Cloned:", cloned)
    
    # Modify cloned
    cloned.title = "Cloned"
    print("\nAfter modification:")
    print("Original:", original)
    print("Cloned:", cloned)
```

### Using __copy__ and __deepcopy__

```python
import copy


class Document(object):
    title: str
    content: str
    
    def __init__(self, title: str, content: str) -> None:
        self.title = title
        self.content = content
    
    def __copy__(self) -> 'Document':
        return Document(self.title, self.content)
    
    def __deepcopy__(self, memo: dict) -> 'Document':
        return Document(
            copy.deepcopy(self.title, memo),
            copy.deepcopy(self.content, memo)
        )
    
    def clone(self) -> 'Document':
        return copy.deepcopy(self)


# Usage
if __name__ == "__main__":
    original = Document("Original", "Content")
    cloned = original.clone()
    print(original, cloned)
```

## When to Use

- When creating new objects is expensive (database calls, network requests)
- When you need to avoid subclassing to create objects
- When instances can have one of few possible states
