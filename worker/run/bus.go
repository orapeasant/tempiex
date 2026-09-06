package run

import "sync"

const bufferSize = 1000

// Bus fans Events out to subscribers, keyed by run id, and keeps a bounded
// per-run backlog so a late subscriber can replay recent history before
// following the live stream.
//
// Publishing never blocks: if a subscriber's channel is full its event is
// dropped and OnDrop (when set) is called. Activity execution must not be
// stalled by a slow HTTP client.
type Bus struct {
	mu          sync.Mutex
	buffers     map[string][]Event
	subscribers map[string][]chan Event

	// OnDrop is called whenever an event is dropped for a full subscriber.
	// Set it before the bus is used; it must not call back into the Bus.
	OnDrop func(runID string)
}

func NewBus() *Bus {
	return &Bus{
		buffers:     make(map[string][]Event),
		subscribers: make(map[string][]chan Event),
	}
}

func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	dropped := 0

	buf := append(b.buffers[e.RunID], e)
	if len(buf) > bufferSize {
		buf = buf[len(buf)-bufferSize:]
	}
	b.buffers[e.RunID] = buf

	for _, ch := range b.subscribers[e.RunID] {
		select {
		case ch <- e:
		default:
			dropped++
		}
	}
	onDrop := b.OnDrop
	runID := e.RunID
	b.mu.Unlock()

	if onDrop != nil {
		for i := 0; i < dropped; i++ {
			onDrop(runID)
		}
	}
}

// Subscribe returns a channel of events for one run plus an unsubscribe func.
// Buffered events published before the call are replayed on the channel first.
func (b *Bus) Subscribe(runID string) (<-chan Event, func()) {
	ch := make(chan Event, bufferSize)

	b.mu.Lock()
	snapshot := make([]Event, len(b.buffers[runID]))
	copy(snapshot, b.buffers[runID])
	b.subscribers[runID] = append(b.subscribers[runID], ch)
	b.mu.Unlock()

	go func() {
		for _, e := range snapshot {
			ch <- e
		}
	}()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			subs := b.subscribers[runID]
			for i, s := range subs {
				if s == ch {
					b.subscribers[runID] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
			close(ch)
		})
	}

	return ch, unsub
}
