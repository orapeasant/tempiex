package run

import (
	"context"
	"testing"
	"time"
)

func newTestManager(t *testing.T, maxCompleted int) *Manager {
	t.Helper()
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return NewManager(store, NewBus(), maxCompleted)
}

func testRun(id string) Run {
	return Run{
		ID:            id,
		Namespace:     "default",
		TaskQueue:     "worker-default",
		WorkflowID:    "wf-" + id,
		WorkflowRunID: "wfr-" + id,
		ActivityID:    "act-" + id,
		ActivityType:  "echo",
		StartedAt:     time.Now().UTC(),
	}
}

func TestOpenAndGet(t *testing.T) {
	m := newTestManager(t, 100)
	ctx := context.Background()

	if err := m.Open(ctx, testRun("r1")); err != nil {
		t.Fatal(err)
	}

	got, err := m.Get(ctx, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkflowID != "wf-r1" {
		t.Fatalf("unexpected workflow id: %s", got.WorkflowID)
	}
	if got.Status != StatusActive {
		t.Fatalf("expected active, got %s", got.Status)
	}
}

func TestCloseSetsStatusAndEndedAt(t *testing.T) {
	m := newTestManager(t, 100)
	ctx := context.Background()
	_ = m.Open(ctx, testRun("r2"))

	if err := m.Close(ctx, "r2", StatusCompleted); err != nil {
		t.Fatal(err)
	}

	got, _ := m.Get(ctx, "r2")
	if got.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", got.Status)
	}
	if got.EndedAt == nil {
		t.Fatal("expected EndedAt to be set")
	}
}

func TestListFilters(t *testing.T) {
	m := newTestManager(t, 100)
	ctx := context.Background()

	for _, id := range []string{"a", "b", "c"} {
		_ = m.Open(ctx, testRun(id))
	}
	_ = m.Close(ctx, "a", StatusFailed)

	all, err := m.List(ctx, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}

	active, err := m.List(ctx, Filter{Status: StatusActive})
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(active))
	}

	byWorkflow, _ := m.List(ctx, Filter{WorkflowID: "wf-b"})
	if len(byWorkflow) != 1 || byWorkflow[0].ID != "b" {
		t.Fatalf("unexpected workflow filter result: %v", byWorkflow)
	}
}

func TestPublishPersistsAndStreams(t *testing.T) {
	m := newTestManager(t, 100)
	ctx := context.Background()
	_ = m.Open(ctx, testRun("r3"))

	ch, unsub := m.Bus().Subscribe("r3")
	defer unsub()

	err := m.Publish(ctx, Event{ID: "e1", Type: EventExecutionStart, RunID: "r3"})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-ch:
		if got.ID != "e1" {
			t.Fatalf("unexpected live event: %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for live event")
	}

	stored, err := m.LoadEvents(ctx, "r3")
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 || stored[0].ID != "e1" {
		t.Fatalf("unexpected stored events: %v", stored)
	}
	if stored[0].Timestamp.IsZero() {
		t.Fatal("expected Publish to stamp a timestamp")
	}
}

func TestPruneKeepsActiveAndNewestCompleted(t *testing.T) {
	m := newTestManager(t, 1)
	ctx := context.Background()

	for _, id := range []string{"old", "new"} {
		r := testRun(id)
		if id == "old" {
			r.StartedAt = time.Now().Add(-time.Hour).UTC()
		}
		_ = m.Open(ctx, r)
	}
	_ = m.Open(ctx, testRun("live"))

	_ = m.Close(ctx, "old", StatusCompleted)
	_ = m.Close(ctx, "new", StatusCompleted)

	remaining, _ := m.List(ctx, Filter{})
	ids := map[string]bool{}
	for _, r := range remaining {
		ids[r.ID] = true
	}
	if ids["old"] {
		t.Fatal("expected the oldest completed run to be pruned")
	}
	if !ids["new"] || !ids["live"] {
		t.Fatalf("expected newest completed and active runs to survive: %v", ids)
	}
}
