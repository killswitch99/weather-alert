package models

import (
	"fmt"
	"testing"
)

// Definition represents a workflow definition for testing purposes.
type Definition struct {
	Nodes []Node
	Edges  []Edge
}

func TestWorkflowInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   WorkflowInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: WorkflowInput{
				Name:      "John Doe",
				Email:     "john@example.com",
				City:      "Sydney",
				Operator:  OperatorGreaterThan,
				Threshold: 20,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: WorkflowInput{
				Name:      "",
				Email:     "john@example.com",
				City:      "Sydney",
				Operator:  OperatorGreaterThan,
				Threshold: 20,
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: WorkflowInput{
				Name:      "John Doe",
				Email:     "invalid-email",
				City:      "Sydney",
				Operator:  OperatorGreaterThan,
				Threshold: 20,
			},
			wantErr: true,
		},
		{
			name: "empty city",
			input: WorkflowInput{
				Name:      "John Doe",
				Email:     "john@example.com",
				City:      "",
				Operator:  OperatorGreaterThan,
				Threshold: 20,
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			input: WorkflowInput{
				Name:      "John Doe",
				Email:     "john@example.com",
				City:      "Sydney",
				Operator:  "invalid_operator",
				Threshold: 20,
			},
			wantErr: true,
		},
		{
			name: "negative threshold",
			input: WorkflowInput{
				Name:      "John Doe",
				Email:     "john@example.com",
				City:      "Sydney",
				Operator:  OperatorGreaterThan,
				Threshold: -10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNodeType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		nodeType NodeType
		want     bool
	}{
		{
			name:     "valid start node",
			nodeType: NodeTypeStart,
			want:     true,
		},
		{
			name:     "valid form node",
			nodeType: NodeTypeForm,
			want:     true,
		},
		{
			name:     "valid integration node",
			nodeType: NodeTypeIntegration,
			want:     true,
		},
		{
			name:     "valid condition node",
			nodeType: NodeTypeCondition,
			want:     true,
		},
		{
			name:     "valid email node",
			nodeType: NodeTypeEmail,
			want:     true,
		},
		{
			name:     "valid end node",
			nodeType: NodeTypeEnd,
			want:     true,
		},
		{
			name:     "invalid node type",
			nodeType: "invalid_type",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nodeType.IsValid(); got != tt.want {
				t.Errorf("NodeType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{
			name:     "valid greater than",
			operator: OperatorGreaterThan,
			want:     true,
		},
		{
			name:     "valid less than",
			operator: OperatorLessThan,
			want:     true,
		},
		{
			name:     "valid equals",
			operator: OperatorEquals,
			want:     true,
		},
		{
			name:     "valid greater than or equal",
			operator: OperatorGreaterThanOrEqual,
			want:     true,
		},
		{
			name:     "valid less than or equal",
			operator: OperatorLessThanOrEqual,
			want:     true,
		},
		{
			name:     "invalid operator",
			operator: "invalid_operator",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.IsValid(); got != tt.want {
				t.Errorf("Operator.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validateWorkflowStructure(nodes []Node, edges []Edge) error {
	if len(nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	hasStart := false
	hasEnd := false
	startNodeIndex := -1
	endNodeIndex := -1

	for i, node := range nodes {
		if node.Type == NodeTypeStart {
			hasStart = true
			startNodeIndex = i
		}
		if node.Type == NodeTypeEnd {
			hasEnd = true
			endNodeIndex = i
		}
	}

	if !hasStart {
		return fmt.Errorf("workflow must begin with a start node")
	}
	if !hasEnd {
		return fmt.Errorf("workflow must end with an end node")
	}
	if startNodeIndex != 0 {
		return fmt.Errorf("start node must be the first node in the workflow")
	}
	if endNodeIndex != len(nodes)-1 {
		return fmt.Errorf("end node must be the last node in the workflow")
	}

	nodeConnections := make(map[string][]string)
	for _, edge := range edges {
		nodeConnections[edge.Source] = append(nodeConnections[edge.Source], edge.Target)
	}

	for _, node := range nodes {
		if node.Type == NodeTypeEnd {
			continue
		}
		if len(nodeConnections[node.ID]) == 0 {
			return fmt.Errorf("node %s has no outgoing connections", node.ID)
		}
	}

	for _, node := range nodes {
		if node.Type == NodeTypeStart {
			continue
		}
		hasIncoming := false
		for _, connections := range nodeConnections {
			for _, target := range connections {
				if target == node.ID {
					hasIncoming = true
					break
				}
			}
			if hasIncoming {
				break
			}
		}
		if !hasIncoming {
			return fmt.Errorf("node %s has no incoming connections", node.ID)
		}
	}

	return nil
}

func TestValidateWorkflowStructure(t *testing.T) {
	tests := []struct {
		name    string
		Nodes   []Node
		Edges   []Edge
		wantErr bool
	}{
		{
			name: "valid workflow",
			Nodes: []Node{
				{
					ID:   "start",
					Type: NodeTypeStart,
					Data: NodeData{
						Label: "Start",
					},
				},
				{
					ID:   "form",
					Type: NodeTypeForm,
					Data: NodeData{
						Label: "Form",
					},
				},
				{
					ID:   "end",
					Type: NodeTypeEnd,
					Data: NodeData{
						Label: "End",
					},
				},
			},
			Edges: []Edge{
				{
					Source: "start",
					Target: "form",
				},
				{
					Source: "form",
					Target: "end",
				},
			},
			wantErr: false,
		},
		{
			name: "empty nodes",
			Nodes:  []Node{},
			Edges:  []Edge{},
			wantErr: true,
		},
		{
			name: "missing start node",
			Nodes: []Node{
				{
					ID:   "form",
					Type: NodeTypeForm,
				},
				{
					ID:   "end",
					Type: NodeTypeEnd,
				},
			},
			Edges: []Edge{
				{
					Source: "form",
					Target: "end",
				},
			},
			wantErr: true,
		},
		{
			name: "missing end node",
			Nodes: []Node{
				{
					ID:   "start",
					Type: NodeTypeStart,
				},
				{
					ID:   "form",
					Type: NodeTypeForm,
				},
			},
			Edges: []Edge{
				{
					Source: "start",
					Target: "form",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid node type",
			Nodes: []Node{
				{
					ID:   "start",
					Type: "invalid_type",
				},
			},
			wantErr: true,
		},
		{
			name: "disconnected nodes",
			Nodes: []Node{
				{
					ID:   "start",
					Type: NodeTypeStart,
				},
				{
					ID:   "form",
					Type: NodeTypeForm,
				},
				{
					ID:   "end",
					Type: NodeTypeEnd,
				},
			},
			Edges: []Edge{
				{
					Source: "start",
					Target: "form",
				},
				// Missing Edges from form to end
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkflowStructure(tt.Nodes, tt.Edges)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}