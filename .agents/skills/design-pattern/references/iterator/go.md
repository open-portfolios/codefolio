# Iterator

Provide a way to access elements of a collection sequentially without exposing its underlying representation.

## Intent

- Access elements of a collection sequentially
- Decouple iteration from the collection
- Support multiple traversals simultaneously

## Implementation

```go
package main

import "fmt"

type Iterator interface {
	HasNext() bool
	Next() any
}

type ArrayIterator struct {
	array   []any
	position int
}

func NewArrayIterator(array []any) *ArrayIterator {
	return &ArrayIterator{array: array}
}

func (i *ArrayIterator) HasNext() bool {
	return i.position < len(i.array)
}

func (i *ArrayIterator) Next() any {
	item := i.array[i.position]
	i.position++
	return item
}

type ListIterator struct {
	list     []any
	position int
}

func NewListIterator(list []any) *ListIterator {
	return &ListIterator{list: list}
}

func (i *ListIterator) HasNext() bool {
	return i.position < len(i.list)
}

func (i *ListIterator) Next() any {
	item := i.list[i.position]
	i.position++
	return item
}

func main() {
	array := []any{"A", "B", "C"}
	arrayIterator := NewArrayIterator(array)
	
	for arrayIterator.HasNext() {
		fmt.Println(arrayIterator.Next())
	}
	
	list := []any{"X", "Y", "Z"}
	listIterator := NewListIterator(list)
	
	for listIterator.HasNext() {
		fmt.Println(listIterator.Next())
	}
}
```

## When to Use

- When you want to access a collection's elements without exposing its internal structure
- When you want to provide multiple traversal methods for the same collection
- When you want to decouple collection logic from iteration logic
