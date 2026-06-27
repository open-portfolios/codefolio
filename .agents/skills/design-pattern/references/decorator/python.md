# Decorator

Attach additional responsibilities to an object dynamically. Decorators provide a flexible alternative to subclassing for extending functionality.

## Intent

- Attach additional responsibilities to an object dynamically
- Provide a flexible alternative to subclassing for extending functionality
- Add responsibilities to objects without affecting other objects of the same class

## Implementation

```python
from abc import ABC, abstractmethod


class DataSource(ABC):
    @abstractmethod
    def write_data(self, data: str) -> None:
        pass
    
    @abstractmethod
    def read_data(self) -> str:
        pass


class FileDataSource(DataSource):
    filename: str
    data: str
    
    def __init__(self, filename: str) -> None:
        self.filename = filename
        self.data = ""
    
    def write_data(self, data: str) -> None:
        self.data = data
    
    def read_data(self) -> str:
        return self.data


class DataSourceDecorator(DataSource):
    source: DataSource
    
    def __init__(self, source: DataSource) -> None:
        self.source = source
    
    def write_data(self, data: str) -> None:
        self.source.write_data(data)
    
    def read_data(self) -> str:
        return self.source.read_data()


class EncryptionDecorator(DataSourceDecorator):
    def write_data(self, data: str) -> None:
        encrypted = f"encrypted({data})"
        self.source.write_data(encrypted)
    
    def read_data(self) -> str:
        data = self.source.read_data()
        return data[11:-1]


if __name__ == "__main__":
    source = FileDataSource("file.txt")
    source.write_data("sensitive data")
    
    encrypted = EncryptionDecorator(source)
    encrypted.write_data("sensitive data")
    
    print(encrypted.read_data())
```

## When to Use

- When you need to add responsibilities to objects dynamically
- When extension by subclassing is impractical or impossible
- When you need to add responsibilities that can be withdrawn later
