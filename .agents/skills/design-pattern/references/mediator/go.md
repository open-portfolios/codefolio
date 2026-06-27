# Mediator

Define an object that encapsulates how a set of objects interact. Mediator promotes loose coupling by keeping objects from referring to each other explicitly.

## Intent

- Encapsulate interaction between objects
- Decouple objects from knowing about each other
- Centralize control logic

## Implementation

```go
package main

import "fmt"

type User struct {
	name     string
	mediator ChatMediator
}

func NewUser(name string, mediator ChatMediator) User {
	return User{name: name, mediator: mediator}
}

func (u *User) Send(message string) {
	fmt.Println(u.name, "sends:", message)
	u.mediator.SendMessage(message, u)
}

func (u *User) ReceiveMessage(message string, senderName string) {
	fmt.Println(u.name, "received from", senderName+":", message)
}

type ChatMediator interface {
	SendMessage(message string, sender User)
	AddUser(user User)
}

type ChatRoom struct {
	users []User
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{}
}

func (c *ChatRoom) SendMessage(message string, sender User) {
	for _, user := range c.users {
		if user.name != sender.name {
			user.ReceiveMessage(message, sender.name)
		}
	}
}

func (c *ChatRoom) AddUser(user User) {
	c.users = append(c.users, user)
}

func main() {
	mediator := NewChatRoom()
	
	user1 := NewUser("Alice", mediator)
	user2 := NewUser("Bob", mediator)
	user3 := NewUser("Charlie", mediator)
	
	mediator.AddUser(user1)
	mediator.AddUser(user2)
	mediator.AddUser(user3)
	
	user1.Send("Hello everyone!")
}
```

## When to Use

- When you have a set of objects with complex communication
- When you want to reduce coupling between objects
- When you want to centralize complex interaction logic
