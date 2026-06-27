# Template Method

Define the skeleton of an algorithm in an operation, deferring some steps to subclasses. Template Method lets subclasses redefine certain steps without changing the algorithm's structure.

## Intent

- Define the skeleton of an algorithm
- Let subclasses override specific steps without changing the algorithm's structure
- Code reuse and common structure

## Implementation

```go
package main

import "fmt"

type DataMiner interface {
	OpenFile(path string) string
	ExtractData(file string) string
	ParseData(rawData string) []string
	Analyze(data []string)
	SendReport()
}

type BaseDataMiner struct{}

func NewBaseDataMiner() *BaseDataMiner {
	return &BaseDataMiner{}
}

func (m *BaseDataMiner) Mine(path string) {
	file := m.OpenFile(path)
	rawData := m.ExtractData(file)
	data := m.ParseData(rawData)
	m.Analyze(data)
	m.SendReport()
}

func (m *BaseDataMiner) SendReport() {
	fmt.Println("Report sent")
}

type PDFDataMiner struct {
	BaseDataMiner
}

func NewPDFDataMiner() *PDFDataMiner {
	return &PDFDataMiner{
		BaseDataMiner: *NewBaseDataMiner(),
	}
}

func (m *PDFDataMiner) OpenFile(path string) string {
	return "PDF content"
}

func (m *PDFDataMiner) ExtractData(file string) string {
	return "extracted data"
}

func (m *PDFDataMiner) ParseData(rawData string) []string {
	return []string{"data1", "data2"}
}

func (m *PDFDataMiner) Analyze(data []string) {
	fmt.Println("Analyzing", len(data), "records")
}

type CSVDataMiner struct {
	BaseDataMiner
}

func NewCSVDataMiner() *CSVDataMiner {
	return &CSVDataMiner{
		BaseDataMiner: *NewBaseDataMiner(),
	}
}

func (m *CSVDataMiner) OpenFile(path string) string {
	return "CSV content"
}

func (m *CSVDataMiner) ExtractData(file string) string {
	return "CSV extracted data"
}

func (m *CSVDataMiner) ParseData(rawData string) []string {
	return []string{"csv1", "csv2"}
}

func (m *CSVDataMiner) Analyze(data []string) {
	fmt.Println("Analyzing", len(data), "CSV records")
}

func main() {
	var miner DataMiner
	miner = NewPDFDataMiner()
	miner.Mine("file.pdf")
	
	miner = NewCSVDataMiner()
	miner.Mine("file.csv")
}
```

## When to Use

- When you have an algorithm with invariant steps
- When you want to let subclasses override specific steps
- When you want to control the extension points in a framework
