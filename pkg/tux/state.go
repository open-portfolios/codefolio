package tux

import "sync"

// State holds a reactive value of type T.
//
// Get(ctx) reads the value and subscribes the current build pass to this state.
// Set(v) and Update(fn) write the value and notify all current subscribers.
//
// Subscribers are cleared on each Set/Update call. Components re-subscribe
// automatically during the next Build pass via Get(ctx).
type State[T any] struct {
	mu         sync.Mutex
	value      T
	watchers   []func()
	registered uint64 // epoch of the last build pass that registered a watcher
}

// NewState creates a new State with the given initial value.
func NewState[T any](initial T) *State[T] {
	return &State[T]{value: initial}
}

// Get returns the current value and subscribes the current build pass so that
// a future Set or Update will trigger a rebuild.
//
// Each build pass has a unique epoch. If this state has already registered a
// watcher for the current epoch, Get skips re-registration — one watcher per
// state per build pass is sufficient because all watchers call app.MarkDirty.
func (s *State[T]) Get(ctx BuildContext) T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.notify != nil && ctx.epoch != s.registered {
		s.watchers = append(s.watchers, ctx.notify)
		s.registered = ctx.epoch
	}
	return s.value
}

// Set replaces the value and notifies all subscribers.
// Subscribers are cleared so they re-subscribe during the next Build pass.
// Safe to call from any goroutine.
func (s *State[T]) Set(v T) {
	s.mu.Lock()
	s.value = v
	watchers := s.watchers
	s.watchers = nil
	s.mu.Unlock()
	for _, w := range watchers {
		w()
	}
}

// Update applies fn to the current value, stores the result, and notifies
// all subscribers. Safe to call from any goroutine.
func (s *State[T]) Update(fn func(T) T) {
	s.mu.Lock()
	s.value = fn(s.value)
	watchers := s.watchers
	s.watchers = nil
	s.mu.Unlock()
	for _, w := range watchers {
		w()
	}
}

// Value returns the current value without subscribing to changes.
// Use this when you need to read state outside of a Build pass.
// Safe to call from any goroutine.
func (s *State[T]) Value() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}
