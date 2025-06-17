package models

// Edge represents a connection between nodes
type Edge struct {
	ID           string      `json:"id" db:"id"`
	WorkflowID   string      `json:"-" db:"workflow_id"`
	Source       string      `json:"source" db:"source_node_id"`
	Target       string      `json:"target" db:"target_node_id"`
	EdgeID       string      `json:"-" db:"edge_id"`
	EdgeType     string      `json:"type" db:"type"`
	Animated     bool        `json:"animated" db:"animated"`
	Style        EdgeStyle   `json:"style" db:"-"`
	Label        string      `json:"label,omitempty" db:"label"`
	SourceHandle string      `json:"sourceHandle,omitempty" db:"source_handle"`
	LabelStyle   *LabelStyle `json:"labelStyle,omitempty" db:"label_style"`
}

// EdgeStyle represents the visual style of an edge
type EdgeStyle struct {
	Stroke     string `json:"stroke"`
	StrokeWidth int    `json:"strokeWidth"`
}

// LabelStyle represents the visual style of an edge label
type LabelStyle struct {
	Fill       string `json:"fill,omitempty"`
	FontWeight string `json:"fontWeight,omitempty"`
}