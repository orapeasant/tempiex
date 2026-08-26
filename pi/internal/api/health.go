package api

import (
	"encoding/json"
	"net/http"
)

type healthCheckFunc func() bool

func healthHandler(isConnected healthCheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "disconnected"
		if isConnected() {
			status = "connected"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"tempiex": status,
		})
	}
}
