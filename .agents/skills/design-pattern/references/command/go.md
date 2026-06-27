# Command

Encapsulate a request as an object, letting you parameterize clients with different requests.

## Intent

- Encapsulate a request as an object
- Parameterize objects with different requests
- Support undoable operations
- Queue or log requests

## Implementation

```go
package main

import "fmt"

type Command interface {
	Execute()
}

type Light struct{}

func NewLight() *Light {
	return &Light{}
}

func (l *Light) On() {
	fmt.Println("Light is on")
}

func (l *Light) Off() {
	fmt.Println("Light is off")
}

type LightOnCommand struct {
	light *Light
}

func NewLightOnCommand(light *Light) *LightOnCommand {
	return &LightOnCommand{light: light}
}

func (c *LightOnCommand) Execute() {
	c.light.On()
}

type LightOffCommand struct {
	light *Light
}

func NewLightOffCommand(light *Light) *LightOffCommand {
	return &LightOffCommand{light: light}
}

func (c *LightOffCommand) Execute() {
	c.light.Off()
}

type RemoteControl struct {
	command Command
}

func NewRemoteControl() *RemoteControl {
	return &RemoteControl{}
}

func (r *RemoteControl) SetCommand(command Command) {
	r.command = command
}

func (r *RemoteControl) PressButton() {
	r.command.Execute()
}

func main() {
	light := NewLight()
	onCommand := NewLightOnCommand(light)
	offCommand := NewLightOffCommand(light)
	
	remote := NewRemoteControl()
	
	remote.SetCommand(onCommand)
	remote.PressButton()
	
	remote.SetCommand(offCommand)
	remote.PressButton()
}
```

## When to Use

- When you want to parameterize objects with actions
- When you want to queue, specify, and execute requests at different times
- When you want to support undo operations
