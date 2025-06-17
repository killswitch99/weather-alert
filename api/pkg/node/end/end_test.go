package end

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
		ID:   "end-1",
		Type: models.NodeTypeEnd,
		Data: models.NodeData{
			Label:       "Workflow End",
			Description: "End of workflow",
		},
	}

	// Create the node
	endNode, err := NewNode(model)
	
	// Assert no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, endNode)
	
	// Assert the node has correct properties
	assert.Equal(t, models.NodeTypeEnd, endNode.Type())
	
	// Check if we can access the base node info
	if baseNode, ok := endNode.(interface{ GetBaseInfo() node.BaseNode }); ok {
		baseInfo := baseNode.GetBaseInfo()
		assert.Equal(t, "end-1", baseInfo.ID)
		assert.Equal(t, "Workflow End", baseInfo.Label)
		assert.Equal(t, "End of workflow", baseInfo.Description)
	} else {
		t.Error("Node does not implement GetBaseInfo method")
	}
}

func TestExecute(t *testing.T) {
	// Create end node
	endNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "end-1",
			Label:       "Workflow End",
			Description: "End of workflow",
		},
	}
	
	// Test cases
	testCases := []struct {
		name         string
		priorOutputs map[string]node.NodeOutputs
		expectSummary bool
	}{
		{
			name: "With prior outputs",
			priorOutputs: map[string]node.NodeOutputs{
				"form": {
					Data: map[string]any{
						"name":  "John Doe",
						"email": "john@example.com",
						"city":  "New York",
					},
				},
				"weather-api": {
					Data: map[string]any{
						"temperature": 25.5,
						"location":    "New York",
					},
				},
				"condition": {
					Data: map[string]any{
						"conditionMet": true,
						"message":      "Temperature 25.5°C > 20.0°C 😎 - condition met",
					},
				},
			},
			expectSummary: true,
		},
		{
			name: "No prior outputs",
			priorOutputs: map[string]node.NodeOutputs{},
			expectSummary: false,
		},
		{
			name: "Empty prior outputs",
			priorOutputs: map[string]node.NodeOutputs{
				"form": {
					Data: map[string]any{},
				},
				"weather-api": {
					Data: map[string]any{},
				},
			},
			expectSummary: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create inputs
			inputs := node.NodeInputs{
				WorkflowInput: models.WorkflowInput{},
				NodeData:      map[string]any{},
				PriorOutputs:  tc.priorOutputs,
			}
			
			// Execute the node
			outputs, err := endNode.Execute(context.Background(), inputs)
			
			// Assert no error occurred
			assert.NoError(t, err)
			
			// Check status
			assert.Equal(t, models.StatusCompleted, outputs.Status)
			
			// Check timestamps
			_, err = time.Parse(time.RFC3339, outputs.StartedAt)
			assert.NoError(t, err, "StartedAt should be in RFC3339 format")
			_, err = time.Parse(time.RFC3339, outputs.EndedAt)
			assert.NoError(t, err, "EndedAt should be in RFC3339 format")
			
			// Verify summary presence
			if tc.expectSummary {
				assert.Contains(t, outputs.Data, "summary")
				summary, ok := outputs.Data["summary"].(map[string]any)
				assert.True(t, ok, "Summary should be a map")
				assert.NotEmpty(t, summary)
				
				// Check that keys from prior outputs are in the summary
				for nodeID := range tc.priorOutputs {
					if len(tc.priorOutputs[nodeID].Data) > 0 {
						assert.Contains(t, summary, nodeID)
					}
				}
			} else {
				_, hasSummary := outputs.Data["summary"]
				assert.False(t, hasSummary, "Should not have summary with empty prior outputs")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	endNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "end-1",
			Label:       "Workflow End",
			Description: "End of workflow",
		},
	}
	
	// End nodes should always validate successfully as they have no special requirements
	err := endNode.Validate()
	assert.NoError(t, err)
}
