package condition

import (
	"context"
	"fmt"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
	"workflow-code-test/api/pkg/node/integration/weather"
)

// Node implements a condition node
type Node struct {
    node.BaseNode
    config Config
}

// Config holds condition node configuration
type Config struct {
    ConditionExpression string
    TrueRoute           string
    FalseRoute          string
}

// NewNode creates a condition node from a model
func NewNode(model models.Node) (node.Node, error) {
    // Parse model.Data.Metadata into Config
    config := Config{}
    
    // Extract metadata from the node model
    if metadata := model.Data.Metadata; metadata != nil {
        if expr, exists := metadata["conditionExpression"].(string); exists {
            config.ConditionExpression = expr
        }
        
        // Check for true/false handles in the metadata
        if handles, exists := metadata["hasHandles"].(map[string]any); exists {
            if sourceHandles, exists := handles["source"].([]any); exists {
                for _, handle := range sourceHandles {
                    if handle.(string) == "true" || handle.(string) == "false" {
                        // Found a conditional handle, this is just to verify the node is set up correctly
                    }
                }
            }
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
    return models.NodeTypeCondition
}

// GetBaseInfo returns the base node information
func (n *Node) GetBaseInfo() node.BaseNode {
    return n.BaseNode
}

// Execute implements the condition check logic
func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
    started := time.Now()
    outputs := node.NodeOutputs{
        Data:      make(map[string]any),
        Status:    models.StatusRunning,
        StartedAt: started.Format(time.RFC3339),
    }
    
    // Get temperature from prior integration node output
    tempNode := inputs.PriorOutputs["weather-api"]
    temperature, ok := tempNode.Data["temperature"].(float64)
    if !ok {
        outputs.Status = models.StatusFailed
        outputs.Data["error"] = "Failed to get temperature"
        outputs.EndedAt = time.Now().Format(time.RFC3339)
        return outputs, fmt.Errorf("missing temperature")
    }
    
    threshold := inputs.WorkflowInput.Threshold
    operator := inputs.WorkflowInput.Operator
    
    // Evaluate condition
    var conditionMet bool
    switch operator {
    case models.OperatorGreaterThan:
        conditionMet = temperature > threshold
    case models.OperatorLessThan:
        conditionMet = temperature < threshold
    case models.OperatorEquals:
        conditionMet = temperature == threshold
    case models.OperatorGreaterThanOrEqual:
        conditionMet = temperature >= threshold
    case models.OperatorLessThanOrEqual:
        conditionMet = temperature <= threshold
    }
    
    // Set next node based on condition
    if conditionMet {
        outputs.NextNodeID = n.config.TrueRoute
    } else {
        outputs.NextNodeID = n.config.FalseRoute
    }
    
    // Set outputs
    weatherEmoji := weather.WeatherEmoji{}
    emoji := weatherEmoji.Emoji(temperature)
    
    // Get operator symbol for display
    operatorSymbol := ">"
    switch operator {
    case models.OperatorLessThan:
        operatorSymbol = "<"
    case models.OperatorEquals:
        operatorSymbol = "="
    case models.OperatorGreaterThanOrEqual:
        operatorSymbol = "≥"
    case models.OperatorLessThanOrEqual:
        operatorSymbol = "≤"
    }

    message := fmt.Sprintf("Temperature %.1f°C %s %.1f°C %s - condition %s", 
               temperature, operatorSymbol, threshold, emoji, 
               map[bool]string{true: "met", false: "not met"}[conditionMet])
    
    // Prepare the expression for displaying in the frontend
    expression := fmt.Sprintf("temperature %s threshold", operatorSymbol)
    
    outputs.Data = map[string]any{
        "message": message,
        "conditionResult": map[string]any{
            "expression": expression,
            "result":     conditionMet,
            "temperature": temperature,
            "operator":   string(operator),
            "threshold":  threshold,
        },
        "details": map[string]any{
            "conditionType": "temperature",
            "evaluatedAt":   time.Now().Format(time.RFC3339),
        },
    }
    
    outputs.Status = models.StatusCompleted
    outputs.EndedAt = time.Now().Format(time.RFC3339)
    return outputs, nil
}

// Validate ensures the node is properly configured
func (n *Node) Validate() error {
    if n.config.TrueRoute == "" || n.config.FalseRoute == "" {
        return fmt.Errorf("condition node requires both true and false routes")
    }
    return nil
}

// SetTrueRoute sets the target node ID for when condition is true
func (n *Node) SetTrueRoute(nodeID string) {
    n.config.TrueRoute = nodeID
}

// SetFalseRoute sets the target node ID for when condition is false
func (n *Node) SetFalseRoute(nodeID string) {
    n.config.FalseRoute = nodeID
}