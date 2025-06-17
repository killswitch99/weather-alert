package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
	"workflow-code-test/api/pkg/node/integration/weather"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	// Test cases for node creation
	testCases := []struct {
		name          string
		model         models.Node
		expectedError bool
	}{
		{
			name: "Valid integration node",
			model: models.Node{
				ID:   "integration-1",
				Type: models.NodeTypeIntegration,
				Data: models.NodeData{
					Label:       "Weather API",
					Description: "Gets weather data",
					Metadata: map[string]any{
						"apiEndpoint": "https://api.example.com/weather?lat={lat}&lon={lon}",
						"options": []any{
							map[string]any{
								"city": "New York",
								"lat":  40.7128,
								"lon":  -74.0060,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Missing API endpoint",
			model: models.Node{
				ID:   "integration-2",
				Type: models.NodeTypeIntegration,
				Data: models.NodeData{
					Label:       "Weather API",
					Description: "Gets weather data",
					Metadata: map[string]any{
						"options": []any{
							map[string]any{
								"city": "New York",
								"lat":  40.7128,
								"lon":  -74.0060,
							},
						},
					},
				},
			},
			expectedError: true,
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
				assert.Equal(t, models.NodeTypeIntegration, n.Type())
				// Check if we can access the base node info
				if nodeWithBase, ok := n.(interface{ GetBaseInfo() node.BaseNode }); ok {
					assert.Equal(t, tc.model.ID, nodeWithBase.GetBaseInfo().ID)
				} else {
					t.Error("Node does not implement GetBaseInfo method")
				}
			}
		})
	}
}

func TestNodeValidate(t *testing.T) {
	// Test cases for node validation
	testCases := []struct {
		name          string
		config        Config
		expectedError bool
	}{
		{
			name: "Valid configuration",
			config: Config{
				APIEndpoint: "https://api.example.com/weather",
				Options: []weather.WeatherOption{
					{
						City: "New York",
						Lat:  40.7128,
						Lon:  -74.0060,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Missing API endpoint",
			config: Config{
				Options: []weather.WeatherOption{
					{
						City: "New York",
						Lat:  40.7128,
						Lon:  -74.0060,
					},
				},
			},
			expectedError: true,
		},
		{
			name: "No location options",
			config: Config{
				APIEndpoint: "https://api.example.com/weather",
				Options:     []weather.WeatherOption{},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := &Node{
				BaseNode: node.BaseNode{
					ID:          "integration-test",
					Label:       "Test Integration",
					Description: "Test integration node",
				},
				config: tc.config,
			}
			
			err := n.Validate()
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Create a test server to mock the weather API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for specific path parameter to return different responses
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		if r.URL.Path == "/invalid" {
			fmt.Fprintln(w, `{"not_current_weather": {}}`)
			return
		}
		
		if r.URL.Path == "/invalid_temp" {
			fmt.Fprintln(w, `{"current_weather": {"temperature": "not-a-number"}}`)
			return
		}
		
		// Default success case
		fmt.Fprintln(w, `{"current_weather": {"temperature": 20.5}}`)
	}))
	defer server.Close()

	// Test cases for execute
	testCases := []struct {
		name           string
		apiPath        string
		city           string
		cityInOptions  bool
		expectedStatus models.Status
	}{
		{
			name:           "Successful API call",
			apiPath:        "/",
			city:           "New York",
			cityInOptions:  true,
			expectedStatus: models.StatusCompleted,
		},
		{
			name:           "API error response",
			apiPath:        "/error",
			city:           "New York",
			cityInOptions:  true,
			expectedStatus: models.StatusFailed,
		},
		{
			name:           "Invalid API response format",
			apiPath:        "/invalid",
			city:           "New York",
			cityInOptions:  true,
			expectedStatus: models.StatusFailed,
		},
		{
			name:           "Invalid temperature value",
			apiPath:        "/invalid_temp",
			city:           "New York",
			cityInOptions:  true,
			expectedStatus: models.StatusFailed,
		},
		{
			name:           "City not in options",
			apiPath:        "/",
			city:           "Unknown City",
			cityInOptions:  false,
			expectedStatus: models.StatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := []weather.WeatherOption{
				{
					City: "New York",
					Lat:  40.7128,
					Lon:  -74.0060,
				},
			}
			
			n := &Node{
				BaseNode: node.BaseNode{
					ID:          "integration-test",
					Label:       "Test Integration",
					Description: "Test integration node",
				},
				config: Config{
					APIEndpoint: server.URL + tc.apiPath,
					Options:     options,
				},
			}

			// Set up inputs
			inputs := node.NodeInputs{
				PriorOutputs: map[string]node.NodeOutputs{
					string(models.NodeIDForm): {
						Data: map[string]any{
							"city": tc.city,
						},
					},
				},
			}

			// Execute the node
			outputs, err := n.Execute(context.Background(), inputs)
			
			// Assert expectations based on test case
			assert.Equal(t, tc.expectedStatus, outputs.Status)
			
			if tc.expectedStatus == models.StatusCompleted {
				assert.NoError(t, err)
				assert.Contains(t, outputs.Data, string(models.OutputKeyTemperature))
				assert.Contains(t, outputs.Data, string(models.OutputKeyLocation))
				assert.Equal(t, 20.5, outputs.Data[string(models.OutputKeyTemperature)])
				assert.Equal(t, tc.city, outputs.Data[string(models.OutputKeyLocation)])
			} else {
				if !tc.cityInOptions {
					assert.Contains(t, outputs.Data["error"], "City not found")
				}
			}
		})
	}
}

func TestExecuteMissingFormData(t *testing.T) {
	n := &Node{
		BaseNode: node.BaseNode{
			ID:          "integration-test",
			Label:       "Test Integration",
			Description: "Test integration node",
		},
		config: Config{
			APIEndpoint: "https://api.example.com/weather",
			Options: []weather.WeatherOption{
				{
					City: "New York",
					Lat:  40.7128,
					Lon:  -74.0060,
				},
			},
		},
	}

	// Test with missing form node output
	t.Run("Missing form output", func(t *testing.T) {
		inputs := node.NodeInputs{
			PriorOutputs: map[string]node.NodeOutputs{},
		}
		
		outputs, err := n.Execute(context.Background(), inputs)
		assert.Error(t, err)
		assert.Equal(t, models.StatusFailed, outputs.Status)
		assert.Contains(t, outputs.Data["error"], "Failed to get form data")
	})

	// Test with missing city in form output
	t.Run("Missing city in form output", func(t *testing.T) {
		inputs := node.NodeInputs{
			PriorOutputs: map[string]node.NodeOutputs{
				string(models.NodeIDForm): {
					Data: map[string]any{
						// City is missing
					},
				},
			},
		}
		
		outputs, err := n.Execute(context.Background(), inputs)
		assert.Error(t, err)
		assert.Equal(t, models.StatusFailed, outputs.Status)
		assert.Contains(t, outputs.Data["error"], "Failed to get city")
	})
}

func TestAPIRequestTimeout(t *testing.T) {
	// Create server that introduces delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a delay longer than the context timeout
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, `{"current_weather": {"temperature": 20.5}}`)
	}))
	defer server.Close()

	n := &Node{
		BaseNode: node.BaseNode{
			ID:          "integration-test",
			Label:       "Test Integration",
			Description: "Test integration node",
		},
		config: Config{
			APIEndpoint: server.URL,
			Options: []weather.WeatherOption{
				{
					City: "New York",
					Lat:  40.7128,
					Lon:  -74.0060,
				},
			},
		},
	}

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	inputs := node.NodeInputs{
		PriorOutputs: map[string]node.NodeOutputs{
			string(models.NodeIDForm): {
				Data: map[string]any{
					"city": "New York",
				},
			},
		},
	}

	outputs, err := n.Execute(ctx, inputs)
	assert.Error(t, err)
	assert.Equal(t, models.StatusFailed, outputs.Status)
	assert.Contains(t, outputs.Data["error"], "Weather API error")
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
