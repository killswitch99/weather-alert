package models

// Node represents a node in the workflow
type Node struct {
	ID          string    `json:"id" db:"id"`
	WorkflowID  string    `json:"-" db:"workflow_id"`
	NodeID     string      `json:"-" db:"node_id"`
	Type        NodeType  `json:"type" db:"node_type"`
	Position    Position  `json:"position" db:"-"`
	Data        NodeData  `json:"data" db:"-"`
}

// Position represents the x,y coordinates of a node
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
