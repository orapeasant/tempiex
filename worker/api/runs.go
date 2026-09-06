package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	commonv1 "github.com/tempiex/tempiex/api/gen/tempiex/api/common/v1"
	workflowservicev1 "github.com/tempiex/tempiex/api/gen/tempiex/api/workflowservice/v1"

	"github.com/tempiex/worker/run"
)

func listRunsHandler(mgr *run.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filter := run.Filter{
			Status:     run.Status(q.Get("status")),
			WorkflowID: q.Get("workflowId"),
			TaskQueue:  q.Get("taskQueue"),
		}

		runs, err := mgr.List(r.Context(), filter)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		if runs == nil {
			runs = []run.Run{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
	}
}

func getRunHandler(mgr *run.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got, err := mgr.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			httpError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, got)
	}
}

// cancelRunHandler requests cancellation of the workflow execution that
// scheduled the activity. The worker cannot cancel an activity in isolation:
// cancellation originates in the workflow and reaches the activity through its
// next heartbeat.
func cancelRunHandler(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		got, err := opts.Runs.Get(r.Context(), id)
		if err != nil {
			httpError(w, http.StatusNotFound, err)
			return
		}
		if opts.Client == nil {
			httpError(w, http.StatusServiceUnavailable, errNoTempiexClient)
			return
		}

		_, err = opts.Client.RequestCancelWorkflowExecution(r.Context(),
			&workflowservicev1.RequestCancelWorkflowExecutionRequest{
				Namespace: namespaceOr(got.Namespace, opts.Namespace),
				WorkflowExecution: &commonv1.WorkflowExecution{
					WorkflowId: got.WorkflowID,
					RunId:      got.WorkflowRunID,
				},
				Reason: "cancelled via worker API",
			})
		if err != nil {
			httpError(w, http.StatusBadGateway, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func namespaceOr(runNamespace, fallback string) string {
	if runNamespace != "" {
		return runNamespace
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func httpError(w http.ResponseWriter, status int, err error) {
	http.Error(w, err.Error(), status)
}
