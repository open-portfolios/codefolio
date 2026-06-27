# State

Allow an object to alter its behavior when its internal state changes. The object will appear to change its class.

## Intent

- Allow an object to change its behavior when its internal state changes
- Let the object appear to change its class
- Encapsulate state-specific behavior

## Implementation

```go
package main

import "fmt"

type State interface {
	InsertCoin(machine *VendingMachine)
	EjectCoin(machine *VendingMachine)
	Dispense(machine *VendingMachine)
}

type NoCoinState struct{}

func NewNoCoinState() *NoCoinState {
	return &NoCoinState{}
}

func (s *NoCoinState) InsertCoin(machine *VendingMachine) {
	fmt.Println("Coin inserted")
	machine.SetState(machine.GetHasCoinState())
}

func (s *NoCoinState) EjectCoin(machine *VendingMachine) {
	fmt.Println("No coin to eject")
}

func (s *NoCoinState) Dispense(machine *VendingMachine) {
	fmt.Println("Insert coin first")
}

type HasCoinState struct{}

func NewHasCoinState() *HasCoinState {
	return &HasCoinState{}
}

func (s *HasCoinState) InsertCoin(machine *VendingMachine) {
	fmt.Println("Coin already inserted")
}

func (s *HasCoinState) EjectCoin(machine *VendingMachine) {
	fmt.Println("Coin returned")
	machine.SetState(machine.GetNoCoinState())
}

func (s *HasCoinState) Dispense(machine *VendingMachine) {
	fmt.Println("Item dispensed")
	machine.SetState(machine.GetNoCoinState())
}

type VendingMachine struct {
	noCoinState    State
	hasCoinState    State
	currentState   State
}

func NewVendingMachine() *VendingMachine {
	vm := &VendingMachine{}
	vm.noCoinState = NewNoCoinState()
	vm.hasCoinState = NewHasCoinState()
	vm.currentState = vm.noCoinState
	return vm
}

func (m *VendingMachine) SetState(state State) {
	m.currentState = state
}

func (m *VendingMachine) GetNoCoinState() State    { return m.noCoinState }
func (m *VendingMachine) GetHasCoinState() State   { return m.hasCoinState }

func (m *VendingMachine) InsertCoin() {
	m.currentState.InsertCoin(m)
}

func (m *VendingMachine) EjectCoin() {
	m.currentState.EjectCoin(m)
}

func (m *VendingMachine) Dispense() {
	m.currentState.Dispense(m)
}

func main() {
	machine := NewVendingMachine()
	
	machine.Dispense()
	machine.InsertCoin()
	machine.Dispense()
}
```

## When to Use

- When an object's behavior depends on its state, and it must change its behavior at runtime
- When operations have large conditional statements that depend on the object's state
- When you want to avoid large if-else or switch statements
