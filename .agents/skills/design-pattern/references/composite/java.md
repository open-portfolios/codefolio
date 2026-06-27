# Composite

Compose objects into tree structures to represent part-whole hierarchies. Let clients treat individual objects and compositions uniformly.

## Intent

- Compose objects into tree structures to represent part-whole hierarchies
- Let clients treat individual objects and compositions uniformly
- Simplify client code by treating complex structures as single objects

## Implementation

```java
public interface Component {
    void execute();
}

public class Leaf implements Component {
    private String name;
    
    public Leaf(String name) {
        this.name = name;
    }
    
    @Override
    public void execute() {
        System.out.println("Leaf " + name + " executed");
    }
}

public class Composite implements Component {
    private List<Component> children = new ArrayList<>();
    
    public void add(Component component) {
        children.add(component);
    }
    
    public void remove(Component component) {
        children.remove(component);
    }
    
    @Override
    public void execute() {
        for (Component child : children) {
            child.execute();
        }
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        Composite tree = new Composite();
        Composite branch1 = new Composite();
        
        branch1.add(new Leaf("A"));
        branch1.add(new Leaf("B"));
        
        tree.add(branch1);
        tree.add(new Leaf("C"));
        
        tree.execute();
    }
}
```

## When to Use

- When you want to represent of part-whole hierarchies objects
- When you want clients to ignore the difference between compositions and individual objects
- When you want to treat primitive and complex objects uniformly
