package api

import (
	"net/http"

	"github.com/tempiex/worker/activity"
)

func healthHandler(connected bool, registry *activity.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "disconnected"
		if connected {
			status = "connected"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":     "ok",
			"tempiex":    status,
			"activities": len(registry.List()),
		})
	}
}
