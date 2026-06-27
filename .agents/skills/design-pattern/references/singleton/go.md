# Singleton

Ensure a type has only one instance and provide a global point of access to it.

## Intent

- Ensure a type has only one instance
- Provide a global point of access to that instance
- Control object creation at a single point

## Implementation

### Using sync.Once

```go
package main

import (
	"fmt"
	"sync"
)

type singleton struct {
	value string
}

var (
	instance *singleton
	once     sync.Once
)

func GetInstance() *singleton {
	once.Do(func() {
		instance = &singleton{value: "singleton"}
	})
	return instance
}

func main() {
	s1 := GetInstance()
	s2 := GetInstance()
	
	fmt.Println(s1 == s2) // true
}
```

### Using Package-Level Variable

```go
package main

type Config struct {
	Host string
	Port int
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = &Config{
			Host: "localhost",
			Port: 8080,
		}
	}
	return config
}
```

## When to Use

- When exactly one instance of a type is needed
- When you need global access to that instance
- Common use cases: configuration managers, logging, connection pools

## Considerations

- Consider dependency injection instead of Singleton when possible
- Be careful with initialization order in Go
- Using sync.Once ensures thread-safe lazy initialization
