# Domain-Driven Design

Domain-Driven Design (DDD) is an approach to software development that emphasizes modeling software around business domains. It is particularly valuable for complex business applications where the domain logic is sophisticated and frequently changing.

## Core Concepts

### Bounded Context

A bounded context is a linguistic boundary around a particular domain model. Within this boundary, all terms and concepts have a specific meaning. Each bounded context should be independent and have its own ubiquitous language.

- Define clear boundaries between different business domains
- Each context has its own model and language
- Communication between contexts happens through explicit interfaces

### Entities

An entity is an object with a distinct identity that runs through time and different representations. Unlike value objects, entities are defined by their identity, not their attributes.

- Objects with a unique identifier that distinguishes them from other objects
- The identity persists through state changes
- Two entities with the same attributes but different IDs are considered different

### Value Objects

A value object is an object that describes some characteristic or attribute but has no conceptual identity. Two value objects with the same attributes are considered equal and interchangeable.

- Immutable objects defined by their attributes, not by a unique identifier
- No lifecycle; created and discarded as needed
- Side-effect-free operations

### Aggregates

An aggregate is a cluster of associated objects that are treated as a single unit for data changes. The aggregate root is the only member of the aggregate that outside objects are allowed to hold references to.

- Define clear consistency boundaries
- The aggregate root controls all changes within the aggregate
- External code only interacts with the aggregate root

### Domain Services

A domain service is a way to organize domain logic that doesn't naturally fit within an entity or value object. It encapsulates domain concepts that are not naturally modeled as part of entities or value objects.

- Coordinate multiple entities or value objects
- Contain domain logic that doesn't belong to a single entity
- Keep domain services thin; most logic should be in entities and value objects

## Project Structure

DDD project structure varies by language. Follow language conventions.

### Go

Go uses lowercase package names and short, lowercase file names without underscores.

```
order/
├── order.go              # Entity
├── aggregate.go         # Aggregate root
├── money.go              # Value object
├── address.go            # Value object
└── pricing.go            # Domain service

user/
├── user.go
└── address.go

application/
├── createorder.go        # Use case
└── dto/
    └── order.go

infrastructure/
├── persistence/
    └── order.go          # Repository interface
└── payment/
    └── gateway.go
```

### Java

Java uses PascalCase for both package and file names.

```
src/main/java/
└── com/example/
    └── order/
        ├── domain/
        │   ├── entity/
        │   │   └── Order.java
        │   ├── valueobject/
        │   │   ├── Money.java
        │   │   └── Address.java
        │   ├── aggregate/
        │   │   └── OrderAggregate.java
        │   └── service/
        │       └── PricingService.java
        ├── application/
        │   ├── usecase/
        │   │   └── CreateOrderUseCase.java
        │   └── dto/
        │       └── OrderDTO.java
        └── infrastructure/
            ├── persistence/
            │   └── OrderRepository.java
            └── payment/
                └── PaymentGateway.java
```

### Python

Python uses snake_case for file names.

```
src/
├── domain/
│   ├── order/
│   │   ├── entity.py
│   │   ├── aggregate.py
│   │   ├── value_objects.py
│   │   └── services.py
│   └── user/
│       ├── entity.py
│       └── value_objects.py
├── application/
│   ├── order/
│   │   ├── use_cases.py
│   │   └── dto.py
│   └── user/
│       └── dto.py
└── infrastructure/
    ├── persistence/
    │   └── repository.py
    └── payment/
        └── gateway.py
```

## When to Use DDD

- Complex domain logic with sophisticated business rules
- Large systems that need clear boundaries between modules
- Teams working on different parts of a complex system
- When the domain model needs to evolve with changing business requirements

## DDD and Layered Architecture

DDD naturally pairs with layered architecture:

- **Domain Layer**: Entities, value objects, aggregates, domain services
- **Application Layer**: Use cases, application services, DTOs
- **Infrastructure Layer**: Repositories, external services
- **Presentation Layer**: APIs, UI controllers

Dependencies should flow from outer layers toward the domain layer, never the reverse.
