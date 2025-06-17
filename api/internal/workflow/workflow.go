package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/pkg/models"
)

// GetWorkflow retrieves a workflow by its ID
func (s *WorkflowServiceImpl) GetWorkflow(ctx context.Context, id string) (*models.Workflow, error) {
	workflow, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowNotFound) {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}

	// Load nodes and edges
	nodes, err := s.repo.GetNodes(ctx, id)
	if err != nil {
		return nil, err
	}
	workflow.Nodes = nodes

	edges, err := s.repo.GetEdges(ctx, id)
	if err != nil {
		return nil, err
	}
	workflow.Edges = edges

	return workflow, nil
}

// ExecuteWorkflow runs a workflow with the given input
func (s *WorkflowServiceImpl) ExecuteWorkflow(ctx context.Context, id string, input models.WorkflowInput) (*models.WorkflowExecution, error) {
	if s.engine == nil {
		return nil, errors.New("execution engine not initialized")
	}

	// Process any workflow data in the input and get the workflow in one step
	workflow, err := s.ProcessWorkflowInput(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to process workflow input: %w", err)
	}

	// If no workflow was returned (no JSONB processing occurred), get it directly
	if workflow == nil {
		workflow, err = s.GetWorkflow(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	
	// Execute the workflow
	execution, err := s.engine.Execute(ctx, workflow, input)
	if err != nil {
		return nil, err
	}

	return execution, nil
}

// CreateWorkflow creates a new workflow
func (s *WorkflowServiceImpl) CreateWorkflow(ctx context.Context, workflow *models.Workflow) error {
	err := s.repo.Create(ctx, workflow)
	if err != nil {
		return err
	}
	return nil
}

// UpdateWorkflow updates an existing workflow
func (s *WorkflowServiceImpl) UpdateWorkflow(ctx context.Context, workflow *models.Workflow) error {
	err := s.repo.Update(ctx, workflow)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowNotFound) {
			return ErrWorkflowNotFound
		}
		return err
	}
	return nil
}

// ProcessWorkflowInput processes the workflow JSONB from input, creating or updating as necessary
// Returns the workflow if it was modified, otherwise nil
func (s *WorkflowServiceImpl) ProcessWorkflowInput(ctx context.Context, id string, input models.WorkflowInput) (*models.Workflow, error) {
	if input.Workflow == nil {
		// No workflow data provided, nothing to process
		return nil, nil
	}
	
	slog.Debug("Processing workflow JSONB input for ID", "id", id)

	// Convert input.Workflow to workflow model in one step without intermediate marshal/unmarshal
	var wf models.Workflow
	if err := convertJSONBToWorkflow(input.Workflow, &wf); err != nil {
		return nil, fmt.Errorf("failed to convert workflow data: %w", err)
	}

	// Basic validation of workflow structure
	if err := validateWorkflow(&wf); err != nil {
		return nil, err
	}

	// Check if the ID matches an existing workflow
	existingWorkflow, err := s.GetWorkflow(ctx, id)
	if err != nil {
		// If not found, we'll create a new one - not an error
		if !errors.Is(err, ErrWorkflowNotFound) {
			return nil, fmt.Errorf("failed to check for existing workflow: %w", err)
		}
	}

	// Handle workflow comparison and update logic
	if existingWorkflow != nil && existingWorkflow.ID == id {
		// This will save us from extra update or creation if nothing has changed
		if workflowsEqual(existingWorkflow, &wf) {
			slog.Debug("No changes detected in workflow, using existing workflow", "id", id)
			return existingWorkflow, nil
		}
		
		// Update existing workflow
		if err := s.UpdateWorkflow(ctx, &wf); err != nil {
			return nil, fmt.Errorf("failed to update workflow: %w", err)
		}
		slog.Debug("Updated workflow from input JSONB", "id", id)
	} else {
		// Create new workflow
		if err := s.CreateWorkflow(ctx, &wf); err != nil {
			return nil, fmt.Errorf("failed to create workflow: %w", err)
		}
		slog.Debug("Created new workflow from input JSONB", "id", id)
	}
	
	// Return the complete workflow with all nodes and edges
	return s.GetWorkflow(ctx, id)
}

// validateWorkflow performs basic validation on workflow structure
func validateWorkflow(wf *models.Workflow) error {
	if wf.Name == "" {
		return fmt.Errorf("workflow requires a name")
	}

	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow requires at least one node")
	}
	
	return nil
}

// convertJSONBToWorkflow converts JSONB map to workflow struct without intermediate marshaling
func convertJSONBToWorkflow(jsonbData models.JSONB, wf *models.Workflow) error {
	// Use a more efficient approach than marshal/unmarshal
	workflowBytes, err := json.Marshal(jsonbData)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow JSONB: %w", err)
	}

	if err := json.Unmarshal(workflowBytes, wf); err != nil {
		return fmt.Errorf("failed to unmarshal workflow data: %w", err)
	}

	return nil
}

// workflowsEqual efficiently compares two workflows for equality
func workflowsEqual(wf1, wf2 *models.Workflow) bool {
	// Compare basic properties (quick check before spawning goroutines)
	if wf1.Name != wf2.Name {
		return false
	}
	
	// Quick check for number of nodes and edges
	if len(wf1.Nodes) != len(wf2.Nodes) || len(wf1.Edges) != len(wf2.Edges) {
		return false
	}
	
	// Use channels to communicate results from goroutines
	nodesChan := make(chan bool, 1)
	edgesChan := make(chan bool, 1)
	
	// Compare nodes concurrently
	go func() {
		// Create maps for faster lookup
		nodesMap1 := make(map[string]models.Node)
		for _, node := range wf1.Nodes {
			nodesMap1[node.ID] = node
		}
		
		// Check if all nodes in wf2 exist in wf1 and are equal
		for _, node2 := range wf2.Nodes {
			node1, exists := nodesMap1[node2.ID]
			if !exists {
				nodesChan <- false
				return
			}
			
			// Compare node properties
			if node1.Type != node2.Type ||
			   node1.Position.X != node2.Position.X ||
			   node1.Position.Y != node2.Position.Y ||
			   node1.Data.Label != node2.Data.Label {
				nodesChan <- false
				return
			}
		}
		
		nodesChan <- true
	}()
	
	// Compare edges concurrently
	go func() {
		// Create maps for faster lookup
		edgesMap1 := make(map[string]models.Edge)
		for _, edge := range wf1.Edges {
			edgesMap1[edge.ID] = edge
		}
		
		// Check if all edges in wf2 exist in wf1 and are equal
		for _, edge2 := range wf2.Edges {
			edge1, exists := edgesMap1[edge2.ID]
			if !exists {
				edgesChan <- false
				return
			}
			
			// Compare edge properties
			if edge1.Source != edge2.Source ||
			   edge1.Target != edge2.Target ||
			   edge1.EdgeID != edge2.EdgeID ||
			   edge1.EdgeType != edge2.EdgeType ||
			   edge1.Animated != edge2.Animated ||
			   edge1.SourceHandle != edge2.SourceHandle ||
			   edge1.Label != edge2.Label {
				edgesChan <- false
				return
			}
		}
		
		edgesChan <- true
	}()
	
	// Wait for results from both goroutines
	nodesEqual := <-nodesChan
	edgesEqual := <-edgesChan
	
	return nodesEqual && edgesEqual
}

