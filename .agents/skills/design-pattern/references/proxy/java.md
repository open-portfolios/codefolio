# Proxy

Provide a surrogate or placeholder for another object to control access to it.

## Intent

- Provide a surrogate or placeholder for another object to control access to it
- Implement lazy initialization (virtual proxy)
- Implement access control (protection proxy)
- Implement local execution of a remote service (remote proxy)
- Implement logging requests (logging proxy)

## Implementation

```java
public interface Image {
    void display();
}

public class RealImage implements Image {
    private String filename;
    
    public RealImage(String filename) {
        this.filename = filename;
        loadFromDisk();
    }
    
    private void loadFromDisk() {
        System.out.println("Loading " + filename);
    }
    
    @Override
    public void display() {
        System.out.println("Displaying " + filename);
    }
}

public class ImageProxy implements Image {
    private RealImage realImage;
    private String filename;
    
    public ImageProxy(String filename) {
        this.filename = filename;
    }
    
    @Override
    public void display() {
        if (realImage == null) {
            realImage = new RealImage(filename);
        }
        realImage.display();
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        Image image = new ImageProxy("photo.jpg");
        
        // Image is not loaded yet
        System.out.println("Image created");
        
        // Now image is loaded and displayed
        image.display();
        
        // Image is already loaded, no need to load again
        image.display();
    }
}
```

## When to Use

- When you need to control access to an object
- When you need lazy initialization
- When you need to add access control, logging, or other functionality
- When you need to interact with remote objects as if they were local
