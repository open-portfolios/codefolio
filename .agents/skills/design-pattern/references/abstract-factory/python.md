# Abstract Factory

Provide an interface for creating families of related objects without specifying concrete classes.

## Intent

- Provide an interface for creating families of related objects
- Ensure that created objects are compatible with each other
- Isolate concrete implementations from client code

## Implementation

```python
from abc import ABC, abstractmethod


# Abstract products
class Button(ABC):
    @abstractmethod
    def render(self) -> None:
        pass


class TextField(ABC):
    @abstractmethod
    def render(self) -> None:
        pass


# Abstract factory
class UIFactory(ABC):
    @abstractmethod
    def create_button(self) -> Button:
        pass
    
    @abstractmethod
    def create_text_field(self) -> TextField:
        pass


# Concrete products - Windows
class WindowsButton(Button):
    def render(self) -> None:
        print("Rendering Windows button")


class WindowsTextField(TextField):
    def render(self) -> None:
        print("Rendering Windows text field")


# Concrete products - Mac
class MacButton(Button):
    def render(self) -> None:
        print("Rendering Mac button")


class MacTextField(TextField):
    def render(self) -> None:
        print("Rendering Mac text field")


# Concrete factories
class WindowsFactory(UIFactory):
    def create_button(self) -> Button:
        return WindowsButton()
    
    def create_text_field(self) -> TextField:
        return WindowsTextField()


class MacFactory(UIFactory):
    def create_button(self) -> Button:
        return MacButton()
    
    def create_text_field(self) -> TextField:
        return MacTextField()


# Client code - depends on abstraction
class UIRenderer(object):
    def render_ui(self, factory: UIFactory) -> None:
        button = factory.create_button()
        text_field = factory.create_text_field()
        
        button.render()
        text_field.render()


# Usage
if __name__ == "__main__":
    renderer = UIRenderer()
    
    # On Mac
    mac_factory = MacFactory()
    renderer.render_ui(mac_factory)
    
    # On Windows
    windows_factory = WindowsFactory()
    renderer.render_ui(windows_factory)
```

## When to Use

- When your system needs to work with various families of related products
- When you want to provide a library without exposing implementation details
- When you need to ensure created objects are compatible with each other
