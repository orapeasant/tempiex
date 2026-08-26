package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "version": "0.1.0"})
}

func Settings(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"auth":    map[string]any{"enabled": false},
		"codec":   map[string]any{"endpoint": ""},
		"version": "0.1.0",
	})
}

func ClusterInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"clusterInfo": map[string]any{
			"supportedClients": map[string]any{},
			"serverVersion":    "0.1.0",
			"clusterId":        "tempiex-dev",
			"clusterName":      "tempiex-dev",
			"historyShardCount": 4,
		},
	})
}
