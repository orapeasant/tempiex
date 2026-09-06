package run

import (
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	b := NewBus()
	ch, unsub := b.Subscribe("run1")
	defer unsub()

	b.Publish(Event{ID: "1", Type: EventExecutionStart, RunID: "run1", Timestamp: time.Now()})

	select {
	case got := <-ch:
		if got.ID != "1" {
			t.Fatalf("expected ID 1, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestReplayBufferedEvents(t *testing.T) {
	b := NewBus()
	b.Publish(Event{ID: "pre", Type: EventExecutionStart, RunID: "run2", Timestamp: time.Now()})

	ch, unsub := b.Subscribe("run2")
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

func TestUnsubscribeIsIdempotent(t *testing.T) {
	b := NewBus()
	_, unsub := b.Subscribe("run3")
	unsub()
	unsub() // must not panic on double close

	// Publishing to a run with no subscribers must not block or panic.
	b.Publish(Event{ID: "x", RunID: "run3", Timestamp: time.Now()})
}

func TestPublishDropsForFullSubscriber(t *testing.T) {
	b := NewBus()
	dropped := 0
	b.OnDrop = func(string) { dropped++ }

	ch, unsub := b.Subscribe("run4")
	defer unsub()

	// Fill the subscriber's buffer, then overflow it.
	for i := 0; i < bufferSize+10; i++ {
		b.Publish(Event{ID: "e", RunID: "run4", Timestamp: time.Now()})
	}
	if dropped == 0 {
		t.Fatal("expected drops once the subscriber buffer filled")
	}
	if len(ch) != bufferSize {
		t.Fatalf("expected a full buffer of %d, got %d", bufferSize, len(ch))
	}
}
