# Flyweight

Use sharing to support large numbers of fine-grained objects efficiently.

## Intent

- Use sharing to support large numbers of fine-grained objects efficiently
- Reduce memory consumption by sharing common state
- Separate intrinsic state (shared) from extrinsic state (unique)

## Implementation

```java
public class TreeType {
    private String name;
    private String color;
    private String texture;
    
    public TreeType(String name, String color, String texture) {
        this.name = name;
        this.color = color;
        this.texture = texture;
    }
    
    public void draw(int x, int y) {
        System.out.println("Drawing " + name + " tree at (" + x + ", " + y + ")");
    }
}

public class TreeFactory {
    private static Map<String, TreeType> treeTypes = new HashMap<>();
    
    public static TreeType getTreeType(String name, String color, String texture) {
        String key = name + "_" + color + "_" + texture;
        
        if (!treeTypes.containsKey(key)) {
            treeTypes.put(key, new TreeType(name, color, texture));
        }
        
        return treeTypes.get(key);
    }
}

public class Tree {
    private int x, y;
    private TreeType type;
    
    public Tree(int x, int y, TreeType type) {
        this.x = x;
        this.y = y;
        this.type = type;
    }
    
    public void draw() {
        type.draw(x, y);
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        TreeType oak = TreeFactory.getTreeType("Oak", "Green", "rough");
        TreeType pine = TreeFactory.getTreeType("Pine", "DarkGreen", "smooth");
        
        Tree tree1 = new Tree(10, 20, oak);
        Tree tree2 = new Tree(30, 40, oak);
        Tree tree3 = new Tree(50, 60, pine);
        
        tree1.draw();
        tree2.draw();
        tree3.draw();
    }
}
```

## When to Use

- When your application uses a large number of objects with significant memory cost
- When most object state can be made extrinsic
- When you need to maintain many objects in memory
