# State

Allow an object to alter its behavior when its internal state changes. The object will appear to change its class.

## Intent

- Allow an object to change its behavior when its internal state changes
- Let the object appear to change its class
- Encapsulate state-specific behavior

## Implementation

```python
from abc import ABC, abstractmethod


class State(ABC):
    @abstractmethod
    def insert_coin(self, machine: 'VendingMachine') -> None:
        pass
    
    @abstractmethod
    def eject_coin(self, machine: 'VendingMachine') -> None:
        pass
    
    @abstractmethod
    def dispense(self, machine: 'VendingMachine') -> None:
        pass


class NoCoinState(State):
    def insert_coin(self, machine: 'VendingMachine') -> None:
        print("Coin inserted")
        machine.set_state(machine.has_coin_state)
    
    def eject_coin(self, machine: 'VendingMachine') -> None:
        print("No coin to eject")
    
    def dispense(self, machine: 'VendingMachine') -> None:
        print("Insert coin first")


class HasCoinState(State):
    def insert_coin(self, machine: 'VendingMachine') -> None:
        print("Coin already inserted")
    
    def eject_coin(self, machine: 'VendingMachine') -> None:
        print("Coin returned")
        machine.set_state(machine.no_coin_state)
    
    def dispense(self, machine: 'VendingMachine') -> None:
        print("Item dispensed")
        machine.set_state(machine.no_coin_state)


class VendingMachine(object):
    no_coin_state: NoCoinState
    has_coin_state: HasCoinState
    current_state: State
    
    def __init__(self) -> None:
        self.no_coin_state = NoCoinState()
        self.has_coin_state = HasCoinState()
        self.current_state = self.no_coin_state
    
    def set_state(self, state: State) -> None:
        self.current_state = state
    
    def insert_coin(self) -> None:
        self.current_state.insert_coin(self)
    
    def eject_coin(self) -> None:
        self.current_state.eject_coin(self)
    
    def dispense(self) -> None:
        self.current_state.dispense(self)


if __name__ == "__main__":
    machine = VendingMachine()
    
    machine.dispense()
    machine.insert_coin()
    machine.dispense()
```

## When to Use

- When an object's behavior depends on its state, and it must change its behavior at runtime
- When operations have large conditional statements that depend on the object's state
- When you want to avoid large if-else or switch statements
