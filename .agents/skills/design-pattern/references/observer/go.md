# Observer

Define a one-to-many dependency between objects so that when one object changes state, all its dependents are notified.

## Intent

- Define a one-to-many dependency between objects
- Notify dependents when an object changes state
- Loose coupling between subject and observers

## Implementation

```go
package main

import "fmt"

type Observer interface {
	Update(message string)
}

type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify()
}

type NewsAgency struct {
	observers []Observer
	news      string
}

func (s *NewsAgency) Attach(observer Observer) {
	s.observers = append(s.observers, observer)
}

func (s *NewsAgency) Detach(observer Observer) {
	for i, o := range s.observers {
		if o == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

func (s *NewsAgency) Notify() {
	for _, observer := range s.observers {
		observer.Update(s.news)
	}
}

func NewNewsAgency() *NewsAgency {
	return &NewsAgency{
		observers: make([]Observer, 0),
	}
}

func (s *NewsAgency) SetNews(news string) {
	s.news = news
	s.Notify()
}

type NewsChannel struct {
	name string
}

func NewNewsChannel(name string) *NewsChannel {
	return &NewsChannel{name: name}
}

func (c *NewsChannel) Update(news string) {
	fmt.Println(c.name, "received:", news)
}

func main() {
	agency := NewNewsAgency()
	
	channel1 := NewNewsChannel("Channel 1")
	channel2 := NewNewsChannel("Channel 2")
	
	agency.Attach(channel1)
	agency.Attach(channel2)
	
	agency.SetNews("Breaking news!")
}
```

## When to Use

- When changes to one object require changing others, and you don't know how many objects need to change
- When an object should notify other objects without making assumptions about these objects
- When you need loose coupling between objects
