# Proxy

Provide a surrogate or placeholder for another object to control access to it.

## Intent

- Provide a surrogate or placeholder for another object to control access to it
- Implement lazy initialization (virtual proxy)
- Implement access control (protection proxy)
- Implement local execution of a remote service (remote proxy)
- Implement logging requests (logging proxy)

## Implementation

```python
from abc import ABC, abstractmethod


class Image(ABC):
    @abstractmethod
    def display(self) -> None:
        pass


class RealImage(Image):
    filename: str
    
    def __init__(self, filename: str) -> None:
        self.filename = filename
        self._load_from_disk()
    
    def _load_from_disk(self) -> None:
        print(f"Loading {self.filename}")
    
    def display(self) -> None:
        print(f"Displaying {self.filename}")


class ImageProxy(Image):
    filename: str
    real_image: RealImage | None
    
    def __init__(self, filename: str) -> None:
        self.filename = filename
        self.real_image = None
    
    def display(self) -> None:
        if self.real_image is None:
            self.real_image = RealImage(self.filename)
        self.real_image.display()


if __name__ == "__main__":
    image = ImageProxy("photo.jpg")
    
    # Image is not loaded yet
    print("Image created")
    
    # Now image is loaded and displayed
    image.display()
    
    # Image is already loaded, no need to load again
    image.display()
```

## When to Use

- When you need to control access to an object
- When you need lazy initialization
- When you need to add access control, logging, or other functionality
- When you need to interact with remote objects as if they were local
