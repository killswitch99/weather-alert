package condition

import (
	"context"
	"testing"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		model         models.Node
		expectedError bool
	}{
		{
			name: "Valid condition node",
			model: models.Node{
				ID:   "condition-1",
				Type: models.NodeTypeCondition,
				Data: models.NodeData{
					Label:       "Temperature Check",
					Description: "Check if temperature meets threshold",
					Metadata: map[string]any{
						"conditionExpression": "temperature > threshold",
						"hasHandles": map[string]any{
							"source": []any{"true", "false"},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "No metadata",
			model: models.Node{
				ID:   "condition-2",
				Type: models.NodeTypeCondition,
				Data: models.NodeData{
					Label:       "Temperature Check",
					Description: "Check if temperature meets threshold",
				},
			},
			expectedError: false, // Should not error, just create with empty config
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := NewNode(tc.model)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, n)
				assert.Equal(t, models.NodeTypeCondition, n.Type())
				
				// Check base info
				if baseNode, ok := n.(interface{ GetBaseInfo() node.BaseNode }); ok {
					baseInfo := baseNode.GetBaseInfo()
					assert.Equal(t, tc.model.ID, baseInfo.ID)
					assert.Equal(t, tc.model.Data.Label, baseInfo.Label)
					assert.Equal(t, tc.model.Data.Description, baseInfo.Description)
				} else {
					t.Error("Node does not implement GetBaseInfo method")
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Test cases for execute
	testCases := []struct {
		name            string
		temperature     float64
		threshold       float64
		operator        models.Operator
		expectedRoute   string
		conditionMet    bool
		trueRoute       string
		falseRoute      string
		operatorSymbol  string
	}{
		{
			name:           "Greater Than - Condition Met",
			temperature:    25.5,
			threshold:      20.0,
			operator:       models.OperatorGreaterThan,
			expectedRoute:  "email-node",
			conditionMet:   true,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: ">",
		},
		{
			name:           "Greater Than - Condition Not Met",
			temperature:    15.5,
			threshold:      20.0,
			operator:       models.OperatorGreaterThan,
			expectedRoute:  "end-node",
			conditionMet:   false,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: ">",
		},
		{
			name:           "Less Than - Condition Met",
			temperature:    15.5,
			threshold:      20.0,
			operator:       models.OperatorLessThan,
			expectedRoute:  "email-node",
			conditionMet:   true,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: "<",
		},
		{
			name:           "Equals - Condition Met",
			temperature:    20.0,
			threshold:      20.0,
			operator:       models.OperatorEquals,
			expectedRoute:  "email-node",
			conditionMet:   true,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: "=",
		},
		{
			name:           "Greater Than Or Equal - Condition Met",
			temperature:    20.0,
			threshold:      20.0,
			operator:       models.OperatorGreaterThanOrEqual,
			expectedRoute:  "email-node",
			conditionMet:   true,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: "≥",
		},
		{
			name:           "Less Than Or Equal - Condition Met",
			temperature:    15.5,
			threshold:      20.0,
			operator:       models.OperatorLessThanOrEqual,
			expectedRoute:  "email-node",
			conditionMet:   true,
			trueRoute:      "email-node",
			falseRoute:     "end-node",
			operatorSymbol: "≤",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create condition node with routes
			conditionNode := &Node{
				BaseNode: node.BaseNode{
					ID:          "condition-1",
					Label:       "Temperature Check",
					Description: "Check if temperature meets threshold",
				},
				config: Config{
					ConditionExpression: "temperature > threshold",
					TrueRoute:           tc.trueRoute,
					FalseRoute:          tc.falseRoute,
				},
			}
			
			// Setup inputs with weather API data and workflow input
			inputs := node.NodeInputs{
				WorkflowInput: models.WorkflowInput{
					Threshold: tc.threshold,
					Operator:  tc.operator,
				},
				PriorOutputs: map[string]node.NodeOutputs{
					"weather-api": {
						Data: map[string]any{
							"temperature": tc.temperature,
						},
					},
				},
			}
			
			// Execute the node
			outputs, err := conditionNode.Execute(context.Background(), inputs)
			
			// Verify no error
			assert.NoError(t, err)
			assert.Equal(t, models.StatusCompleted, outputs.Status)
			
			// Verify timestamps exist and are properly formatted
			_, err = time.Parse(time.RFC3339, outputs.StartedAt)
			assert.NoError(t, err, "StartedAt should be in RFC3339 format")
			_, err = time.Parse(time.RFC3339, outputs.EndedAt)
			assert.NoError(t, err, "EndedAt should be in RFC3339 format")
			
			// Check message field
			assert.Contains(t, outputs.Data["message"], tc.operatorSymbol)
			
			// Check conditionResult structure
			conditionResult, ok := outputs.Data["conditionResult"].(map[string]any)
			assert.True(t, ok, "conditionResult should be a map")
			
			// Verify condition results in the new structure
			assert.Equal(t, tc.conditionMet, conditionResult["result"])
			assert.Equal(t, tc.temperature, conditionResult["temperature"])
			assert.Equal(t, tc.threshold, conditionResult["threshold"])
			assert.Equal(t, string(tc.operator), conditionResult["operator"])
			assert.Contains(t, conditionResult["expression"], tc.operatorSymbol)
			
			// Verify next node routing
			assert.Equal(t, tc.expectedRoute, outputs.NextNodeID)
		})
	}
}

func TestExecuteWithMissingTemperature(t *testing.T) {
	// Create condition node
	conditionNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "condition-1",
			Label:       "Temperature Check",
			Description: "Check if temperature meets threshold",
		},
		config: Config{
			TrueRoute:  "email-node",
			FalseRoute: "end-node",
		},
	}
	
	// Test with missing temperature data
	inputs := node.NodeInputs{
		WorkflowInput: models.WorkflowInput{
			Threshold: 20.0,
			Operator:  models.OperatorGreaterThan,
		},
		PriorOutputs: map[string]node.NodeOutputs{
			"weather-api": {
				Data: map[string]any{
					// Missing temperature
				},
			},
		},
	}
	
	// Execute the node
	outputs, err := conditionNode.Execute(context.Background(), inputs)
	
	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing temperature")
	assert.Equal(t, models.StatusFailed, outputs.Status)
	assert.Contains(t, outputs.Data, "error")
	
	// In the new structure, there should be no conditionResult field
	_, ok := outputs.Data["conditionResult"]
	assert.False(t, ok, "conditionResult should not be present when there's an error")
}

func TestValidate(t *testing.T) {
	// Test cases for validation
	testCases := []struct {
		name          string
		config        Config
		expectedError bool
	}{
		{
			name: "Valid config",
			config: Config{
				ConditionExpression: "temperature > threshold",
				TrueRoute:           "email-node",
				FalseRoute:          "end-node",
			},
			expectedError: false,
		},
		{
			name: "Missing TrueRoute",
			config: Config{
				ConditionExpression: "temperature > threshold",
				FalseRoute:          "end-node",
			},
			expectedError: true,
		},
		{
			name: "Missing FalseRoute",
			config: Config{
				ConditionExpression: "temperature > threshold",
				TrueRoute:           "email-node",
			},
			expectedError: true,
		},
		{
			name: "Missing Both Routes",
			config: Config{
				ConditionExpression: "temperature > threshold",
			},
			expectedError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := &Node{
				BaseNode: node.BaseNode{
					ID:          "condition-1",
					Label:       "Temperature Check",
					Description: "Check if temperature meets threshold",
				},
				config: tc.config,
			}
			
			err := node.Validate()
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requires both true and false routes")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetRoutes(t *testing.T) {
	// Create condition node
	node := &Node{
		BaseNode: node.BaseNode{
			ID:          "condition-1",
			Label:       "Temperature Check",
			Description: "Check if temperature meets threshold",
		},
		config: Config{},
	}
	
	// Initially validate should fail
	assert.Error(t, node.Validate(), "Node should not validate before routes are set")
	
	// Set the routes
	node.SetTrueRoute("email-node")
	node.SetFalseRoute("end-node")
	
	// Now validate should succeed
	assert.NoError(t, node.Validate(), "Node should validate after routes are set")
	assert.Equal(t, "email-node", node.config.TrueRoute)
	assert.Equal(t, "end-node", node.config.FalseRoute)
}
