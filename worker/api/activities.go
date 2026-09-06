package api

import (
	"errors"
	"net/http"

	"github.com/tempiex/worker/activity"
)

var errNoTempiexClient = errors.New("no tempiex connection configured")

func activitiesHandler(registry *activity.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		descs := registry.List()
		if descs == nil {
			descs = []activity.Descriptor{}
		}
		// "tools" is the deprecated key kept alongside "activities" so existing
		// CLI output parsing keeps working during the rename.
		writeJSON(w, http.StatusOK, map[string]any{
			"activities": descs,
			"tools":      descs,
		})
	}
}
