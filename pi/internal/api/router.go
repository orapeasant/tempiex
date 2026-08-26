package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/pi/internal/event"
	"github.com/tempiex/pi/internal/session"
	"github.com/tempiex/pi/internal/tool"
)

func NewRouter(
	sessions *session.Manager,
	bus *event.Bus,
	registry *tool.Registry,
	tempiexClient workflowservicev1.WorkflowServiceClient,
	corsOrigins []string,
	log zerolog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(corsOrigins))

	isConnected := func() bool {
		if tempiexClient == nil {
			return false
		}
		return true
	}

	r.Get("/health", healthHandler(isConnected))
	r.Route("/api", func(r chi.Router) {
		r.Get("/sessions", sessionsHandler(sessions))
		r.Get("/sessions/{id}", sessionHandler(sessions))
		r.Post("/sessions/{id}/cancel", cancelSessionHandler(sessions))
		r.Get("/sessions/{id}/events", sseHandler(bus))
		r.Get("/tools", toolsHandler(registry))
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
		if strings.EqualFold(a, origin) || a == "*" {
			return true
		}
	}
	return false
}
