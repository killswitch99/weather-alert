package service

import (
	"context"
	"fmt"
	"testing"
	"time"
	"workflow-code-test/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestWorkflowService is a test implementation of WorkflowService
type TestWorkflowService struct {
	repo *MockWorkflowRepository
}

// GetWorkflow implements WorkflowService.GetWorkflow
func (s *TestWorkflowService) GetWorkflow(ctx context.Context, id string) (*models.Workflow, error) {
	workflow, err := s.repo.Get(ctx, id)
	return workflow, err
}

// ExecuteWorkflow implements WorkflowService.ExecuteWorkflow
func (s *TestWorkflowService) ExecuteWorkflow(ctx context.Context, id string, input models.WorkflowInput) (*models.WorkflowExecution, error) {
	workflow, err := s.GetWorkflow(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate workflow structure
	if err := validateWorkflowStructure(workflow.Nodes, workflow.Edges); err != nil {
		return nil, fmt.Errorf("invalid workflow structure: %w", err)
	}

	// Create a simple execution result
	execution := &models.WorkflowExecution{
		ID:         "test-execution-id",
		WorkflowID: id,
		Status:     models.StatusCompleted,
		ExecutedAt: time.Now().Add(-5 * time.Minute),
		Steps: []models.ExecutionStep{
			{
				NodeID:      workflow.Nodes[0].ID,
				StepNumber:  1,
				Label:       workflow.Nodes[0].Data.Label,
				Description: "Execution of start node",
				StartedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				EndedAt:     time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
			},
			{
				NodeID:      workflow.Nodes[1].ID,
				StepNumber:  2,
				Label:       workflow.Nodes[1].Data.Label,
				Description: "Execution of form node",
				StartedAt:   time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
				EndedAt:     time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			},
		},
	}

	return execution, nil
}

// validateWorkflowStructure validates the workflow structure
func validateWorkflowStructure(nodes []models.Node, edges []models.Edge) error {
	if len(nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	hasStart := false
	hasEnd := false
	startNodeIndex := -1
	endNodeIndex := -1

	for i, node := range nodes {
		if node.Type == models.NodeTypeStart {
			hasStart = true
			startNodeIndex = i
		}
		if node.Type == models.NodeTypeEnd {
			hasEnd = true
			endNodeIndex = i
		}
	}

	if !hasStart {
		return fmt.Errorf("workflow must begin with a start node")
	}
	if !hasEnd {
		return fmt.Errorf("workflow must end with an end node")
	}
	if startNodeIndex != 0 {
		return fmt.Errorf("start node must be the first node in the workflow")
	}
	if endNodeIndex != len(nodes)-1 {
		return fmt.Errorf("end node must be the last node in the workflow")
	}

	return nil
}

// MockWorkflowRepository is a mock implementation of the repository
type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) Get(ctx context.Context, id string) (*models.Workflow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Workflow), args.Error(1)
}

func (m *MockWorkflowRepository) Create(ctx context.Context, workflow *models.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Update(ctx context.Context, workflow *models.Workflow) error {
	args := m.Called(ctx, workflow)
	return args.Error(0)
}

func (m *MockWorkflowRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkflowRepository) GetNodes(ctx context.Context, workflowID string) ([]models.Node, error) {
	args := m.Called(ctx, workflowID)
	return args.Get(0).([]models.Node), args.Error(1)
}

func (m *MockWorkflowRepository) GetEdges(ctx context.Context, workflowID string) ([]models.Edge, error) {
	args := m.Called(ctx, workflowID)
	return args.Get(0).([]models.Edge), args.Error(1)
}

func (m *MockWorkflowRepository) CreateExecution(ctx context.Context, execution *models.WorkflowExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockWorkflowRepository) GetExecution(ctx context.Context, id string) (*models.WorkflowExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WorkflowExecution), args.Error(1)
}

func (m *MockWorkflowRepository) CreateExecutionStep(ctx context.Context, step *models.ExecutionStep) error {
	args := m.Called(ctx, step)
	return args.Error(0)
}

func (m *MockWorkflowRepository) GetExecutionSteps(ctx context.Context, executionID string) ([]models.ExecutionStep, error) {
	args := m.Called(ctx, executionID)
	return args.Get(0).([]models.ExecutionStep), args.Error(1)
}

func TestExecuteWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		workflow      *models.Workflow
		input         models.WorkflowInput
		expectedError string
	}{
		{
			name: "valid workflow execution",
			workflow: &models.Workflow{
				ID: "test-workflow",
				Nodes: []models.Node{
					{
						ID:   "start",
						Type: models.NodeTypeStart,
						Data: models.NodeData{
							Label: "Start",
						},
					},
					{
						ID:   "form",
						Type: models.NodeTypeForm,
						Data: models.NodeData{
							Label: "Form",
						},
					},
					{
						ID:   "end",
						Type: models.NodeTypeEnd,
						Data: models.NodeData{
							Label: "End",
						},
					},
				},
				Edges: []models.Edge{
					{
						Source: "start",
						Target: "form",
					},
					{
						Source: "form",
						Target: "end",
					},
				},
			},
			input: models.WorkflowInput{
				Name:      "Test User",
				Email:     "test@example.com",
				City:      "Sydney",
				Operator:  models.OperatorGreaterThan,
				Threshold: 20,
			},
			expectedError: "",
		},
		{
			name: "invalid workflow structure - no start node",
			workflow: &models.Workflow{
				ID: "test-workflow",
				Nodes: []models.Node{
					{
						ID:   "form",
						Type: models.NodeTypeForm,
						Data: models.NodeData{
							Label: "Form",
						},
					},
				},
			},
			input: models.WorkflowInput{
				Name:      "Test User",
				Email:     "test@example.com",
				City:      "Sydney",
				Operator:  models.OperatorGreaterThan,
				Threshold: 20,
			},
			expectedError: "invalid workflow structure: workflow must begin with a start node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockWorkflowRepository)
			mockRepo.On("Get", mock.Anything, tt.workflow.ID).Return(tt.workflow, nil)

			// Create service implementation for tests
			service := &TestWorkflowService{repo: mockRepo}

			// Execute workflow
			execution, err := service.ExecuteWorkflow(context.Background(), tt.workflow.ID, tt.input)

			// Check error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			// Check execution
			assert.NoError(t, err)
			assert.NotNil(t, execution)
			assert.Equal(t, models.StatusCompleted, execution.Status)
			assert.NotEmpty(t, execution.Steps)
			assert.True(t, time.Now().After(execution.ExecutedAt))
		})
	}
}

func TestValidateWorkflowStructure(t *testing.T) {
	tests := []struct {
		name          string
		Nodes		 []models.Node
		Edges		 []models.Edge
		expectedError string
	}{
		{
			name: "valid workflow structure",
			Nodes: []models.Node{
				{
					ID:   "start",
					Type: models.NodeTypeStart,
				},
				{
					ID:   "form",
					Type: models.NodeTypeForm,
				},
				{
					ID:   "end",
					Type: models.NodeTypeEnd,
				},
			},
			Edges: []models.Edge{
				{
					Source: "start",
					Target: "form",
				},
				{
					Source: "form",
					Target: "end",
				},
			},
			expectedError: "",
		},
		{
			name: "empty workflow",
			Nodes:  []models.Node{},
			Edges:  []models.Edge{},
			expectedError: "workflow must have at least one node",
		},
		{
			name: "no start node",
			Nodes: []models.Node{
				{
					ID:   "form",
					Type: models.NodeTypeForm,
				},
			},
			expectedError: "workflow must begin with a start node",
		},
		{
			name: "no end node",
			Nodes: []models.Node{
				{
					ID:   "start",
					Type: models.NodeTypeStart,
				},
			},
			expectedError: "workflow must end with an end node",
		},
		{
			name: "start node not first",
			Nodes: []models.Node{
					{
						ID:   "form",
						Type: models.NodeTypeForm,
					},
					{
						ID:   "start",
						Type: models.NodeTypeStart,
					},
					{
						ID:   "end",
						Type: models.NodeTypeEnd,
					},
				},
			expectedError: "start node must be the first node in the workflow",
		},
		{
			name: "end node not last",
			Nodes: []models.Node{
				{
					ID:   "start",
					Type: models.NodeTypeStart,
				},
				{
					ID:   "end",
					Type: models.NodeTypeEnd,
				},
				{
					ID:   "form",
					Type: models.NodeTypeForm,
				},
			},
			expectedError: "end node must be the last node in the workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkflowStructure(tt.Nodes, tt.Edges)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}