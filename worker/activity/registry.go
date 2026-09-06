package activity

import (
	"fmt"
	"sort"
	"sync"
)

// Registry maps activity type names to implementations. It is safe for
// concurrent use; registration is expected to happen at startup.
type Registry struct {
	mu         sync.RWMutex
	activities map[string]AnyActivity
}

func NewRegistry() *Registry {
	return &Registry{activities: make(map[string]AnyActivity)}
}

// Register adds a typed activity. It panics on a duplicate name: two
// implementations answering to one activity type is a wiring bug, and failing
// at startup beats dispatching to whichever won the race.
func Register[I, O any](r *Registry, a Activity[I, O]) {
	r.RegisterAny(&a)
}

// RegisterAny adds an already type-erased activity.
func (r *Registry) RegisterAny(a AnyActivity) {
	name := a.Descriptor().Name
	if name == "" {
		panic("activity: registered activity has no name")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.activities[name]; exists {
		panic(fmt.Sprintf("activity: duplicate activity name %q", name))
	}
	r.activities[name] = a
}

func (r *Registry) Get(name string) (AnyActivity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.activities[name]
	return a, ok
}

// List returns the registered activities, ordered by name.
func (r *Registry) List() []Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	descs := make([]Descriptor, 0, len(r.activities))
	for _, a := range r.activities {
		descs = append(descs, a.Descriptor())
	}
	sort.Slice(descs, func(i, j int) bool { return descs[i].Name < descs[j].Name })
	return descs
}
