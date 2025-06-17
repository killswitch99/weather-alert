package execution

import (
	"context"
	"fmt"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
	"workflow-code-test/api/pkg/node/condition"

	"github.com/google/uuid"
)

// Engine executes workflows
type Engine struct {
	registry *node.Registry
}

// NewEngine creates a workflow execution engine
func NewEngine(registry *node.Registry) *Engine {
	return &Engine{
		registry: registry,
	}
}

// Execute runs a workflow from start to finish
func (e *Engine) Execute(ctx context.Context, workflow *models.Workflow, input models.WorkflowInput) (*models.WorkflowExecution, error) {
	// Record start time
	startTime := time.Now()
	startTimeStr := startTime.Format(time.RFC3339)
	
	// Initialize workflow execution
	execution := &models.WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: workflow.ID,
		ExecutedAt: startTime,
		Status:     models.StatusRunning,
		StartTime:  startTimeStr,
		Steps:      make([]models.ExecutionStep, 0),
		Metadata:   models.JSONB{
			"workflowVersion": workflow.Version, 
			"triggeredBy":     input.Name, 
		},
	}

	// Initialize workflow routing structures
	nodes, edges, startNodeID, err := e.initializeWorkflow(workflow)
	if err != nil {
		return nil, err
	}

	// Store node outputs for access by subsequent nodes
	priorOutputs := make(map[string]node.NodeOutputs)
	nodeData := make(map[string]any) // For storing intermediate data across nodes
	
	// Execute nodes in sequence
	currentNodeID := startNodeID
	stepNumber := 1
	
	for {
		// Get and validate current node
		currentNode := nodes[currentNodeID]
		if currentNode == nil {
			return nil, fmt.Errorf("node %s not found in workflow", currentNodeID)
		}

		// Execute node
		nodeInputs := node.NodeInputs{
			WorkflowInput: input,
			NodeData:      nodeData,
			PriorOutputs:  priorOutputs,
		}
		outputs, err := currentNode.Execute(ctx, nodeInputs)
		
		// Record execution step
		step := e.createExecutionStep(currentNode, currentNodeID, outputs, workflow)
		step.StepNumber = stepNumber
		execution.Steps = append(execution.Steps, step)
		stepNumber++
		priorOutputs[currentNodeID] = outputs

		// Handle errors or failed steps
		if err != nil || outputs.Status == models.StatusFailed {
			execution.Status = models.StatusFailed
			endTime := time.Now()
			execution.EndTime = endTime.Format(time.RFC3339)
			startTime, _ := time.Parse(time.RFC3339, execution.StartTime)
			execution.TotalDuration = endTime.Sub(startTime).Milliseconds()
			return execution, nil
		}

		// Check if workflow is complete
		if currentNode.Type() == models.NodeTypeEnd {
			execution.Status = models.StatusCompleted
			endTime := time.Now()
			execution.EndTime = endTime.Format(time.RFC3339)
			startTime, _ := time.Parse(time.RFC3339, execution.StartTime)
			execution.TotalDuration = endTime.Sub(startTime).Milliseconds()
			break
		}

		// Find next node
		nextNodeID, err := e.findNextNode(currentNode, currentNodeID, outputs, edges)
		if err != nil {
			return nil, err
		}
		
		currentNodeID = nextNodeID
	}

	return execution, nil
}

// initializeWorkflow sets up all node instances and connection maps
func (e *Engine) initializeWorkflow(workflow *models.Workflow) (
	nodes map[string]node.Node,
	edges map[string]map[string]string,
	startNodeID string,
	err error) {
	
	// Create nodes
	nodes = make(map[string]node.Node)
	for _, nodeModel := range workflow.Nodes {
		n, err := e.registry.Create(nodeModel)
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to create node %s: %w", nodeModel.ID, err)
		}
		nodes[nodeModel.ID] = n
		
		// Find the start node while we're iterating
		if n.Type() == models.NodeTypeStart {
			startNodeID = nodeModel.ID
		}
	}
	
	if startNodeID == "" {
		return nil, nil, "", fmt.Errorf("no start node found in workflow")
	}
	
	// Build unified edge routing map
	// Key: sourceNodeID, Value: map[routeKey]targetNodeID
	// For regular edges, routeKey is empty string
	// For conditional edges, routeKey is "true" or "false"
	edges = make(map[string]map[string]string)
	
	for _, edge := range workflow.Edges {
		if edges[edge.Source] == nil {
			edges[edge.Source] = make(map[string]string)
		}
		
		routeKey := edge.SourceHandle // Empty for regular edges, "true"/"false" for conditional edges
		edges[edge.Source][routeKey] = edge.Target
		
		// Configure condition nodes with their routes
		if routeKey == "true" || routeKey == "false" {
			if node, ok := nodes[edge.Source]; ok && node.Type() == models.NodeTypeCondition {
				if condNode, ok := node.(*condition.Node); ok {
					if routeKey == "true" {
						condNode.SetTrueRoute(edge.Target)
					} else {
						condNode.SetFalseRoute(edge.Target)
					}
				}
			}
		}
	}
	
	return nodes, edges, startNodeID, nil
}

// createExecutionStep creates an execution step record from node outputs
func (e *Engine) createExecutionStep(
	node node.Node, 
	nodeID string, 
	outputs node.NodeOutputs,
	_ *models.Workflow) models.ExecutionStep {
	
	// Parse timestamps to calculate duration
	startTime, _ := time.Parse(time.RFC3339, outputs.StartedAt)
	endTime, _ := time.Parse(time.RFC3339, outputs.EndedAt)
	duration := endTime.Sub(startTime).Milliseconds()
	
	status := models.StatusCompleted
	if outputs.Status == models.StatusFailed {
		status = models.StatusFailed
	}
	
	// Extract error message if present
	var errorMsg string
	if err, ok := outputs.Data["error"]; ok {
		if errStr, ok := err.(string); ok {
			errorMsg = errStr
		}
	}
	
	step := models.ExecutionStep{
		NodeID:      nodeID,
		NodeType:    node.Type(),
		Status:      status,
		Duration:    duration,
		Output:      outputs.Data,
		Timestamp:   outputs.StartedAt,
		Error:       errorMsg,
		StartedAt:   outputs.StartedAt,  // Keep for internal use
		EndedAt:     outputs.EndedAt,    // Keep for internal use
	}
	
	// Use the node's current base information (may have been updated during execution)
	baseInfo := node.GetBaseInfo()
	step.Label = baseInfo.Label
	step.Description = baseInfo.Description
	
	return step
}

// findNextNode determines the next node to execute based on current node's output
func (e *Engine) findNextNode(
	currentNode node.Node, 
	currentNodeID string, 
	outputs node.NodeOutputs, 
	edges map[string]map[string]string) (string, error) {
	
	// Check if NextNodeID is explicitly set (from condition nodes)
	if outputs.NextNodeID != "" {
		return outputs.NextNodeID, nil
	}
	
	// Handle node types that use specific routing
	if currentNode.Type() == models.NodeTypeCondition {
		// Determine route based on condition result
		routeKey := ""
		if conditionMet, ok := outputs.Data["conditionMet"].(bool); ok && conditionMet {
			routeKey = "true"
		} else {
			routeKey = "false"
		}
		
		if nextNode, exists := edges[currentNodeID][routeKey]; exists {
			return nextNode, nil
		}
	}
	
	// Default to first available edge
	if nextNode, exists := edges[currentNodeID][""]; exists {
		return nextNode, nil
	}
	
	// No valid edge found
	return "", fmt.Errorf("node %s has no outgoing edges", currentNodeID)
}