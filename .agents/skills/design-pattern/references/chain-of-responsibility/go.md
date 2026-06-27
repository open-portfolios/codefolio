# Chain of Responsibility

Pass a request along a Each handler decides to chain of handlers. process the request or pass it to the next handler.

## Intent

- Pass a request along a chain of handlers
- Each handler handles the request or forwards it to the next handler
- Decouple sender and receiver of a request

## Implementation

```go
package main

import "fmt"

type Handler interface {
	SetNext(handler Handler) Handler
	Handle(request string)
}

type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) SetNext(handler Handler) Handler {
	h.next = handler
	return handler
}

func (h *BaseHandler) Handle(request string) {
	if h.next != nil {
		h.next.Handle(request)
	}
}

type AuthHandler struct {
	BaseHandler
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Handle(request string) {
	if len(request) > 4 && request[:4] == "auth" {
		fmt.Println("Authentication handled")
		return
	}
	h.BaseHandler.Handle(request)
}

type ValidationHandler struct {
	BaseHandler
}

func NewValidationHandler() *ValidationHandler {
	return &ValidationHandler{}
}

func (h *ValidationHandler) Handle(request string) {
	if len(request) > 5 && request[:5] == "valid" {
		fmt.Println("Validation handled")
		return
	}
	h.BaseHandler.Handle(request)
}

func main() {
	authHandler := NewAuthHandler()
	validationHandler := NewValidationHandler()
	
	authHandler.SetNext(validationHandler)
	
	authHandler.Handle("auth:login")
	authHandler.Handle("valid:data")
	authHandler.Handle("other:request")
}
```

## When to Use

- When more than one object may handle a request
- When you want to issue a request to one of several objects without specifying the receiver explicitly
- When you want to set up a chain of handlers dynamically
