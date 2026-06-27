# Decorator

Attach additional responsibilities to an object dynamically. Decorators provide a flexible alternative to subclassing for extending functionality.

## Intent

- Attach additional responsibilities to an object dynamically
- Provide a flexible alternative to subclassing for extending functionality
- Add responsibilities to objects without affecting other objects of the same class

## Implementation

```java
public interface DataSource {
    void writeData(String data);
    String readData();
}

public class FileDataSource implements DataSource {
    private String filename;
    private String data;
    
    public FileDataSource(String filename) {
        this.filename = filename;
    }
    
    @Override
    public void writeData(String data) {
        this.data = data;
    }
    
    @Override
    public String readData() {
        return data;
    }
}

public class DataSourceDecorator implements DataSource {
    protected DataSource source;
    
    public DataSourceDecorator(DataSource source) {
        this.source = source;
    }
    
    @Override
    public void writeData(String data) {
        source.writeData(data);
    }
    
    @Override
    public String readData() {
        return source.readData();
    }
}

public class EncryptionDecorator extends DataSourceDecorator {
    public EncryptionDecorator(DataSource source) {
        super(source);
    }
    
    @Override
    public void writeData(String data) {
        super.writeData(encrypt(data));
    }
    
    @Override
    public String readData() {
        return decrypt(super.readData());
    }
    
    private String encrypt(String data) {
        return "encrypted(" + data + ")";
    }
    
    private String decrypt(String data) {
        return data.replace("encrypted(", "").replace(")", "");
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        DataSource source = new FileDataSource("file.txt");
        source.writeData("sensitive data");
        
        DataSource encrypted = new EncryptionDecorator(source);
        encrypted.writeData("sensitive data");
        
        System.out.println(encrypted.readData());
    }
}
```

## When to Use

- When you need to add responsibilities to objects dynamically
- When extension by subclassing is impractical or impossible
- When you need to add responsibilities that can be withdrawn later
