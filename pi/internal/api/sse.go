package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tempiex/pi/internal/event"
)

func sseHandler(bus *event.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch, unsub := bus.Subscribe(sessionID)
		defer unsub()

		for {
			select {
			case <-r.Context().Done():
				return
			case e, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(e)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
