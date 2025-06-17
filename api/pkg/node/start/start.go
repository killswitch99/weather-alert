package start

import (
	"context"
	"time"

	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
)

// Node implements a start node
type Node struct {
	node.BaseNode
}

// NewNode creates a start node from a model
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
	return models.NodeTypeStart
}
// GetBaseInfo returns the base node information
func (n *Node) GetBaseInfo() node.BaseNode {
	return n.BaseNode
}

// Execute implements the start node logic
func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
	started := time.Now()
	
	// Start nodes don't do much - they just start the workflow
	outputs := node.NodeOutputs{
		Data:      make(map[string]any),
		Status:    models.StatusCompleted,
		StartedAt: started.Format(time.RFC3339),
		EndedAt:   time.Now().Format(time.RFC3339),
	}
	
	return outputs, nil
}

// Validate ensures the node is properly configured
func (n *Node) Validate() error {
	// Start nodes don't have any special configuration to validate
	return nil
}