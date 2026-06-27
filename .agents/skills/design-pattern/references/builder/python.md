# Builder

Separate the construction of a complex object from its representation.

## Intent

- Construct complex objects step by step
- Separate the construction of an object from its representation
- Create different representations using the same construction process

## Implementation

```python
class User(object):
    name: str
    email: str
    age: int
    address: str | None
    phone: str | None
    
    def __init__(self, name: str, email: str, age: int, address: str | None = None, phone: str | None = None) -> None:
        self.name = name
        self.email = email
        self.age = age
        self.address = address
        self.phone = phone
    
    def __repr__(self) -> str:
        return f"User(name={self.name}, email={self.email}, age={self.age})"


class UserBuilder(object):
    name: str | None
    email: str | None
    age: int | None
    address: str | None
    phone: str | None
    
    def __init__(self) -> None:
        pass
    
    def name(self, name: str) -> 'UserBuilder':
        self.name = name
        return self
    
    def email(self, email: str) -> 'UserBuilder':
        self.email = email
        return self
    
    def age(self, age: int) -> 'UserBuilder':
        self.age = age
        return self
    
    def address(self, address: str) -> 'UserBuilder':
        self.address = address
        return self
    
    def phone(self, phone: str) -> 'UserBuilder':
        self.phone = phone
        return self
    
    def build(self) -> User:
        if self.name is None:
            raise ValueError("name is required")
        if self.email is None:
            raise ValueError("email is required")
        return User(
            name=self.name,
            email=self.email,
            age=self.age,
            address=self.address,
            phone=self.phone
        )


# Usage
if __name__ == "__main__":
    user = (UserBuilder()
            .name("John")
            .email("john@example.com")
            .age(30)
            .address("123 Main St")
            .phone("555-1234")
            .build())
    
    print(user)
```

## When to Use

- When object construction is complex with many optional parameters
- When you need different representations of the same construction process
- When you want to construct objects step by step
