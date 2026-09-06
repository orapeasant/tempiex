package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tempiex/worker/run"
)

// sseHandler streams a run's events. Events already stored for the run are
// replayed first, then the live stream continues, so a client that connects
// mid-run sees the whole execution.
func sseHandler(mgr *run.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := chi.URLParam(r, "id")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Subscribe before replaying history so nothing published in between is
		// lost; the bus replays its own buffer too, so a client may see a
		// duplicate event id, which SSE consumers dedupe on.
		ch, unsub := mgr.Bus().Subscribe(runID)
		defer unsub()

		stored, err := mgr.LoadEvents(r.Context(), runID)
		if err == nil {
			for _, e := range stored {
				if !writeEvent(w, flusher, e) {
					return
				}
			}
		}

		for {
			select {
			case <-r.Context().Done():
				return
			case e, ok := <-ch:
				if !ok {
					return
				}
				if !writeEvent(w, flusher, e) {
					return
				}
			}
		}
	}
}

func writeEvent(w http.ResponseWriter, flusher http.Flusher, e run.Event) bool {
	data, err := json.Marshal(e)
	if err != nil {
		return true
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return false
	}
	flusher.Flush()
	return true
}
