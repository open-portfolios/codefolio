# Builder

Separate the construction of a complex object from its representation.

## Intent

- Construct complex objects step by step
- Separate the construction of an object from its representation
- Create different representations using the same construction process

## Structure

```
Builder (interface)
    ├── buildPartA()
    ├── buildPartB()
    └── getResult() ──> Product

ConcreteBuilder
    └── implements Builder
```

## Implementation

```java
public class User {
    private final String name;
    private final String email;
    private final int age;
    private final String address;
    private final String phone;
    
    private User(Builder builder) {
        this.name = builder.name;
        this.email = builder.email;
        this.age = builder.age;
        this.address = builder.address;
        this.phone = builder.phone;
    }
    
    // Getters
    
    public static class Builder {
        private String name;
        private String email;
        private int age;
        private String address;
        private String phone;
        
        public Builder name(String name) {
            this.name = name;
            return this;
        }
        
        public Builder email(String email) {
            this.email = email;
            return this;
        }
        
        public Builder age(int age) {
            this.age = age;
            return this;
        }
        
        public Builder address(String address) {
            this.address = address;
            return this;
        }
        
        public Builder phone(String phone) {
            this.phone = phone;
            return this;
        }
        
        public User build() {
            return new User(this);
        }
    }
}

// Usage
public class Main {
    public static void main(String[] args) {
        User user = new User.Builder()
            .name("John")
            .email("john@example.com")
            .age(30)
            .address("123 Main St")
            .phone("555-1234")
            .build();
    }
}
```

## When to Use

- When object construction is complex with many optional parameters
- When you need different representations of the same construction process
- When you want to construct objects step by step
