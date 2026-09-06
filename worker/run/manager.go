package run

import (
	"context"
	"time"
)

// Manager is the write path for runs: it persists lifecycle transitions and
// mirrors every event to both the Bus (live consumers) and the Store (replay).
type Manager struct {
	store        *Store
	bus          *Bus
	maxCompleted int
}

func NewManager(store *Store, bus *Bus, maxCompleted int) *Manager {
	return &Manager{store: store, bus: bus, maxCompleted: maxCompleted}
}

// Bus exposes the event bus so read-side consumers (SSE) can subscribe.
func (m *Manager) Bus() *Bus { return m.bus }

// Open records a new active run.
func (m *Manager) Open(ctx context.Context, r Run) error {
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now().UTC()
	}
	r.Status = StatusActive
	return m.store.SaveRun(ctx, &r)
}

// Close transitions a run to a terminal status and prunes old finished runs.
func (m *Manager) Close(ctx context.Context, id string, status Status) error {
	r, err := m.store.LoadRun(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	r.Status = status
	r.EndedAt = &now
	if err := m.store.SaveRun(ctx, r); err != nil {
		return err
	}
	return m.store.PruneCompleted(ctx, m.maxCompleted)
}

func (m *Manager) Get(ctx context.Context, id string) (*Run, error) {
	return m.store.LoadRun(ctx, id)
}

func (m *Manager) List(ctx context.Context, f Filter) ([]Run, error) {
	all, err := m.store.ListRuns(ctx)
	if err != nil {
		return nil, err
	}
	runs := make([]Run, 0, len(all))
	for _, r := range all {
		if f.matches(r) {
			runs = append(runs, r)
		}
	}
	return runs, nil
}

// Publish mirrors an event to live subscribers and to the store. A store
// failure is not fatal: the live stream still gets the event.
func (m *Manager) Publish(ctx context.Context, e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	m.bus.Publish(e)
	return m.store.AppendEvent(ctx, &e)
}

func (m *Manager) LoadEvents(ctx context.Context, runID string) ([]Event, error) {
	return m.store.LoadEvents(ctx, runID)
}
