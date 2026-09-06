package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/tempiex/worker/activity"
	"github.com/tempiex/worker/run"
)

func newTestServer(t *testing.T) (*httptest.Server, *run.Manager) {
	t.Helper()

	store, err := run.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	runs := run.NewManager(store, run.NewBus(), 100)

	registry := activity.NewRegistry()
	activity.Register(registry, activity.Activity[map[string]any, map[string]any]{
		Name:          "echo",
		Description:   "Return the input unchanged",
		ExecutionMode: "parallel",
		Execute:       func(_ activity.Context, in map[string]any) (map[string]any, error) { return in, nil },
	})

	srv := httptest.NewServer(NewRouter(Options{
		Runs:      runs,
		Registry:  registry,
		Namespace: "default",
		Log:       zerolog.Nop(),
	}))
	t.Cleanup(srv.Close)

	return srv, runs
}

func getJSON(t *testing.T, url string, into any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		t.Fatal(err)
	}
}

func TestHealth(t *testing.T) {
	srv, _ := newTestServer(t)

	var body map[string]any
	getJSON(t, srv.URL+"/health", &body)

	if body["status"] != "ok" {
		t.Fatalf("unexpected health body: %v", body)
	}
	// No client was configured for this server.
	if body["tempiex"] != "disconnected" {
		t.Fatalf("expected disconnected, got %v", body["tempiex"])
	}
}

func TestListRunsAndFilter(t *testing.T) {
	srv, runs := newTestServer(t)
	ctx := context.Background()

	_ = runs.Open(ctx, run.Run{ID: "r1", WorkflowID: "wf1", TaskQueue: "q1", StartedAt: time.Now()})
	_ = runs.Open(ctx, run.Run{ID: "r2", WorkflowID: "wf2", TaskQueue: "q2", StartedAt: time.Now()})
	_ = runs.Close(ctx, "r2", run.StatusCompleted)

	var all struct{ Runs []run.Run }
	getJSON(t, srv.URL+"/api/runs", &all)
	if len(all.Runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(all.Runs))
	}

	var active struct{ Runs []run.Run }
	getJSON(t, srv.URL+"/api/runs?status=active", &active)
	if len(active.Runs) != 1 || active.Runs[0].ID != "r1" {
		t.Fatalf("unexpected active runs: %v", active.Runs)
	}

	var byQueue struct{ Runs []run.Run }
	getJSON(t, srv.URL+"/api/runs?taskQueue=q2", &byQueue)
	if len(byQueue.Runs) != 1 || byQueue.Runs[0].ID != "r2" {
		t.Fatalf("unexpected queue-filtered runs: %v", byQueue.Runs)
	}
}

func TestGetRunNotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/api/runs/missing")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestActivitiesCatalog(t *testing.T) {
	srv, _ := newTestServer(t)

	var body struct {
		Activities []activity.Descriptor
		Tools      []activity.Descriptor
	}
	getJSON(t, srv.URL+"/api/activities", &body)

	if len(body.Activities) != 1 || body.Activities[0].Name != "echo" {
		t.Fatalf("unexpected activities: %v", body.Activities)
	}
	// "tools" is the deprecated alias key and must mirror "activities".
	if len(body.Tools) != len(body.Activities) {
		t.Fatalf("expected tools alias to mirror activities: %v", body.Tools)
	}
}

func TestDeprecatedSessionAliases(t *testing.T) {
	srv, runs := newTestServer(t)
	_ = runs.Open(context.Background(), run.Run{ID: "r1", WorkflowID: "wf1", StartedAt: time.Now()})

	var viaSessions struct{ Runs []run.Run }
	getJSON(t, srv.URL+"/api/sessions", &viaSessions)
	if len(viaSessions.Runs) != 1 {
		t.Fatalf("expected the sessions alias to list runs, got %v", viaSessions.Runs)
	}

	var viaTools struct{ Activities []activity.Descriptor }
	getJSON(t, srv.URL+"/api/tools", &viaTools)
	if len(viaTools.Activities) != 1 {
		t.Fatalf("expected the tools alias to list activities, got %v", viaTools.Activities)
	}
}

func TestSSEStreamsStoredAndLiveEvents(t *testing.T) {
	srv, runs := newTestServer(t)
	ctx := context.Background()

	_ = runs.Open(ctx, run.Run{ID: "r1", StartedAt: time.Now()})
	_ = runs.Publish(ctx, run.Event{ID: "stored", Type: run.EventExecutionStart, RunID: "r1"})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/runs/r1/events", nil)
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(reqCtx))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	// The stored event must arrive without any further publishing.
	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); !strings.Contains(got, "stored") {
		t.Fatalf("expected the stored event to be replayed, got %q", got)
	}
}
