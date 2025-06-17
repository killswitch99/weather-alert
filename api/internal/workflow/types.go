package workflow

import (
	"context"
	"errors"
	"workflow-code-test/api/internal/execution"
	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/pkg/models"
)

// Define service errors
var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrInvalidInput     = errors.New("invalid input")
)

// WorkflowServiceImpl implements the workflow.WorkflowService interface
type WorkflowServiceImpl struct {
	repo repository.WorkflowRepository
	engine *execution.Engine
}

// WorkflowService defines the interface for workflow operations
type WorkflowService interface {
	GetWorkflow(ctx context.Context, id string) (*models.Workflow, error)
	ExecuteWorkflow(ctx context.Context, id string, input models.WorkflowInput) (*models.WorkflowExecution, error)
	CreateWorkflow(ctx context.Context, workflow *models.Workflow) error
	UpdateWorkflow(ctx context.Context, workflow *models.Workflow) error
	ProcessWorkflowInput(ctx context.Context, id string, input models.WorkflowInput) (*models.Workflow, error)
	SetEngine(engine *execution.Engine)
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(repo repository.WorkflowRepository) WorkflowService {
	return &WorkflowServiceImpl{repo: repo}
}

// SetEngine sets the execution engine for the service
func (s *WorkflowServiceImpl) SetEngine(engine *execution.Engine) {
	s.engine = engine
}
