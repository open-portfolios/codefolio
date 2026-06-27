# Decorator

Attach additional responsibilities to an object dynamically. Decorators provide a flexible alternative to subclassing for extending functionality.

## Intent

- Attach additional responsibilities to an object dynamically
- Provide a flexible alternative to subclassing for extending functionality
- Add responsibilities to objects without affecting other objects of the same type

## Implementation

```go
package main

import "fmt"

type DataSource interface {
	WriteData(data string)
	ReadData() string
}

type FileDataSource struct {
	filename string
	data     string
}

func NewFileDataSource(filename string) *FileDataSource {
	return &FileDataSource{filename: filename}
}

func (f *FileDataSource) WriteData(data string) {
	f.data = data
}

func (f *FileDataSource) ReadData() string {
	return f.data
}

type DataSourceDecorator struct {
	source DataSource
}

func NewDataSourceDecorator(source DataSource) *DataSourceDecorator {
	return &DataSourceDecorator{source: source}
}

func (d *DataSourceDecorator) WriteData(data string) {
	d.source.WriteData(data)
}

func (d *DataSourceDecorator) ReadData() string {
	return d.source.ReadData()
}

type EncryptionDecorator struct {
	DataSourceDecorator
}

func NewEncryptionDecorator(source DataSource) *EncryptionDecorator {
	return &EncryptionDecorator{
		DataSourceDecorator: DataSourceDecorator{source: source},
	}
}

func (e *EncryptionDecorator) WriteData(data string) {
	encrypted := "encrypted(" + data + ")"
	e.DataSourceDecorator.WriteData(encrypted)
}

func (e *EncryptionDecorator) ReadData() string {
	data := e.DataSourceDecorator.ReadData()
	return data[11 : len(data)-1]
}

// Client code
func main() {
	source := NewFileDataSource("file.txt")
	source.WriteData("sensitive data")
	
	encrypted := NewEncryptionDecorator(source)
	encrypted.WriteData("sensitive data")
	
	fmt.Println(encrypted.ReadData())
}
```

## When to Use

- When you need to add responsibilities to objects dynamically
- When extension by subclassing is impractical or impossible
- When you need to add responsibilities that can be withdrawn later
