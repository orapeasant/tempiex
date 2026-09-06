// Package api serves the worker's HTTP inspection surface: run listings, live
// event streams, and the registered activity catalog. It is consumed by the
// tempiex CLI and the web UI.
package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/worker/activity"
	"github.com/tempiex/worker/run"
)

// Options configures the router.
type Options struct {
	Runs        *run.Manager
	Registry    *activity.Registry
	Client      workflowservicev1.WorkflowServiceClient
	Namespace   string
	CORSOrigins []string
	Log         zerolog.Logger
}

// NewRouter builds the worker's HTTP handler.
//
// The /api/sessions and /api/tools paths are deprecated aliases of /api/runs
// and /api/activities, kept so existing `tempiex pi ...` CLI commands keep
// working while they are renamed to `tempiex worker ...`.
func NewRouter(opts Options) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(opts.CORSOrigins))

	r.Get("/health", healthHandler(opts.Client != nil, opts.Registry))

	r.Route("/api", func(r chi.Router) {
		r.Get("/runs", listRunsHandler(opts.Runs))
		r.Get("/runs/{id}", getRunHandler(opts.Runs))
		r.Post("/runs/{id}/cancel", cancelRunHandler(opts))
		r.Get("/runs/{id}/events", sseHandler(opts.Runs))
		r.Get("/activities", activitiesHandler(opts.Registry))

		// Deprecated aliases.
		r.Get("/sessions", listRunsHandler(opts.Runs))
		r.Get("/sessions/{id}", getRunHandler(opts.Runs))
		r.Post("/sessions/{id}/cancel", cancelRunHandler(opts))
		r.Get("/sessions/{id}/events", sseHandler(opts.Runs))
		r.Get("/tools", activitiesHandler(opts.Registry))
	})

	return r
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isAllowed(origin, origins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
