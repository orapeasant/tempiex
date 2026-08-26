package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tempiex/pi/internal/session"
)

func sessionsHandler(mgr *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := mgr.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if sessions == nil {
			sessions = []session.Session{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sessions": sessions})
	}
}

func sessionHandler(mgr *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		sess, err := mgr.Get(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sess)
	}
}

func cancelSessionHandler(mgr *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := mgr.Close(r.Context(), id, session.StatusCancelled); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}
