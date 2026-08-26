package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	historyv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/history/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var statusIntToString = map[int]string{
	0: "Running", 1: "Completed", 2: "Failed", 3: "Cancelled",
	4: "TimedOut", 5: "ContinuedAsNew", 6: "Terminated",
}

var statusStringToInt = map[string]int{
	"Running": 0, "Completed": 1, "Failed": 2, "Cancelled": 3,
	"TimedOut": 4, "ContinuedAsNew": 5, "Terminated": 6,
}

var eventTypeNames = map[int]string{
	0: "Unspecified", 1: "WorkflowExecutionStarted", 2: "WorkflowTaskScheduled",
	3: "WorkflowTaskStarted", 4: "WorkflowTaskCompleted", 5: "ActivityTaskScheduled",
	6: "ActivityTaskStarted", 7: "ActivityTaskCompleted", 8: "ActivityTaskFailed",
	9: "WorkflowExecutionCompleted", 10: "WorkflowExecutionFailed",
	11: "WorkflowExecutionCancelled", 12: "WorkflowExecutionTimedOut",
	13: "WorkflowExecutionTerminated", 14: "WorkflowExecutionContinuedAsNew",
}

type WorkflowHandler struct {
	DB *pgxpool.Pool
}

func (h *WorkflowHandler) List(c echo.Context) error {
	ns := c.Param("ns")
	pageSize := 50
	if ps := c.QueryParam("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}

	offset := 0
	if token := c.QueryParam("nextPageToken"); token != "" {
		if decoded, err := base64.StdEncoding.DecodeString(token); err == nil {
			if v, err := strconv.Atoi(string(decoded)); err == nil {
				offset = v
			}
		}
	}

	query := `SELECT we.workflow_id, we.run_id, we.workflow_type, we.task_queue, we.status, we.started_at, we.closed_at,
		(SELECT count(*) FROM history_events he WHERE he.namespace_id=we.namespace_id AND he.workflow_id=we.workflow_id AND he.run_id=we.run_id) as history_length
	FROM workflow_executions we
	JOIN namespaces n ON n.id = we.namespace_id
	WHERE n.name = $1`

	args := []any{ns}
	argIdx := 2

	if status := c.QueryParam("status"); status != "" {
		if si, ok := statusStringToInt[status]; ok {
			query += fmt.Sprintf(" AND we.status = $%d", argIdx)
			args = append(args, si)
			argIdx++
		}
	}
	if wt := c.QueryParam("workflowType"); wt != "" {
		query += fmt.Sprintf(" AND we.workflow_type = $%d", argIdx)
		args = append(args, wt)
		argIdx++
	}
	if tq := c.QueryParam("taskQueue"); tq != "" {
		query += fmt.Sprintf(" AND we.task_queue = $%d", argIdx)
		args = append(args, tq)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY we.started_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pageSize+1, offset)

	rows, err := h.DB.Query(c.Request().Context(), query, args...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var executions []map[string]any
	for rows.Next() {
		var wfID, runID, wfType, taskQueue string
		var status int
		var startedAt time.Time
		var closedAt *time.Time
		var histLen int64
		if err := rows.Scan(&wfID, &runID, &wfType, &taskQueue, &status, &startedAt, &closedAt, &histLen); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		var closeTimeVal any
		if closedAt != nil {
			closeTimeVal = closedAt.Format(time.RFC3339)
		}
		executions = append(executions, map[string]any{
			"execution":     map[string]any{"workflowId": wfID, "runId": runID},
			"type":          map[string]any{"name": wfType},
			"startTime":     startedAt.Format(time.RFC3339),
			"closeTime":     closeTimeVal,
			"status":        statusIntToString[status],
			"taskQueue":     taskQueue,
			"historyLength": histLen,
		})
	}
	if executions == nil {
		executions = []map[string]any{}
	}

	nextToken := ""
	if len(executions) > pageSize {
		executions = executions[:pageSize]
		nextToken = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset + pageSize)))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"executions":    executions,
		"nextPageToken": nextToken,
	})
}

func (h *WorkflowHandler) Get(c echo.Context) error {
	ns := c.Param("ns")
	wfID := c.Param("wfId")
	runID := c.Param("runId")

	var wfType, taskQueue string
	var status int
	var startedAt time.Time
	var closedAt *time.Time

	err := h.DB.QueryRow(c.Request().Context(),
		`SELECT we.workflow_type, we.task_queue, we.status, we.started_at, we.closed_at
		FROM workflow_executions we
		JOIN namespaces n ON n.id = we.namespace_id
		WHERE n.name=$1 AND we.workflow_id=$2 AND we.run_id=$3`,
		ns, wfID, runID).Scan(&wfType, &taskQueue, &status, &startedAt, &closedAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}

	var closeTimeVal any
	if closedAt != nil {
		closeTimeVal = closedAt.Format(time.RFC3339)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workflowExecutionInfo": map[string]any{
			"execution": map[string]any{"workflowId": wfID, "runId": runID},
			"type":      map[string]any{"name": wfType},
			"startTime": startedAt.Format(time.RFC3339),
			"closeTime": closeTimeVal,
			"status":    statusIntToString[status],
			"taskQueue": taskQueue,
		},
	})
}

func (h *WorkflowHandler) History(c echo.Context) error {
	ns := c.Param("ns")
	wfID := c.Param("wfId")
	runID := c.Param("runId")

	rows, err := h.DB.Query(c.Request().Context(),
		`SELECT he.event_id, he.event_type, he.event_data, he.created_at
		FROM history_events he
		JOIN namespaces n ON n.id = he.namespace_id
		WHERE n.name=$1 AND he.workflow_id=$2 AND he.run_id=$3
		ORDER BY he.event_id`, ns, wfID, runID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var events []map[string]any
	for rows.Next() {
		var eventID int64
		var eventType int
		var eventData []byte
		var createdAt time.Time
		if err := rows.Scan(&eventID, &eventType, &eventData, &createdAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		var attrs any = map[string]any{}
		if len(eventData) > 0 {
			var ev historyv1.HistoryEvent
			if err := proto.Unmarshal(eventData, &ev); err == nil {
				b, jsonErr := protojson.MarshalOptions{EmitUnpopulated: false, UseProtoNames: true}.Marshal(&ev)
				if jsonErr == nil {
					var m map[string]any
					if parseErr := parseJSON(b, &m); parseErr == nil {
						attrs = m
					}
				}
			}
		}

		etName := eventTypeNames[eventType]
		if etName == "" {
			etName = fmt.Sprintf("Unknown(%d)", eventType)
		}

		events = append(events, map[string]any{
			"eventId":    strconv.FormatInt(eventID, 10),
			"eventTime":  createdAt.Format(time.RFC3339),
			"eventType":  etName,
			"attributes": attrs,
		})
	}
	if events == nil {
		events = []map[string]any{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"history": map[string]any{
			"events": events,
		},
	})
}

func (h *WorkflowHandler) Terminate(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{})
}

func (h *WorkflowHandler) Cancel(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{})
}

func parseJSON(b []byte, v any) error {
	return json.Unmarshal(b, v)
}
