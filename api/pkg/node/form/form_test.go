package form

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
		ID:   "form-1",
		Type: models.NodeTypeForm,
		Data: models.NodeData{
			Label:       "User Info Form",
			Description: "Collects user information",
		},
	}

	// Create the node
	formNode, err := NewNode(model)
	
	// Assert no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, formNode)
	
	// Assert the node has correct properties
	assert.Equal(t, models.NodeTypeForm, formNode.Type())
	
	// Check if we can access the base node info
	if baseNode, ok := formNode.(interface{ GetBaseInfo() node.BaseNode }); ok {
		baseInfo := baseNode.GetBaseInfo()
		assert.Equal(t, "form-1", baseInfo.ID)
		assert.Equal(t, "User Info Form", baseInfo.Label)
		assert.Equal(t, "Collects user information", baseInfo.Description)
	} else {
		t.Error("Node does not implement GetBaseInfo method")
	}
}

func TestExecute(t *testing.T) {
	// Create form node
	formNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "form-1",
			Label:       "User Info Form",
			Description: "Collects user information",
		},
	}
	
	// Test cases
	testCases := []struct {
		name          string
		workflowInput models.WorkflowInput
		expectedData  map[string]interface{}
	}{
		{
			name: "Complete form data",
			workflowInput: models.WorkflowInput{
				Name:  "John Doe",
				Email: "john@example.com",
				City:  "New York",
			},
			expectedData: map[string]interface{}{
				string(models.OutputKeyName):  "John Doe",
				string(models.OutputKeyEmail): "john@example.com",
				string(models.OutputKeyCity):  "New York",
			},
		},
		{
			name: "Partial form data",
			workflowInput: models.WorkflowInput{
				Name: "Jane Doe",
				City: "London",
			},
			expectedData: map[string]interface{}{
				string(models.OutputKeyName):  "Jane Doe",
				string(models.OutputKeyEmail): "",
				string(models.OutputKeyCity):  "London",
			},
		},
		{
			name:          "Empty form data",
			workflowInput: models.WorkflowInput{},
			expectedData: map[string]interface{}{
				string(models.OutputKeyName):  "",
				string(models.OutputKeyEmail): "",
				string(models.OutputKeyCity):  "",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create inputs
			inputs := node.NodeInputs{
				WorkflowInput: tc.workflowInput,
				NodeData:      map[string]interface{}{},
				PriorOutputs:  map[string]node.NodeOutputs{},
			}
			
			// Execute the node
			outputs, err := formNode.Execute(context.Background(), inputs)
			
			// Assert no error occurred
			assert.NoError(t, err)
			
			// Check status
			assert.Equal(t, models.StatusCompleted, outputs.Status)
			
			// Check timestamps
			_, err = time.Parse(time.RFC3339, outputs.StartedAt)
			assert.NoError(t, err, "StartedAt should be in RFC3339 format")
			_, err = time.Parse(time.RFC3339, outputs.EndedAt)
			assert.NoError(t, err, "EndedAt should be in RFC3339 format")
			
			// Check data matches expected values
			for key, expectedValue := range tc.expectedData {
				assert.Equal(t, expectedValue, outputs.Data[key], "Data field %s should match", key)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	formNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "form-1",
			Label:       "User Info Form",
			Description: "Collects user information",
		},
	}
	
	// Form nodes should always validate successfully as they have no special requirements
	err := formNode.Validate()
	assert.NoError(t, err)
}
