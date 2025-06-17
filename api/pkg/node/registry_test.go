package node

import (
	"context"
	"fmt"
	"testing"
	"workflow-code-test/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

// mockNode implements the Node interface for testing
type mockNode struct {
	id          string
	nodeType    models.NodeType
	validateErr error
}

func (m *mockNode) Type() models.NodeType {
	return m.nodeType
}

func (m *mockNode) Execute(ctx context.Context, inputs NodeInputs) (NodeOutputs, error) {
	return NodeOutputs{}, nil
}

func (m *mockNode) Validate() error {
	return m.validateErr
}

func (m *mockNode) GetBaseInfo() BaseNode {
	return BaseNode{
		ID:          m.id,
		Label:       fmt.Sprintf("Mock Node %s", m.id),
		Description: "This is a mock node for testing",
	}
}

// mockFactory is a test node factory that returns a mockNode
func mockFactory(nodeType models.NodeType, validateErr error) NodeFactory {
	return func(model models.Node) (Node, error) {
		return &mockNode{
			id:          model.ID,
			nodeType:    nodeType,
			validateErr: validateErr,
		}, nil
	}
}

// errorFactory is a test node factory that returns an error
func errorFactory(err error) NodeFactory {
	return func(model models.Node) (Node, error) {
		return nil, err
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	
	// Assert registry is created with empty factories map
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.factories)
	assert.Empty(t, registry.factories)
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()
	
	// Register a factory
	startFactory := mockFactory(models.NodeTypeStart, nil)
	registry.Register(models.NodeTypeStart, startFactory)
	
	// Verify factory was registered
	assert.Len(t, registry.factories, 1)
	assert.Contains(t, registry.factories, models.NodeTypeStart)
	
	// Register another factory
	endFactory := mockFactory(models.NodeTypeEnd, nil)
	registry.Register(models.NodeTypeEnd, endFactory)
	
	// Verify both factories are registered
	assert.Len(t, registry.factories, 2)
	assert.Contains(t, registry.factories, models.NodeTypeStart)
	assert.Contains(t, registry.factories, models.NodeTypeEnd)
	
	// Test overriding a factory
	newStartFactory := mockFactory(models.NodeTypeStart, nil)
	registry.Register(models.NodeTypeStart, newStartFactory)
	
	// Verify factory count remains the same (no duplicates)
	assert.Len(t, registry.factories, 2)
}

func TestCreate(t *testing.T) {
	registry := NewRegistry()
	
	// Register some test factories
	registry.Register(models.NodeTypeStart, mockFactory(models.NodeTypeStart, nil))
	registry.Register(models.NodeTypeEnd, mockFactory(models.NodeTypeEnd, nil))
	registry.Register(models.NodeTypeForm, errorFactory(fmt.Errorf("factory error")))
	
	// Test cases
	testCases := []struct {
		name          string
		model         models.Node
		expectError   bool
		errorContains string
	}{
		{
			name: "Create start node",
			model: models.Node{
				ID:   "start-1",
				Type: models.NodeTypeStart,
			},
			expectError: false,
		},
		{
			name: "Create end node",
			model: models.Node{
				ID:   "end-1",
				Type: models.NodeTypeEnd,
			},
			expectError: false,
		},
		{
			name: "Factory returns error",
			model: models.Node{
				ID:   "form-1",
				Type: models.NodeTypeForm,
			},
			expectError:   true,
			errorContains: "factory error",
		},
		{
			name: "No factory registered",
			model: models.Node{
				ID:   "unknown-1",
				Type: models.NodeTypeEmail,
			},
			expectError:   true,
			errorContains: "no factory registered",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := registry.Create(tc.model)
			
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)
				assert.Equal(t, tc.model.Type, node.Type())
			}
		})
	}
}

func TestRegistryWithMultipleTypes(t *testing.T) {
	registry := NewRegistry()
	
	// Define all node types we want to test
	nodeTypes := []models.NodeType{
		models.NodeTypeStart,
		models.NodeTypeEnd,
		models.NodeTypeForm,
		models.NodeTypeEmail,
		models.NodeTypeCondition,
		models.NodeTypeIntegration,
	}
	
	// Register all node types with unique factories
	for _, nodeType := range nodeTypes {
		registry.Register(nodeType, mockFactory(nodeType, nil))
	}
	
	// Verify registry contains all factories
	assert.Len(t, registry.factories, len(nodeTypes))
	
	// Try creating each type of node
	for _, nodeType := range nodeTypes {
		model := models.Node{
			ID:   fmt.Sprintf("%s-1", nodeType),
			Type: nodeType,
		}
		
		node, err := registry.Create(model)
		assert.NoError(t, err)
		assert.NotNil(t, node)
		assert.Equal(t, nodeType, node.Type())
	}
}

func TestRegistryValidationPassthrough(t *testing.T) {
	registry := NewRegistry()
	
	// Create factory that returns nodes with validation errors
	validationError := fmt.Errorf("validation failed")
	registry.Register(models.NodeTypeCondition, mockFactory(models.NodeTypeCondition, validationError))
	
	// Create a node model
	model := models.Node{
		ID:   "condition-1",
		Type: models.NodeTypeCondition,
	}
	
	// Create the node
	node, err := registry.Create(model)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	
	// Validate should return the error from the mock node
	err = node.Validate()
	assert.Error(t, err)
	assert.Equal(t, validationError, err)
}
