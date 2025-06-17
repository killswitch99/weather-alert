package form

import (
	"context"
	"time"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
)

// Node implements a form node
type Node struct {
	node.BaseNode
}

// NewNode creates a form node from a model
func NewNode(model models.Node) (node.Node, error) {
	return &Node{
		BaseNode: node.BaseNode{
			ID:          model.ID,
			Label:       model.Data.Label,
			Description: model.Data.Description,
		},
	}, nil
}

// Type returns the node type
func (n *Node) Type() models.NodeType {
	return models.NodeTypeForm
}

// Execute implements the form node logic
func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
	started := time.Now()

	// Create form data matching WorkflowFormData in frontend
	formData := map[string]any{
		"name":      inputs.WorkflowInput.Name,
		"email":     inputs.WorkflowInput.Email,
		"city":      inputs.WorkflowInput.City,
		"threshold": inputs.WorkflowInput.Threshold,
		"operator":  string(inputs.WorkflowInput.Operator),
	}

	// Determine form type based on the node's label or use default
	formType := "user_input" // default form type
	if n.Label != "" {
		// Use lowercase label as form type if available
		formType = n.Label
	}

	outputs := node.NodeOutputs{
		Data: map[string]any{
			"message": "Form data processed successfully",
			"formData": formData,
			"details": map[string]any{
				"formType": formType,
				"fieldCount": len(formData),
			},
			// Keep these for backwards compatibility with existing code
			string(models.OutputKeyName):  inputs.WorkflowInput.Name,
			string(models.OutputKeyEmail): inputs.WorkflowInput.Email,
			string(models.OutputKeyCity):  inputs.WorkflowInput.City,
		},
		Status:    models.StatusCompleted,
		StartedAt: started.Format(time.RFC3339),
		EndedAt:   time.Now().Format(time.RFC3339),
	}

	return outputs, nil
}

// Validate ensures the node is properly configured
func (n *Node) Validate() error {
	return nil
}
