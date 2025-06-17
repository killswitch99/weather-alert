package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"workflow-code-test/api/pkg/models"
)

var (
    ErrWorkflowNotFound  = errors.New("workflow not found")
    ErrInvalidUUID       = errors.New("invalid UUID format")
    ErrExecutionNotFound = errors.New("execution not found")
)
// NodeRow represents a node row from the database.
type NodeRow struct {
    ID          string           `db:"id"`
    NodeID      string           `db:"node_id"`
    NodeType    models.NodeType  `db:"node_type"`
    PositionX   float64          `db:"position_x"`
    PositionY   float64          `db:"position_y"`
    Label       string           `db:"label"`
    Description string           `db:"description"`
    Metadata    []byte           `db:"metadata"`
}

// EdgeRow represents an edge row from the database.
type EdgeRow struct {
    ID           string  `db:"id"`
    Source       string  `db:"source_node_id"`
    Target       string  `db:"target_node_id"`
    EdgeID       string  `db:"edge_id"`
    EdgeType     string  `db:"type"`
    Animated     bool    `db:"animated"`
    StrokeColor  string  `db:"stroke_color"`
    StrokeWidth  int     `db:"stroke_width"`
    Label        string  `db:"label"`
    SourceHandle string  `db:"source_handle"`
    LabelStyle   []byte  `db:"label_style"`
}

// toModelNode converts a NodeRow to a *models.Node.
func toModelNode(row NodeRow) (*models.Node, error) {
    var metadata map[string]any
    if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal node metadata: %w", err)
    }
    return &models.Node{
        ID:      row.NodeID,
        NodeID:  row.NodeID,
        Type:    row.NodeType,
        Position: models.Position{
            X: row.PositionX,
            Y: row.PositionY,
        },
        Data: models.NodeData{
            Label:       row.Label,
            Description: row.Description,
            Metadata:    metadata,
        },
    }, nil
}

// toModelEdge converts an EdgeRow to a *models.Edge.
func toModelEdge(row EdgeRow) (*models.Edge, error) {
    var labelStyle *models.LabelStyle
    if len(row.LabelStyle) > 0 {
        if err := json.Unmarshal(row.LabelStyle, &labelStyle); err != nil {
            return nil, fmt.Errorf("failed to unmarshal label style: %w", err)
        }
    }
    edge := &models.Edge{
        ID:           row.ID,
        EdgeID:       row.EdgeID,
        Source:       row.Source,
        Target:       row.Target,
        EdgeType:     row.EdgeType,
        Animated:     row.Animated,
        Style: models.EdgeStyle{
            Stroke:     row.StrokeColor,
            StrokeWidth: row.StrokeWidth,
        },
        Label:        row.Label,
        SourceHandle: row.SourceHandle,
    }
    if labelStyle != nil && (labelStyle.Fill != "" || labelStyle.FontWeight != "") {
        edge.LabelStyle = labelStyle
    }
    return edge, nil
}
