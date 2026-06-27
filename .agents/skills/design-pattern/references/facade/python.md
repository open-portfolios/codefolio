# Facade

Provide a unified interface to a set of interfaces in a subsystem. Facade defines a higher-level interface that makes the subsystem easier to use.

## Intent

- Provide a unified interface to a set of interfaces in a subsystem
- Make a subsystem easier to use
- Wrap a complex subsystem with a simpler interface

## Implementation

```python
class CPU(object):
    def freeze(self) -> None:
        print("CPU freeze")
    
    def jump(self, position: int) -> None:
        print(f"CPU jump to {position}")
    
    def execute(self) -> None:
        print("CPU execute")


class Memory(object):
    def load(self, position: int, data: bytes) -> None:
        print(f"Memory load at {position}")


class HardDrive(object):
    def read(self, lba: int, size: int) -> bytes:
        print(f"HardDrive read {size} bytes")
        return bytes(size)


# Facade
class ComputerFacade(object):
    cpu: CPU
    memory: Memory
    hard_drive: HardDrive
    
    def __init__(self) -> None:
        self.cpu = CPU()
        self.memory = Memory()
        self.hard_drive = HardDrive()
    
    def start(self) -> None:
        self.cpu.freeze()
        self.memory.load(0, self.hard_drive.read(0, 1024))
        self.cpu.jump(0)
        self.cpu.execute()


if __name__ == "__main__":
    computer = ComputerFacade()
    computer.start()
```

## When to Use

- When you want to provide a simple interface to a complex subsystem
- When there are many dependencies between clients and implementation classes
- When you want to layer your subsystems
