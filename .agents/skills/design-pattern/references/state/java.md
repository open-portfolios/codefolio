# State

Allow an object to alter its behavior when its internal state changes. The object will appear to change its class.

## Intent

- Allow an object to change its behavior when its internal state changes
- Let the object appear to change its class
- Encapsulate state-specific behavior

## Implementation

```java
public interface State {
    void insertCoin(VendingMachine machine);
    void ejectCoin(VendingMachine machine);
    void dispense(VendingMachine machine);
}

public class NoCoinState implements State {
    @Override
    public void insertCoin(VendingMachine machine) {
        System.out.println("Coin inserted");
        machine.setState(machine.getHasCoinState());
    }
    
    @Override
    public void ejectCoin(VendingMachine machine) {
        System.out.println("No coin to eject");
    }
    
    @Override
    public void dispense(VendingMachine machine) {
        System.out.println("Insert coin first");
    }
}

public class HasCoinState implements State {
    @Override
    public void insertCoin(VendingMachine machine) {
        System.out.println("Coin already inserted");
    }
    
    @Override
    public void ejectCoin(VendingMachine machine) {
        System.out.println("Coin returned");
        machine.setState(machine.getNoCoinState());
    }
    
    @Override
    public void dispense(VendingMachine machine) {
        System.out.println("Item dispensed");
        machine.setState(machine.getNoCoinState());
    }
}

public class VendingMachine {
    private State noCoinState;
    private State hasCoinState;
    private State currentState;
    
    public VendingMachine() {
        noCoinState = new NoCoinState();
        hasCoinState = new HasCoinState();
        currentState = noCoinState;
    }
    
    public void insertCoin() {
        currentState.insertCoin(this);
    }
    
    public void ejectCoin() {
        currentState.ejectCoin(this);
    }
    
    public void dispense() {
        currentState.dispense(this);
    }
    
    public void setState(State state) {
        this.currentState = state;
    }
    
    public State getNoCoinState() {
        return noCoinState;
    }
    
    public State getHasCoinState() {
        return hasCoinState;
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        VendingMachine machine = new VendingMachine();
        
        machine.dispense();
        machine.insertCoin();
        machine.dispense();
    }
}
```

## When to Use

- When an object's behavior depends on its state, and it must change its behavior at runtime
- When operations have large conditional statements that depend on the object's state
- When you want to avoid large if-else or switch statements
