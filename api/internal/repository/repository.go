package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"workflow-code-test/api/pkg/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkflowRepository defines the interface for workflow data operations
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *models.Workflow) error
	Get(ctx context.Context, id string) (*models.Workflow, error)
	Update(ctx context.Context, workflow *models.Workflow) error
	Delete(ctx context.Context, id string) error
	GetNodes(ctx context.Context, workflowID string) ([]models.Node, error)
	GetEdges(ctx context.Context, workflowID string) ([]models.Edge, error)
}

// WorkflowRepositoryImpl implements the WorkflowRepository interface
type WorkflowRepositoryImpl struct {
	pool *pgxpool.Pool
}

// NewWorkflowRepository creates a new repository with pgx
func NewWorkflowRepository(pool *pgxpool.Pool) WorkflowRepository {
	return &WorkflowRepositoryImpl{
		pool: pool,
	}
}

// Create creates a new workflow in the database
func (r *WorkflowRepositoryImpl) Create(ctx context.Context, workflow *models.Workflow) error {
	// Validate UUID
	if err := validateUUID(workflow.ID); err != nil {
		return fmt.Errorf("invalid workflow ID: %w", err)
	}

	// Use transaction
	return pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set initial version to 1 if not provided
		if workflow.Version == 0 {
			workflow.Version = 1
		}
		
		// Insert workflow
		err := tx.QueryRow(ctx, `
			INSERT INTO workflows (id, name, version)
			VALUES ($1, $2, $3)
			RETURNING created_at, updated_at
		`, workflow.ID, workflow.Name, workflow.Version).Scan(&workflow.CreatedAt, &workflow.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create workflow: %w", err)
		}

		// Insert nodes
		for _, node := range workflow.Nodes {
			metadataJSON, err := json.Marshal(node.Data.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal node metadata: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO workflow_nodes (
					id, workflow_id, node_id, node_type, position_x, position_y,
					label, description, metadata
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, 
				node.ID,
				workflow.ID,
				node.ID,
				node.Type,
				node.Position.X,
				node.Position.Y,
				node.Data.Label,
				node.Data.Description,
				metadataJSON,
			)
			if err != nil {
				return fmt.Errorf("failed to create node: %w", err)
			}
		}

		// Insert edges
		for _, edge := range workflow.Edges {
			labelStyleJSON, err := json.Marshal(edge.LabelStyle)
			if err != nil {
				return fmt.Errorf("failed to marshal edge label style: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO workflow_edges (
					id, workflow_id, source_node_id, target_node_id,
					edge_id, type, animated, stroke_color, stroke_width,
					label, source_handle, label_style
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`,
				edge.ID,
				workflow.ID,
				edge.Source,
				edge.Target,
				edge.EdgeID,
				edge.EdgeType,
				edge.Animated,
				edge.Style.Stroke,
				edge.Style.StrokeWidth,
				edge.Label,
				edge.SourceHandle,
				labelStyleJSON,
			)
			if err != nil {
				return fmt.Errorf("failed to create edge: %w", err)
			}
		}

		return nil
	})
}

// Get retrieves a workflow by its ID
func (r *WorkflowRepositoryImpl) Get(ctx context.Context, id string) (*models.Workflow, error) {
	if err := validateUUID(id); err != nil {
		return nil, ErrWorkflowNotFound
	}

	// Get workflow
	var workflow models.Workflow
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, version, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`, id).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Version,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Get nodes
	nodes, err := r.GetNodes(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	workflow.Nodes = nodes

	// Get edges
	edges, err := r.GetEdges(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}
	workflow.Edges = edges

	return &workflow, nil
}

// GetNodes retrieves all nodes for a workflow
func (r *WorkflowRepositoryImpl) GetNodes(ctx context.Context, workflowID string) ([]models.Node, error) {
	if err := validateUUID(workflowID); err != nil {
		return nil, fmt.Errorf("invalid workflow ID: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, node_id, node_type, position_x, position_y,
			label, description, metadata
		FROM workflow_nodes
		WHERE workflow_id = $1
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodeRows []NodeRow
	for rows.Next() {
		var nodeRow NodeRow
		
		err := rows.Scan(
			&nodeRow.ID, &nodeRow.NodeID, &nodeRow.NodeType, &nodeRow.PositionX, &nodeRow.PositionY,
			&nodeRow.Label, &nodeRow.Description, &nodeRow.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node row: %w", err)
		}
		
		nodeRows = append(nodeRows, nodeRow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating node rows: %w", err)
	}

	var processedNodes []models.Node
	for _, row := range nodeRows {
		node, err := toModelNode(row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node row: %w", err)
		}
		processedNodes = append(processedNodes, *node)
	}

	return processedNodes, nil
}

// GetEdges retrieves all edges for a workflow
func (r *WorkflowRepositoryImpl) GetEdges(ctx context.Context, workflowID string) ([]models.Edge, error) {
	if err := validateUUID(workflowID); err != nil {
		return nil, fmt.Errorf("invalid workflow ID: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, source_node_id, target_node_id,
			edge_id, type, animated, stroke_color, stroke_width,
			label, source_handle, label_style
		FROM workflow_edges
		WHERE workflow_id = $1
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer rows.Close()

	var edgeRows []EdgeRow
	for rows.Next() {
		var edgeRow EdgeRow
		err := rows.Scan(
			&edgeRow.ID, &edgeRow.Source, &edgeRow.Target, &edgeRow.EdgeID,
			&edgeRow.EdgeType, &edgeRow.Animated, &edgeRow.StrokeColor, &edgeRow.StrokeWidth,
			&edgeRow.Label, &edgeRow.SourceHandle, &edgeRow.LabelStyle,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge row: %w", err)
		}
		edgeRows = append(edgeRows, edgeRow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edge rows: %w", err)
	}

	var processedEdges []models.Edge
	for _, row := range edgeRows {
		edge, err := toModelEdge(row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert edge row: %w", err)
		}
		processedEdges = append(processedEdges, *edge)
	}

	return processedEdges, nil
}

// Update updates an existing workflow
func (r *WorkflowRepositoryImpl) Update(ctx context.Context, workflow *models.Workflow) error {
	// Validate UUID
	if err := validateUUID(workflow.ID); err != nil {
		return ErrWorkflowNotFound
	}

	return pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// First get the current version
		var currentVersion int
		err := tx.QueryRow(ctx, `
			SELECT version FROM workflows WHERE id = $1
		`, workflow.ID).Scan(&currentVersion)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrWorkflowNotFound
			}
			return fmt.Errorf("failed to get current workflow version: %w", err)
		}
		
		// Increment the version in our code
		workflow.Version = currentVersion + 1
		
		// Update workflow with new version
		row := tx.QueryRow(ctx, `
			UPDATE workflows
			SET name = $1, version = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
			RETURNING created_at, updated_at
		`, workflow.Name, workflow.Version, workflow.ID)

		err = row.Scan(&workflow.CreatedAt, &workflow.UpdatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrWorkflowNotFound
			}
			return fmt.Errorf("failed to update workflow: %w", err)
		}

		// Delete existing nodes and edges
		_, err = tx.Exec(ctx, "DELETE FROM workflow_edges WHERE workflow_id = $1", workflow.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing edges: %w", err)
		}

		_, err = tx.Exec(ctx, "DELETE FROM workflow_nodes WHERE workflow_id = $1", workflow.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing nodes: %w", err)
		}

		// Insert new nodes
		for _, node := range workflow.Nodes {
			metadataJSON, err := json.Marshal(node.Data.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal node metadata: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO workflow_nodes (
					id, workflow_id, node_id, node_type, position_x, position_y,
					label, description, metadata
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, 
				uuid.NewString(),
				workflow.ID,
				node.ID,
				node.Type,
				node.Position.X,
				node.Position.Y,
				node.Data.Label,
				node.Data.Description,
				metadataJSON,
			)
			if err != nil {
				return fmt.Errorf("failed to create node: %w", err)
			}
		}

		// Insert new edges
		for _, edge := range workflow.Edges {
			labelStyleJSON, err := json.Marshal(edge.LabelStyle)
			if err != nil {
				return fmt.Errorf("failed to marshal edge label style: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO workflow_edges (
					id, workflow_id, source_node_id, target_node_id,
					edge_id, type, animated, stroke_color, stroke_width,
					label, source_handle, label_style
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			`,
				edge.ID,
				workflow.ID,
				edge.Source,
				edge.Target,
				edge.EdgeID,
				edge.EdgeType,
				edge.Animated,
				edge.Style.Stroke,
				edge.Style.StrokeWidth,
				edge.Label,
				edge.SourceHandle,
				labelStyleJSON,
			)
			if err != nil {
				return fmt.Errorf("failed to create edge: %w", err)
			}
		}

		return nil
	})
}

// Delete deletes a workflow by its ID
func (r *WorkflowRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := validateUUID(id); err != nil {
		return fmt.Errorf("invalid workflow ID: %w", err)
	}

	commandTag, err := r.pool.Exec(ctx, `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkflowNotFound
	}

	return nil
}