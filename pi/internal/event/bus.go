package event

import "sync"

const bufferSize = 1000

type Bus struct {
	mu          sync.Mutex
	buffers     map[string][]Event
	subscribers map[string][]chan Event
}

func New() *Bus {
	return &Bus{
		buffers:     make(map[string][]Event),
		subscribers: make(map[string][]chan Event),
	}
}

func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	buf := b.buffers[e.SessionID]
	buf = append(buf, e)
	if len(buf) > bufferSize {
		buf = buf[len(buf)-bufferSize:]
	}
	b.buffers[e.SessionID] = buf

	for _, ch := range b.subscribers[e.SessionID] {
		select {
		case ch <- e:
		default:
		}
	}
}

func (b *Bus) Subscribe(sessionID string) (<-chan Event, func()) {
	ch := make(chan Event, bufferSize)

	b.mu.Lock()
	snapshot := make([]Event, len(b.buffers[sessionID]))
	copy(snapshot, b.buffers[sessionID])
	b.subscribers[sessionID] = append(b.subscribers[sessionID], ch)
	b.mu.Unlock()

	go func() {
		for _, e := range snapshot {
			ch <- e
		}
	}()

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[sessionID]
		for i, s := range subs {
			if s == ch {
				b.subscribers[sessionID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsub
}
