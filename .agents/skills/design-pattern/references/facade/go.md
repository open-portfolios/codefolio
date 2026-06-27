# Facade

Provide a unified interface to a set of interfaces in a subsystem. Facade defines a higher-level interface that makes the subsystem easier to use.

## Intent

- Provide a unified interface to a set of interfaces in a subsystem
- Make a subsystem easier to use
- Wrap a complex subsystem with a simpler interface

## Implementation

```go
package main

import "fmt"

type CPU struct{}

func NewCPU() *CPU {
	return &CPU{}
}

func (c *CPU) Freeze() {
	fmt.Println("CPU freeze")
}

func (c *CPU) Jump(position int64) {
	fmt.Println("CPU jump to", position)
}

func (c *CPU) Execute() {
	fmt.Println("CPU execute")
}

type Memory struct{}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Load(position int64, data []byte) {
	fmt.Println("Memory load at", position)
}

type HardDrive struct{}

func NewHardDrive() *HardDrive {
	return &HardDrive{}
}

func (h *HardDrive) Read(lba int64, size int) []byte {
	fmt.Println("HardDrive read", size, "bytes")
	return make([]byte, size)
}

// Facade
type ComputerFacade struct {
	cpu       *CPU
	memory    *Memory
	hardDrive *HardDrive
}

func NewComputerFacade() *ComputerFacade {
	return &ComputerFacade{
		cpu:       NewCPU(),
		memory:    NewMemory(),
		hardDrive: NewHardDrive(),
	}
}

func (c *ComputerFacade) Start() {
	c.cpu.Freeze()
	c.memory.Load(0, c.hardDrive.Read(0, 1024))
	c.cpu.Jump(0)
	c.cpu.Execute()
}

// Client code
func main() {
	computer := NewComputerFacade()
	computer.Start()
}
```

## When to Use

- When you want to provide a simple interface to a complex subsystem
- When there are many dependencies between clients and implementation types
- When you want to layer your subsystems
