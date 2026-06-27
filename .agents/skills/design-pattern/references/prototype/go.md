# Prototype

Create new objects by cloning an existing object.

## Intent

- Create new objects by cloning an existing object
- Avoid costly object creation
- Produce objects that are independent of their creation process

## Implementation

```go
package main

import "fmt"

type Document interface {
	Clone() Document
}

type PDFDocument struct {
	Title   string
	Content string
}

func (d *PDFDocument) Clone() Document {
	return &PDFDocument{
		Title:   d.Title,
		Content: d.Content,
	}
}

type WordDocument struct {
	Title   string
	Content string
}

func (d *WordDocument) Clone() Document {
	return &WordDocument{
		Title:   d.Title,
		Content: d.Content,
	}
}

// Usage
func main() {
	original := &PDFDocument{
		Title:   "Original",
		Content: "Original content",
	}
	
	cloned := original.Clone()
	
	fmt.Println("Original:", original.Title)
	fmt.Println("Cloned:", cloned.Title)
	
	// Modify cloned
	cloned.(*PDFDocument).Title = "Cloned"
	fmt.Println("After modification:")
	fmt.Println("Original:", original.Title)
	fmt.Println("Cloned:", cloned.Title)
}
```

## When to Use

- When creating new objects is expensive (database calls, network requests)
- When you need to avoid subclassing to create objects
- When instances can have one of few possible states
