# Strategy

Define a family of algorithms, encapsulate each one, and make them interchangeable. Strategy lets the algorithm vary independently from clients that use it.

## Intent

- Define a family of algorithms
- Encapsulate each algorithm
- Make algorithms interchangeable
- Let clients choose the algorithm to use

## Implementation

```python
from abc import ABC, abstractmethod


class PaymentStrategy(ABC):
    @abstractmethod
    def pay(self, amount: float) -> None:
        pass


class CreditCardPayment(PaymentStrategy):
    card_number: str
    
    def __init__(self, card_number: str) -> None:
        self.card_number = card_number
    
    def pay(self, amount: float) -> None:
        print(f"Paid {amount} with credit card {self.card_number}")


class PayPalPayment(PaymentStrategy):
    email: str
    
    def __init__(self, email: str) -> None:
        self.email = email
    
    def pay(self, amount: float) -> None:
        print(f"Paid {amount} via PayPal {self.email}")


class Item(object):
    name: str
    price: float
    
    def __init__(self, name: str, price: float) -> None:
        self.name = name
        self.price = price


class ShoppingCart(object):
    items: list
    payment_strategy: PaymentStrategy | None
    
    def __init__(self) -> None:
        self.items = []
        self.payment_strategy = None
    
    def add_item(self, item: Item) -> None:
        self.items.append(item)
    
    def set_payment_strategy(self, strategy: PaymentStrategy) -> None:
        self.payment_strategy = strategy
    
    def checkout(self) -> None:
        total = sum(item.price for item in self.items)
        if self.payment_strategy is not None:
            self.payment_strategy.pay(total)


if __name__ == "__main__":
    cart = ShoppingCart()
    cart.add_item(Item("Book", 29.99))
    cart.add_item(Item("Pen", 5.99))
    
    cart.set_payment_strategy(CreditCardPayment("1234-5678-9012-3456"))
    cart.checkout()
    
    cart.set_payment_strategy(PayPalPayment("user@example.com"))
    cart.checkout()
```

## When to Use

- When you have multiple algorithms that accomplish the same task differently
- When you need to switch between algorithms at runtime
- When you want to avoid conditional statements for selecting behavior
