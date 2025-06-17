package node

import (
	"fmt"
	"workflow-code-test/api/pkg/models"
)

// Registry holds all registered node types
type Registry struct {
    factories map[models.NodeType]NodeFactory
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
    return &Registry{
        factories: make(map[models.NodeType]NodeFactory),
    }
}

// Register adds a node factory for the given type
func (r *Registry) Register(nodeType models.NodeType, factory NodeFactory) {
    r.factories[nodeType] = factory
}

// Create instantiates a node from its model definition
func (r *Registry) Create(nodeModel models.Node) (Node, error) {
    factory, exists := r.factories[nodeModel.Type]
    if !exists {
        return nil, fmt.Errorf("no factory registered for node type %s", nodeModel.Type)
    }
    return factory(nodeModel)
}