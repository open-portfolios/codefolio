# Memento

Without violating encapsulation, capture and externalize an object's internal state so that the object can be restored to this state later.

## Intent

- Capture and save an object's internal state
- Restore the object to a previous state
- Support undo operations

## Implementation

```java
public class Memento {
    private String state;
    
    public Memento(String state) {
        this.state = state;
    }
    
    public String getState() {
        return state;
    }
}

public class Originator {
    private String state;
    
    public void setState(String state) {
        this.state = state;
    }
    
    public String getState() {
        return state;
    }
    
    public Memento save() {
        return new Memento(state);
    }
    
    public void restore(Memento memento) {
        this.state = memento.getState();
    }
}

public class Caretaker {
    private List<Memento> mementos = new ArrayList<>();
    private Originator originator;
    
    public Caretaker(Originator originator) {
        this.originator = originator;
    }
    
    public void backup() {
        mementos.add(originator.save());
    }
    
    public void undo() {
        if (mementos.isEmpty()) {
            return;
        }
        originator.restore(mementos.remove(mementos.size() - 1));
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        Originator originator = new Originator();
        Caretaker caretaker = new Caretaker(originator);
        
        originator.setState("State 1");
        caretaker.backup();
        
        originator.setState("State 2");
        caretaker.backup();
        
        originator.setState("State 3");
        
        System.out.println("Current: " + originator.getState());
        
        caretaker.undo();
        System.out.println("After undo: " + originator.getState());
    }
}
```

## When to Use

- When you need to save and restore an object's state
- When you want to implement undo functionality
- When you need to capture snapshots of an object's state
