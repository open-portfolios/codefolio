# Composite

Compose objects into tree structures to represent part-whole hierarchies. Let clients treat individual objects and compositions uniformly.

## Intent

- Compose objects into tree structures to represent part-whole hierarchies
- Let clients treat individual objects and compositions uniformly
- Simplify client code by treating complex structures as single objects

## Implementation

```go
package main

import "fmt"

type Component interface {
	Execute()
}

type Leaf struct {
	Name string
}

func NewLeaf(name string) *Leaf {
	return &Leaf{Name: name}
}

func (l *Leaf) Execute() {
	fmt.Println("Leaf", l.Name, "executed")
}

type Composite struct {
	Children []Component
}

func NewComposite() *Composite {
	return &Composite{}
}

func (c *Composite) Add(component Component) {
	c.Children = append(c.Children, component)
}

func (c *Composite) Remove(component Component) {
	for i, child := range c.Children {
		if child == component {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			return
		}
	}
}

func (c *Composite) Execute() {
	for _, child := range c.Children {
		child.Execute()
	}
}

// Client code
func main() {
	tree := NewComposite()
	branch1 := NewComposite()
	
	branch1.Add(NewLeaf("A"))
	branch1.Add(NewLeaf("B"))
	
	tree.Add(branch1)
	tree.Add(NewLeaf("C"))
	
	tree.Execute()
}
```

## When to Use

- When you want to represent part-whole hierarchies objects
- When you want clients to ignore the difference between compositions and individual objects
- When you want to treat primitive and complex objects uniformly
