# Memento

Without violating encapsulation, capture and externalize an object's internal state so that the object can be restored to this state later.

## Intent

- Capture and save an object's internal state
- Restore the object to a previous state
- Support undo operations

## Implementation

```go
package main

type Memento struct {
	state string
}

type Originator struct {
	state string
}

func NewOriginator() *Originator {
	return &Originator{}
}

func (o *Originator) SetState(state string) {
	o.state = state
}

func (o *Originator) GetState() string {
	return o.state
}

func (o *Originator) Save() *Memento {
	return &Memento{state: o.state}
}

func (o *Originator) Restore(memento *Memento) {
	o.state = memento.state
}

type Caretaker struct {
	mementos []*Memento
	originator *Originator
}

func NewCaretaker(originator *Originator) *Caretaker {
	return &Caretaker{originator: originator}
}

func (c *Caretaker) Backup() {
	c.mementos = append(c.mementos, c.originator.Save())
}

func (c *Caretaker) Undo() {
	if len(c.mementos) == 0 {
		return
	}
	memento := c.mementos[len(c.mementos)-1]
	c.mementos = c.mementos[:len(c.mementos)-1]
	c.originator.Restore(memento)
}

func main() {
	originator := NewOriginator()
	caretaker := NewCaretaker(originator)
	
	originator.SetState("State 1")
	caretaker.Backup()
	
	originator.SetState("State 2")
	caretaker.Backup()
	
	originator.SetState("State 3")
	
	println("Current:", originator.GetState())
	
	caretaker.Undo()
	println("After undo:", originator.GetState())
}
```

## When to Use

- When you need to save and restore an object's state
- When you want to implement undo functionality
- When you need to capture snapshots of an object's state
