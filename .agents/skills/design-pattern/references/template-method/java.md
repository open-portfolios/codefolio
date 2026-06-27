# Template Method

Define the skeleton of an algorithm in an operation, deferring some steps to subclasses. Template Method lets subclasses redefine certain steps without changing the algorithm's structure.

## Intent

- Define the skeleton of an algorithm
- Let subclasses override specific steps without changing the algorithm's structure
- Code reuse and common structure

## Implementation

```java
public abstract class DataMiner {
    public final void mine(String path) {
        String file = openFile(path);
        String rawData = extractData(file);
        String[] data = parseData(rawData);
        analyze(data);
        sendReport();
    }
    
    protected abstract String openFile(String path);
    protected abstract String extractData(String file);
    protected abstract String[] parseData(String rawData);
    protected abstract void analyze(String[] data);
    
    protected void sendReport() {
        System.out.println("Report sent");
    }
}

public class PDFDataMiner extends DataMiner {
    @Override
    protected String openFile(String path) {
        return "PDF content";
    }
    
    @Override
    protected String extractData(String file) {
        return "extracted data";
    }
    
    @Override
    protected String[] parseData(String rawData) {
        return rawData.split(",");
    }
    
    @Override
    protected void analyze(String[] data) {
        System.out.println("Analyzing " + data.length + " records");
    }
}

public class CSVDataMiner extends DataMiner {
    @Override
    protected String openFile(String path) {
        return "CSV content";
    }
    
    @Override
    protected String extractData(String file) {
        return "CSV extracted data";
    }
    
    @Override
    protected String[] parseData(String rawData) {
        return rawData.split(",");
    }
    
    @Override
    protected void analyze(String[] data) {
        System.out.println("Analyzing " + data.length + " CSV records");
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        DataMiner miner = new PDFDataMiner();
        miner.mine("file.pdf");
        
        miner = new CSVDataMiner();
        miner.mine("file.csv");
    }
}
```

## When to Use

- When you have an algorithm with invariant steps
- When you want to let subclasses override specific steps
- When you want to control the extension points in a framework
