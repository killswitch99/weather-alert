package integration

import (
	"context"
	"fmt"
	"strings"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
	"workflow-code-test/api/pkg/node/integration/weather"
)

// Node implements an integration node
type Node struct {
	node.BaseNode
	config Config
}

// Config holds integration node configuration
type Config struct {
	APIEndpoint string
	Options     []weather.WeatherOption
}

// NewNode creates an integration node from a model
func NewNode(model models.Node) (node.Node, error) {
	// Parse model.Data.Metadata into Config
	var config Config
	
	// Extract API endpoint
	apiEndpoint, ok := model.Data.Metadata["apiEndpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("missing API endpoint")
	}
	config.APIEndpoint = apiEndpoint
	
	// Extract location options
	optionsRaw, ok := model.Data.Metadata["options"].([]any)
	if ok {
		for _, opt := range optionsRaw {
			option, ok := opt.(map[string]any)
			if !ok {
				continue
			}
			
			city, _ := option["city"].(string)
			lat, _ := option["lat"].(float64)
			lon, _ := option["lon"].(float64)

			config.Options = append(config.Options, weather.WeatherOption{
				City: city,
				Lat:  lat,
				Lon:  lon,
			})
		}
	}
	
	return &Node{
		BaseNode: node.BaseNode{
			ID:          model.ID,
			Label:       model.Data.Label,
			Description: model.Data.Description,
		},
		config: config,
	}, nil
}

// Type returns the node type
func (n *Node) Type() models.NodeType {
	return models.NodeTypeIntegration
}

// GetBaseInfo returns the base node information
func (n *Node) GetBaseInfo() node.BaseNode {
	return n.BaseNode
}

// Execute implements the integration node logic
func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
	started := time.Now()
	outputs := node.NodeOutputs{
		Data:      make(map[string]any),
		Status:    models.StatusRunning,
		StartedAt: started.Format(time.RFC3339),
	}
	
	// Get city from form output
	formOutput, ok := inputs.PriorOutputs[string(models.NodeIDForm)]
	if !ok {
		outputs.Status = models.StatusFailed
		outputs.Data["error"] = "Failed to get form data"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("missing form data")
	}
	
	city, ok := formOutput.Data["city"].(string)
	if !ok {
		outputs.Status = models.StatusFailed
		outputs.Data["error"] = "Failed to get city from form output"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("missing city")
	}
	// Update the node description with the actual city name
	if strings.Contains(n.Description, "{{city}}") {
		n.Description = strings.ReplaceAll(n.Description, "{{city}}", city)
	}

	// Find location coordinates for the city
	var lat, lon float64
	found := false
	for _, option := range n.config.Options {
		if option.City == city {
			lat = option.Lat
			lon = option.Lon
			found = true
			break
		}
	}
	
	if !found {
		outputs.Status = models.StatusFailed
		outputs.Data["error"] = fmt.Sprintf("City not found: %s", city)
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("city not found: %s", city)
	}
	
	// Call the weather API using the client
	weatherClient := weather.NewClient(10 * time.Second)
	weatherData, err := weatherClient.GetWeather(ctx, n.config.APIEndpoint, lat, lon, city)
	if err != nil {
		outputs.Status = models.StatusFailed
		outputs.Data["error"] = fmt.Sprintf("Weather API error: %v", err)
		outputs.Data["message"] = "Weather API request failed"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("weather API error: %w", err)
	}
	
	temperature := weatherData.Temperature

	outputs.Status = models.StatusCompleted
	outputs.Data = map[string]any{
		"message": fmt.Sprintf("Retrieved temperature for %s: %.1f°C", city, temperature),
		"apiResponse": map[string]any{
			"endpoint": n.config.APIEndpoint,
			"method": "GET",
			"data": map[string]any{
				"temperature": temperature,
				"location": city,
			},
		},
		string(models.OutputKeyTemperature): temperature,
		string(models.OutputKeyLocation):    city,
	}
	outputs.EndedAt = time.Now().Format(time.RFC3339)
	
	return outputs, nil
}

// Validate ensures the node is properly configured
func (n *Node) Validate() error {
	if n.config.APIEndpoint == "" {
		return fmt.Errorf("missing API endpoint")
	}
	if len(n.config.Options) == 0 {
		return fmt.Errorf("no location options configured")
	}
	return nil
}
