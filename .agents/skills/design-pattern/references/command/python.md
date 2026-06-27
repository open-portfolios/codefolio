# Command

Encapsulate a request as an object, letting you parameterize clients with different requests.

## Intent

- Encapsulate a request as an object
- Parameterize objects with different requests
- Support undoable operations
- Queue or log requests

## Implementation

```python
from abc import ABC, abstractmethod


class Command(ABC):
    @abstractmethod
    def execute(self) -> None:
        pass


class Light(object):
    def on(self) -> None:
        print("Light is on")
    
    def off(self) -> None:
        print("Light is off")


class LightOnCommand(Command):
    light: Light
    
    def __init__(self, light: Light) -> None:
        self.light = light
    
    def execute(self) -> None:
        self.light.on()


class LightOffCommand(Command):
    light: Light
    
    def __init__(self, light: Light) -> None:
        self.light = light
    
    def execute(self) -> None:
        self.light.off()


class RemoteControl(object):
    command: Command | None
    
    def set_command(self, command: Command) -> None:
        self.command = command
    
    def press_button(self) -> None:
        if self.command is not None:
            self.command.execute()


if __name__ == "__main__":
    light = Light()
    on_command = LightOnCommand(light)
    off_command = LightOffCommand(light)
    
    remote = RemoteControl()
    
    remote.set_command(on_command)
    remote.press_button()
    
    remote.set_command(off_command)
    remote.press_button()
```

## When to Use

- When you want to parameterize objects with actions
- When you want to queue, specify, and execute requests at different times
- When you want to support undo operations
