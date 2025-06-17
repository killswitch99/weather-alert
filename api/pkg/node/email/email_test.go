package email

import (
	"context"
	"testing"
	"time"
	"workflow-code-test/api/pkg/mailer"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	// Create a model for testing
	model := models.Node{
		ID:   "email-1",
		Type: models.NodeTypeEmail,
		Data: models.NodeData{
			Label:       "Send Weather Alert",
			Description: "Sends an email alert about the weather",
		},
	}

	// Create the node
	emailNode, err := NewNode(model)
	
	// Assert no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, emailNode)
	
	// Assert the node has correct properties
	assert.Equal(t, models.NodeTypeEmail, emailNode.Type())
	
	// Check if we can access the base node info
	if baseNode, ok := emailNode.(interface{ GetBaseInfo() node.BaseNode }); ok {
		baseInfo := baseNode.GetBaseInfo()
		assert.Equal(t, "email-1", baseInfo.ID)
		assert.Equal(t, "Send Weather Alert", baseInfo.Label)
		assert.Equal(t, "Sends an email alert about the weather", baseInfo.Description)
	} else {
		t.Error("Node does not implement GetBaseInfo method")
	}
}

func TestExecute(t *testing.T) {
	// Create email node with email template
	emailNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "email-1",
			Label:       "Send Alert",
			Description: "Email weather alert notification",
		},
		InputVariables: []string{"city", "temperature"},
		EmailTemplate: mailer.EmailTemplate{
			Subject: "Weather Alert",
			Body:    "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
		},
	}
	
	// Test cases
	testCases := []struct {
		name           string
		conditionMet   bool
		formData       map[string]any
		weatherData    map[string]any
		expectedOutput map[string]any
		expectEmail    bool
	}{
		{
			name:         "Condition Met - Send Email",
			conditionMet: true,
			formData: map[string]any{
				"email": "atopu95@gmail.com",
				"name":  "John Doe",
				"city":  "Sydney",
			},
			weatherData: map[string]any{
				"temperature": 6.1,
				"location":    "Sydney",
			},
			expectedOutput: map[string]any{
				"message": "Email sent successfully",
				"details": map[string]any{
					"outputVariables": []string{"emailSent"},
				},
				"emailContent": map[string]any{
					"to":        "atopu95@gmail.com",
					"subject":   "Weather Alert",
					"body":      "Weather alert for Sydney! Temperature is 6.1°C!",
				},
			},
			expectEmail: true,
		},
		{
			name:         "Condition Not Met - Don't Send Email",
			conditionMet: false,
			formData:     map[string]any{},
			weatherData:  map[string]any{},
			expectedOutput: map[string]any{
				"message": "Email not sent - condition not met",
				"details": map[string]any{
					"reason": "Condition not met",
				},
			},
			expectEmail: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create inputs
			inputs := node.NodeInputs{
				PriorOutputs: map[string]node.NodeOutputs{
					string(models.NodeIDCondition): {
						Data: map[string]any{
							"message": "Temperature condition check",
							"conditionResult": map[string]any{
								"expression": "temperature < threshold",
								"result": tc.conditionMet,
								"temperature": 6.1,
								"operator": "less_than",
								"threshold": 10.0,
							},
						},
					},
				},
			}
			
			// Add form and weather data if condition is met
			if tc.conditionMet {
				inputs.PriorOutputs[string(models.NodeIDForm)] = node.NodeOutputs{
					Data: tc.formData,
				}
				inputs.PriorOutputs[string(models.NodeIDWeatherAPI)] = node.NodeOutputs{
					Data: tc.weatherData,
				}
			}
			
			// Execute the node
			outputs, err := emailNode.Execute(context.Background(), inputs)
			
			// Verify response
			assert.NoError(t, err)
			
			// Check timestamps
			_, err = time.Parse(time.RFC3339, outputs.StartedAt)
			assert.NoError(t, err, "StartedAt should be in RFC3339 format")
			_, err = time.Parse(time.RFC3339, outputs.EndedAt)
			assert.NoError(t, err, "EndedAt should be in RFC3339 format")
			
			// Check the expected output data
			if tc.expectEmail {
				// Check message
				assert.Equal(t, tc.expectedOutput["message"], outputs.Data["message"])
				
				// Check email content
				expectedContent := tc.expectedOutput["emailContent"].(map[string]any)
				emailContent, ok := outputs.Data["emailContent"].(map[string]any)
				assert.True(t, ok, "Should have emailContent")
				assert.Equal(t, expectedContent["to"], emailContent["to"])
				assert.Equal(t, expectedContent["subject"], emailContent["subject"])
				assert.Equal(t, expectedContent["body"], emailContent["body"])
				assert.NotEmpty(t, emailContent["timestamp"])
				
				// Check details
				expectedDetails := tc.expectedOutput["details"].(map[string]any)
				details, ok := outputs.Data["details"].(map[string]any)
				assert.True(t, ok, "Should have details")
				assert.Equal(t, expectedDetails["outputVariables"], details["outputVariables"])
			} else {
				// Verify message for condition not met
				assert.Equal(t, tc.expectedOutput["message"], outputs.Data["message"])
				
				// Check details
				expectedDetails := tc.expectedOutput["details"].(map[string]any)
				details, ok := outputs.Data["details"].(map[string]any)
				assert.True(t, ok, "Should have details")
				assert.Equal(t, expectedDetails["reason"], details["reason"])
			}
		})
	}
}

func TestExecuteErrors(t *testing.T) {
	// Create email node with required input variables
	emailNode := &Node{
		BaseNode: node.BaseNode{
			ID:          "email-1",
			Label:       "Send Alert",
			Description: "Email weather alert notification",
		},
		InputVariables: []string{"city", "temperature"},
		EmailTemplate: mailer.EmailTemplate{
			Subject: "Weather Alert",
			Body:    "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
		},
	}
	
	// Test cases for error scenarios
	testCases := []struct {
		name           string
		priorOutputs   map[string]node.NodeOutputs
		expectedError  string
		expectedStatus models.Status
	}{
		{
			name: "Missing Condition Output",
			priorOutputs: map[string]node.NodeOutputs{
				// No condition output
			},
			expectedError:  "failed to get condition result",
			expectedStatus: models.StatusFailed,
		},
		{
			name: "Invalid Condition Output Format",
			priorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDCondition): {
					Data: map[string]any{
						// Missing conditionResult field
						"message": "Temperature condition check",
					},
				},
			},
			expectedError:  "invalid condition result format",
			expectedStatus: models.StatusFailed,
		},
		{
			name: "Missing Form Data",
			priorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDCondition): {
					Data: map[string]any{
						"conditionResult": map[string]any{
							"expression": "temperature < threshold",
							"result": true,
							"temperature": 6.1,
							"operator": "less_than",
							"threshold": 10.0,
						},
					},
				},
				// No form output
			},
			expectedError:  "missing form data",
			expectedStatus: models.StatusFailed,
		},
		{
			name: "Missing Email Field",
			priorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDCondition): {
					Data: map[string]any{
						"conditionResult": map[string]any{
							"expression": "temperature < threshold",
							"result": true,
							"temperature": 6.1,
							"operator": "less_than",
							"threshold": 10.0,
						},
					},
				},
				string(models.NodeIDForm): {
					Data: map[string]any{
						// Missing email
						"name": "John Doe",
						"city": "Sydney",
					},
				},
				string(models.NodeIDWeatherAPI): {
					Data: map[string]any{
						"temperature": 6.1,
						"location":    "Sydney",
					},
				},
			},
			expectedError:  "missing email",
			expectedStatus: models.StatusFailed,
		},
		{
			name: "Missing Required Variable - City",
			priorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDCondition): {
					Data: map[string]any{
						"conditionResult": map[string]any{
							"expression": "temperature < threshold",
							"result": true,
							"temperature": 6.1,
							"operator": "less_than",
							"threshold": 10.0,
						},
					},
				},
				string(models.NodeIDForm): {
					Data: map[string]any{
						"email": "atopu95@gmail.com",
						"name":  "John Doe",
						// Missing city
					},
				},
				string(models.NodeIDWeatherAPI): {
					Data: map[string]any{
						"temperature": 6.1,
					},
				},
			},
			expectedError:  "missing required variable: city",
			expectedStatus: models.StatusFailed,
		},
		{
			name: "Missing Required Variable - Temperature",
			priorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDCondition): {
					Data: map[string]any{
						"conditionResult": map[string]any{
							"expression": "temperature < threshold",
							"result": true,
							"temperature": 6.1,
							"operator": "less_than",
							"threshold": 10.0,
						},
					},
				},
				string(models.NodeIDForm): {
					Data: map[string]any{
						"email": "atopu95@gmail.com",
						"name":  "John Doe",
						"city":  "Sydney",
					},
				},
				string(models.NodeIDWeatherAPI): {
					Data: map[string]any{
						// Missing temperature
						"location": "Sydney",
					},
				},
			},
			expectedError:  "missing required variable: temperature",
			expectedStatus: models.StatusFailed,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create inputs
			inputs := node.NodeInputs{
				PriorOutputs: tc.priorOutputs,
			}
			
			// Execute the node
			outputs, err := emailNode.Execute(context.Background(), inputs)
			
			// Verify error response
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
			assert.Equal(t, tc.expectedStatus, outputs.Status)
			assert.Contains(t, outputs.Data, "error")
			assert.Contains(t, outputs.Data, "message")
			assert.Equal(t, "Failed to process email", outputs.Data["message"])
		})
	}
}

func TestValidate(t *testing.T) {
	t.Run("Valid Configuration", func(t *testing.T) {
		emailNode := &Node{
			BaseNode: node.BaseNode{
				ID:          "email-1",
				Label:       "Send Alert",
				Description: "Email weather alert notification",
			},
			InputVariables: []string{"city", "temperature"},
			EmailTemplate: mailer.EmailTemplate{
				Subject: "Weather Alert",
				Body:    "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
			},
		}
		
		err := emailNode.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("Missing Input Variables", func(t *testing.T) {
		emailNode := &Node{
			BaseNode: node.BaseNode{
				ID:          "email-1",
				Label:       "Send Alert",
				Description: "Email weather alert notification",
			},
			InputVariables: []string{}, // No input variables
			EmailTemplate: mailer.EmailTemplate{
				Subject: "Weather Alert",
				Body:    "Weather alert! Temperature is too low!",
			},
		}
		
		err := emailNode.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email node requires at least one input variable")
	})
	
	t.Run("Missing Email Template", func(t *testing.T) {
		emailNode := &Node{
			BaseNode: node.BaseNode{
				ID:          "email-1",
				Label:       "Send Alert",
				Description: "Email weather alert notification",
			},
			InputVariables: []string{"city", "temperature"},
			EmailTemplate: mailer.EmailTemplate{
				// Missing subject
				Body: "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
			},
		}
		
		err := emailNode.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email node requires both subject and body templates")
	})
}
