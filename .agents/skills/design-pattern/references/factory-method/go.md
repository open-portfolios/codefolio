# Factory Method

Define an interface for creating an object, but let subclasses decide which class to instantiate.

## Intent

- Define an interface for creating an object, but let subclasses decide which class to instantiate
- Delegate instantiation to subclasses
- Useful when a class cannot anticipate the type of objects it needs to create

## Implementation

```go
package main

import "fmt"

type Notification interface {
	Send(message string)
}

// Concrete products
type EmailNotification struct{}

func (e *EmailNotification) Send(message string) {
	fmt.Println("Sending email:", message)
}

type SMSNotification struct{}

func (s *SMSNotification) Send(message string) {
	fmt.Println("Sending SMS:", message)
}

// Creator interface
type NotificationFactory interface {
	CreateNotification() Notification
}

// Concrete creators
type EmailFactory struct{}

func NewEmailFactory() *EmailFactory {
	return &EmailFactory{}
}

func (e *EmailFactory) CreateNotification() Notification {
	return &EmailNotification{}
}

type SMSFactory struct{}

func NewSMSFactory() *SMSFactory {
	return &SMSFactory{}
}

func (s *SMSFactory) CreateNotification() Notification {
	return &SMSNotification{}
}

// Client code - depends on abstraction
func sendNotification(factory NotificationFactory, message string) {
	notification := factory.CreateNotification()
	notification.Send(message)
}

// Usage
func main() {
	// Use email factory
	emailFactory := NewEmailFactory()
	sendNotification(emailFactory, "Hello via Email")
	
	// Use SMS factory
	smsFactory := NewSMSFactory()
	sendNotification(smsFactory, "Hello via SMS")
}
```

## When to Use

- When a function cannot anticipate the type of objects it needs to create
- When you want to delegate object creation to subclasses
- When you need to choose between different implementations at runtime
