package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tempiex/server/internal/api"
	"github.com/tempiex/server/internal/config"
)

//go:embed ui/*
var uiFS embed.FS

func New(cfg *config.Config, db *pgxpool.Pool) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))

	nsHandler := &api.NamespaceHandler{DB: db}
	wfHandler := &api.WorkflowHandler{DB: db}

	// Health
	e.GET("/health", api.Health)
	e.GET("/ready", func(c echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// API
	v1 := e.Group("/api/v1")
	v1.GET("/settings", api.Settings)
	v1.GET("/cluster", api.ClusterInfo)
	v1.GET("/namespaces", nsHandler.List)
	v1.POST("/namespaces", nsHandler.Create)
	v1.GET("/namespaces/:ns", nsHandler.Get)
	v1.GET("/namespaces/:ns/workflows", wfHandler.List)
	v1.GET("/namespaces/:ns/workflows/:wfId/:runId", wfHandler.Get)
	v1.GET("/namespaces/:ns/workflows/:wfId/:runId/history", wfHandler.History)
	v1.POST("/namespaces/:ns/workflows/:wfId/:runId/terminate", wfHandler.Terminate)
	v1.POST("/namespaces/:ns/workflows/:wfId/:runId/cancel", wfHandler.Cancel)

	// Static UI files
	subFS, err := fs.Sub(uiFS, "ui")
	if err == nil {
		fileServer := http.FileServer(http.FS(subFS))
		e.GET("/*", echo.WrapHandler(fileServer), func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				path := c.Request().URL.Path
				if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/ready" {
					return next(c)
				}
				// Try serving the file; if not found, serve index.html (SPA fallback)
				f, err := subFS.Open(strings.TrimPrefix(path, "/"))
				if err != nil {
					indexData, err2 := fs.ReadFile(subFS, "index.html")
					if err2 != nil {
						return echo.NewHTTPError(http.StatusNotFound)
					}
					return c.HTMLBlob(http.StatusOK, indexData)
				}
				f.Close()
				return next(c)
			}
		})
	}

	return e
}

func Start(e *echo.Echo, cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return e.Start(addr)
}
