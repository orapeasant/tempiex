package api

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type NamespaceHandler struct {
	DB *pgxpool.Pool
}

func (h *NamespaceHandler) List(c echo.Context) error {
	rows, err := h.DB.Query(c.Request().Context(),
		`SELECT id, name, COALESCE(description,''), retention_days, created_at FROM namespaces ORDER BY name`)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var namespaces []map[string]any
	for rows.Next() {
		var id, name, desc string
		var retDays int
		var createdAt any
		if err := rows.Scan(&id, &name, &desc, &retDays, &createdAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		namespaces = append(namespaces, map[string]any{
			"namespaceInfo": map[string]any{
				"name":        name,
				"id":          id,
				"description": desc,
				"data":        map[string]any{},
			},
			"config": map[string]any{
				"workflowExecutionRetentionTtl": map[string]any{
					"days": retDays,
				},
			},
		})
	}
	if namespaces == nil {
		namespaces = []map[string]any{}
	}
	return c.JSON(http.StatusOK, map[string]any{"namespaces": namespaces})
}

func (h *NamespaceHandler) Get(c echo.Context) error {
	ns := c.Param("ns")
	var id, name, desc string
	var retDays int
	err := h.DB.QueryRow(c.Request().Context(),
		`SELECT id, name, COALESCE(description,''), retention_days FROM namespaces WHERE name=$1`, ns).
		Scan(&id, &name, &desc, &retDays)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("namespace %q not found", ns))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"namespaceInfo": map[string]any{
			"name":        name,
			"id":          id,
			"description": desc,
			"data":        map[string]any{},
		},
		"config": map[string]any{
			"workflowExecutionRetentionTtl": map[string]any{
				"days": retDays,
			},
		},
	})
}

func (h *NamespaceHandler) Create(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{})
}
