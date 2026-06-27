# Strategy

Define a family of algorithms, encapsulate each one, and make them interchangeable. Strategy lets the algorithm vary independently from clients that use it.

## Intent

- Define a family of algorithms
- Encapsulate each algorithm
- Make algorithms interchangeable
- Let clients choose the algorithm to use

## Implementation

```go
package main

import "fmt"

type PaymentStrategy interface {
	Pay(amount float64)
}

type CreditCardPayment struct {
	CardNumber string
}

func NewCreditCardPayment(cardNumber string) *CreditCardPayment {
	return &CreditCardPayment{CardNumber: cardNumber}
}

func (p *CreditCardPayment) Pay(amount float64) {
	fmt.Printf("Paid %.2f with credit card %s\n", amount, p.CardNumber)
}

type PayPalPayment struct {
	Email string
}

func NewPayPalPayment(email string) *PayPalPayment {
	return &PayPalPayment{Email: email}
}

func (p *PayPalPayment) Pay(amount float64) {
	fmt.Printf("Paid %.2f via PayPal %s\n", amount, p.Email)
}

type Item struct {
	Name  string
	Price float64
}

type ShoppingCart struct {
	items           []Item
	paymentStrategy PaymentStrategy
}

func NewShoppingCart() *ShoppingCart {
	return &ShoppingCart{
		items: make([]Item, 0),
	}
}

func (c *ShoppingCart) AddItem(item Item) {
	c.items = append(c.items, item)
}

func (c *ShoppingCart) SetPaymentStrategy(strategy PaymentStrategy) {
	c.paymentStrategy = strategy
}

func (c *ShoppingCart) Checkout() {
	var total float64
	for _, item := range c.items {
		total += item.Price
	}
	c.paymentStrategy.Pay(total)
}

func main() {
	cart := NewShoppingCart()
	cart.AddItem(Item{Name: "Book", Price: 29.99})
	cart.AddItem(Item{Name: "Pen", Price: 5.99})
	
	cart.SetPaymentStrategy(NewCreditCardPayment("1234-5678-9012-3456"))
	cart.Checkout()
	
	cart.SetPaymentStrategy(NewPayPalPayment("user@example.com"))
	cart.Checkout()
}
```

## When to Use

- When you have multiple algorithms that accomplish the same task differently
- When you need to switch between algorithms at runtime
- When you want to avoid conditional statements for selecting behavior
