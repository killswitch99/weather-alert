package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"workflow-code-test/api/internal/workflow"
	"workflow-code-test/api/pkg/models"

	"github.com/gorilla/mux"
)

type WorkflowHandler struct {
	Service workflow.WorkflowService
}

func NewWorkflowHandler(service workflow.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		Service: service,
	}
}

func (h *WorkflowHandler) HandleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Returning workflow definition for id", "id", id)

	workflowObj, err := h.Service.GetWorkflow(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get workflow", "error", err)
		if errors.Is(err, workflow.ErrWorkflowNotFound) {
			http.Error(w, "Workflow not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(workflowObj)
}

func (h *WorkflowHandler) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Handling workflow execution for id", "id", id)

	var input models.WorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the input
	if err := input.Validate(); err != nil {
		slog.Error("Invalid input", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	execution, err := h.Service.ExecuteWorkflow(r.Context(), id, input)
	if err != nil {
		slog.Error("Failed to execute workflow", "error", err)
		if errors.Is(err, workflow.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, workflow.ErrWorkflowNotFound) {
			http.Error(w, "Workflow not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to execute workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(execution)
}
