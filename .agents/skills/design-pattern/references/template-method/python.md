# Template Method

Define the skeleton of an algorithm in an operation, deferring some steps to subclasses. Template Method lets subclasses redefine certain steps without changing the algorithm's structure.

## Intent

- Define the skeleton of an algorithm
- Let subclasses override specific steps without changing the algorithm's structure
- Code reuse and common structure

## Implementation

```python
from abc import ABC, abstractmethod


class DataMiner(ABC):
    def mine(self, path: str) -> None:
        file = self.open_file(path)
        raw_data = self.extract_data(file)
        data = self.parse_data(raw_data)
        self.analyze(data)
        self.send_report()
    
    @abstractmethod
    def open_file(self, path: str) -> str:
        pass
    
    @abstractmethod
    def extract_data(self, file: str) -> str:
        pass
    
    @abstractmethod
    def parse_data(self, raw_data: str) -> list:
        pass
    
    @abstractmethod
    def analyze(self, data: list) -> None:
        pass
    
    def send_report(self) -> None:
        print("Report sent")


class PDFDataMiner(DataMiner):
    def open_file(self, path: str) -> str:
        return "PDF content"
    
    def extract_data(self, file: str) -> str:
        return "extracted data"
    
    def parse_data(self, raw_data: str) -> list:
        return raw_data.split(",")
    
    def analyze(self, data: list) -> None:
        print(f"Analyzing {len(data)} records")


class CSVDataMiner(DataMiner):
    def open_file(self, path: str) -> str:
        return "CSV content"
    
    def extract_data(self, file: str) -> str:
        return "CSV extracted data"
    
    def parse_data(self, raw_data: str) -> list:
        return raw_data.split(",")
    
    def analyze(self, data: list) -> None:
        print(f"Analyzing {len(data)} CSV records")


if __name__ == "__main__":
    miner = PDFDataMiner()
    miner.mine("file.pdf")
    
    miner = CSVDataMiner()
    miner.mine("file.csv")
```

## When to Use

- When you have an algorithm with invariant steps
- When you want to let subclasses override specific steps
- When you want to control the extension points in a framework
