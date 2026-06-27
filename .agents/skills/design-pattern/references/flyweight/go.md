# Flyweight

Use sharing to support large numbers of fine-grained objects efficiently.

## Intent

- Use sharing to support large numbers of fine-grained objects efficiently
- Reduce memory consumption by sharing common state
- Separate intrinsic state (shared) from extrinsic state (unique)

## Implementation

```go
package main

import "fmt"

type TreeType struct {
	Name   string
	Color  string
	Texture string
}

type TreeFactory struct {
	treeTypes map[string]*TreeType
}

var treeFactoryInstance *TreeFactory

func GetTreeFactory() *TreeFactory {
	if treeFactoryInstance == nil {
		treeFactoryInstance = &TreeFactory{
			treeTypes: make(map[string]*TreeType),
		}
	}
	return treeFactoryInstance
}

func (f *TreeFactory) GetTreeType(name, color, texture string) *TreeType {
	key := name + "_" + color + "_" + texture
	if treeType, ok := f.treeTypes[key]; ok {
		return treeType
	}
	treeType := &TreeType{Name: name, Color: color, Texture: texture}
	f.treeTypes[key] = treeType
	return treeType
}

type Tree struct {
	x, y int
	type_ *TreeType
}

func NewTree(x, y int, treeType *TreeType) *Tree {
	return &Tree{x: x, y: y, type_: treeType}
}

func (t *Tree) Draw() {
	fmt.Println("Drawing", t.type_.Name, "tree at", t.x, t.y)
}

// Client code
func main() {
	factory := GetTreeFactory()
	
	oak := factory.GetTreeType("Oak", "Green", "rough")
	pine := factory.GetTreeType("Pine", "DarkGreen", "smooth")
	
	tree1 := NewTree(10, 20, oak)
	tree2 := NewTree(30, 40, oak)
	tree3 := NewTree(50, 60, pine)
	
	tree1.Draw()
	tree2.Draw()
	tree3.Draw()
}
```

## When to Use

- When your application uses a large number of objects with significant memory cost
- When most object state can be made extrinsic
- When you need to maintain many objects in memory
