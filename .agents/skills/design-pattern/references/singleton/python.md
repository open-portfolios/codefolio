# Singleton

Ensure a class has only one instance and provide a global point of access to it.

## Intent

- Ensure a class has only one instance
- Provide a global point of access to that instance
- Control object creation at a single point

## Implementation

### Using `__new__`

```python
class Singleton(object):
    _instance: 'Singleton' | None
    
    def __new__(cls, *args, **kwargs) -> 'Singleton':
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance


# Usage
if __name__ == "__main__":
    s1 = Singleton()
    s2 = Singleton()
    
    print(s1 is s2)  # True
```

### Using Module-Level Variable

```python
# singleton.py

class _Config(object):
    host: str
    port: int
    
    def __init__(self, host: str = "localhost", port: int = 8080) -> None:
        self.host = host
        self.port = port


config = _Config()


# Usage
if __name__ == "__main__":
    c1 = config
    c2 = config
    
    print(c1 is c2)  # True
```

### Using Decorator

```python
def singleton(cls):
    instances: dict[type, object]
    
    def get_instance(*args, **kwargs):
        if cls not in instances:
            instances[cls] = cls(*args, **kwargs)
        return instances[cls]
    
    return get_instance


@singleton
class Config(object):
    host: str
    port: int
    
    def __init__(self, host: str = "localhost", port: int = 8080) -> None:
        self.host = host
        self.port = port


# Usage
if __name__ == "__main__":
    c1 = Config()
    c2 = Config()
    
    print(c1 is c2)  # True
```

## When to Use

- When exactly one instance of a class is needed
- When you need global access to that instance
- Common use cases: configuration managers, logging, connection pools

## Considerations

- Consider dependency injection instead of Singleton when possible
- Singletons can make testing difficult
- Module-level variable is the simplest and most Pythonic approach
