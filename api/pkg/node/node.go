package node

import (
	"context"
	"workflow-code-test/api/pkg/models"
)

// Node defines the interface that all node types must implement
type Node interface {
	// Type returns the type of this node
	Type() models.NodeType

	// Execute runs the node's logic with the given context and inputs
	Execute(ctx context.Context, inputs NodeInputs) (NodeOutputs, error)

	// Validate checks if the node configuration is valid
	Validate() error

	// GetBaseInfo returns the base information about this node
	GetBaseInfo() BaseNode
}

// BaseNode provides common node functionality
type BaseNode struct {
	ID          string
	Label       string
	Description string
}

// GetBaseInfo returns the base information about this node
func (n BaseNode) GetBaseInfo() BaseNode {
	return n
}

// NodeInputs contains all inputs available to a node during execution
type NodeInputs struct {
	WorkflowInput models.WorkflowInput
	NodeData      map[string]any
	PriorOutputs  map[string]NodeOutputs
}

// NodeOutputs represents the output of a node's execution
type NodeOutputs struct {
	Data       map[string]any
	Status     models.Status
	StartedAt  string
	EndedAt    string
	NextNodeID string // For conditional routing
}

// NodeFactory is a function that creates a node from a model
type NodeFactory func(models.Node) (Node, error)