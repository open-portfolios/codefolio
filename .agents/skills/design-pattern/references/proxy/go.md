# Proxy

Provide a surrogate or placeholder for another object to control access to it.

## Intent

- Provide a surrogate or placeholder for another object to control access to it
- Implement lazy initialization (virtual proxy)
- Implement access control (protection proxy)
- Implement local execution of a remote service (remote proxy)
- Implement logging requests (logging proxy)

## Implementation

```go
package main

import "fmt"

type Image interface {
	Display()
}

type RealImage struct {
	filename string
}

func NewRealImage(filename string) *RealImage {
	fmt.Println("Loading", filename)
	return &RealImage{filename: filename}
}

func (i *RealImage) Display() {
	fmt.Println("Displaying", i.filename)
}

type ImageProxy struct {
	realImage *RealImage
	filename  string
}

func NewImageProxy(filename string) *ImageProxy {
	return &ImageProxy{filename: filename}
}

func (p *ImageProxy) Display() {
	if p.realImage == nil {
		p.realImage = NewRealImage(p.filename)
	}
	p.realImage.Display()
}

// Client code
func main() {
	image := NewImageProxy("photo.jpg")
	
	// Image is not loaded yet
	fmt.Println("Image created")
	
	// Now image is loaded and displayed
	image.Display()
	
	// Image is already loaded, no need to load again
	image.Display()
}
```

## When to Use

- When you need to control access to an object
- When you need lazy initialization
- When you need to add access control, logging, or other functionality
- When you need to interact with remote objects as if they were local
