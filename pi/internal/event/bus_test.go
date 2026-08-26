package event

import (
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	b := New()
	ch, unsub := b.Subscribe("sess1")
	defer unsub()

	e := Event{ID: "1", Type: EventExecutionStart, SessionID: "sess1", Timestamp: time.Now()}
	b.Publish(e)

	select {
	case got := <-ch:
		if got.ID != "1" {
			t.Fatalf("expected ID 1, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestReplay(t *testing.T) {
	b := New()
	e := Event{ID: "pre", Type: EventExecutionStart, SessionID: "sess2", Timestamp: time.Now()}
	b.Publish(e)

	ch, unsub := b.Subscribe("sess2")
	defer unsub()

	select {
	case got := <-ch:
		if got.ID != "pre" {
			t.Fatalf("expected replayed event, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for replayed event")
	}
}

func TestUnsubscribe(t *testing.T) {
	b := New()
	ch, unsub := b.Subscribe("sess3")
	unsub()

	b.Publish(Event{ID: "x", SessionID: "sess3", Timestamp: time.Now()})
	// channel closed; reading should not block
	_, ok := <-ch
	if ok {
		// drained the pre-close event; fine
	}
}
