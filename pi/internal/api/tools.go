package api

import (
	"encoding/json"
	"net/http"

	"github.com/tempiex/pi/internal/tool"
)

func toolsHandler(registry *tool.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tools := registry.List()
		if tools == nil {
			tools = []tool.ToolInfo{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"tools": tools})
	}
}
