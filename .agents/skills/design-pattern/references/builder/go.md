# Builder

Separate the construction of a complex object from its representation.

## Intent

- Construct complex objects step by step
- Separate the construction of an object from its representation
- Create different representations using the same construction process

## Implementation

```go
package main

import "fmt"

type User struct {
	Name    string
	Email   string
	Age     int
	Address string
	Phone   string
}

type UserBuilder struct {
	user User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

func (b *UserBuilder) Name(name string) *UserBuilder {
	b.user.Name = name
	return b
}

func (b *UserBuilder) Email(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) Age(age int) *UserBuilder {
	b.user.Age = age
	return b
}

func (b *UserBuilder) Address(address string) *UserBuilder {
	b.user.Address = address
	return b
}

func (b *UserBuilder) Phone(phone string) *UserBuilder {
	b.user.Phone = phone
	return b
}

func (b *UserBuilder) Build() *User {
	return &b.user
}

// Usage
func main() {
	user := NewUserBuilder().
		Name("John").
		Email("john@example.com").
		Age(30).
		Address("123 Main St").
		Phone("555-1234").
		Build()
	
	fmt.Println(user.Name, user.Email)
}
```

## When to Use

- When object construction is complex with many optional parameters
- When you need different representations of the same construction process
- When you want to construct objects step by step
