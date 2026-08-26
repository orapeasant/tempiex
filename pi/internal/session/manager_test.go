package session

import (
	"context"
	"testing"
	"time"
)

func openMemory(t *testing.T) *Manager {
	t.Helper()
	store, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return NewManager(store)
}

func TestOpenAndGet(t *testing.T) {
	m := openMemory(t)
	ctx := context.Background()

	sess := Session{
		ID:         "s1",
		WorkflowID: "wf1",
		RunID:      "r1",
		Status:     StatusActive,
		CreatedAt:  time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt:  time.Now().UTC().Truncate(time.Millisecond),
	}
	if err := m.Open(ctx, sess); err != nil {
		t.Fatal(err)
	}

	got, err := m.Get(ctx, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkflowID != "wf1" {
		t.Fatalf("unexpected workflow id: %s", got.WorkflowID)
	}
}

func TestCloseSession(t *testing.T) {
	m := openMemory(t)
	ctx := context.Background()

	sess := Session{ID: "s2", WorkflowID: "wf2", RunID: "r2", Status: StatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = m.Open(ctx, sess)

	if err := m.Close(ctx, "s2", StatusCompleted); err != nil {
		t.Fatal(err)
	}
	got, _ := m.Get(ctx, "s2")
	if got.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", got.Status)
	}
}

func TestList(t *testing.T) {
	m := openMemory(t)
	ctx := context.Background()

	for _, id := range []string{"a", "b", "c"} {
		_ = m.Open(ctx, Session{ID: id, WorkflowID: "wf", RunID: "r", Status: StatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	}
	list, err := m.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}
}

func TestAppendAndLoadEvents(t *testing.T) {
	m := openMemory(t)
	ctx := context.Background()
	_ = m.Open(ctx, Session{ID: "s3", WorkflowID: "wf", RunID: "r", Status: StatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	e := StoredEvent{ID: "e1", SessionID: "s3", Type: "execution_start", Timestamp: time.Now()}
	if err := m.AppendEvent(ctx, "s3", e); err != nil {
		t.Fatal(err)
	}

	events, err := m.LoadEvents(ctx, "s3")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != "e1" {
		t.Fatalf("unexpected events: %v", events)
	}
}
