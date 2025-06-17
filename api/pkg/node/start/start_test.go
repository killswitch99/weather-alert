package start

import (
	"context"
	"testing"
	"time"

	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	// Create a model for testing
	model := models.Node{
		ID:   "start-1",
		Type: models.NodeTypeStart,
		Data: models.NodeData{
			Label:       "Start Workflow",
			Description: "This is the beginning of the workflow",
		},
	}

	// Create the node
	startNode, err := NewNode(model)
	
	// Assert no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, startNode)
	
	// Assert the node has correct properties
	assert.Equal(t, models.NodeTypeStart, startNode.Type())
	
	// Check if we can access the base node info
	if baseNode, ok := startNode.(interface{ GetBaseInfo() node.BaseNode }); ok {
		baseInfo := baseNode.GetBaseInfo()
		assert.Equal(t, "start-1", baseInfo.ID)
		assert.Equal(t, "Start Workflow", baseInfo.Label)
		assert.Equal(t, "This is the beginning of the workflow", baseInfo.Description)
	} else {
		t.Error("Node does not implement GetBaseInfo method")
	}
}

func TestStartNodeExecute(t *testing.T) {
	startNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "start-1",
			Label:       "Start Workflow",
			Description: "This is the beginning of the workflow",
		},
	}
	
	// Create inputs
	inputs := node.NodeInputs{
		WorkflowInput: models.WorkflowInput{
			Name:  "John Doe",
			Email: "john@example.com",
			City:  "New York",
		},
		NodeData:     map[string]any{},
		PriorOutputs: map[string]node.NodeOutputs{},
	}
	
	// Execute the node
	outputs, err := startNode.Execute(context.Background(), inputs)
	
	// Assert no error occurred
	assert.NoError(t, err)
	
	// Check outputs
	assert.Equal(t, models.StatusCompleted, outputs.Status)
	assert.NotEmpty(t, outputs.StartedAt)
	assert.NotEmpty(t, outputs.EndedAt)
	
	// Just verify timestamps are present and in correct format
	_, err = time.Parse(time.RFC3339, outputs.StartedAt)
	assert.NoError(t, err, "StartedAt should be in RFC3339 format")
	_, err = time.Parse(time.RFC3339, outputs.EndedAt)
	assert.NoError(t, err, "EndedAt should be in RFC3339 format")
}

func TestStartNodeValidate(t *testing.T) {
	startNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "start-1",
			Label:       "Start Workflow",
			Description: "This is the beginning of the workflow",
		},
	}
	
	// Test validate method
	err := startNode.Validate()
	assert.NoError(t, err, "Start nodes should always validate successfully")
}
