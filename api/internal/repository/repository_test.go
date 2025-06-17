package repository

import (
	"context"
	"testing"
	"workflow-code-test/api/pkg/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func setupTestPgxDB(t *testing.T) *pgxpool.Pool {
	// Use PostgreSQL for testing
	pool, err := pgxpool.New(context.Background(), "postgres://workflow:workflow123@localhost:5876/workflow_engine?sslmode=disable")
	assert.NoError(t, err)

	// Create the workflows table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS workflows (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	assert.NoError(t, err)

	// Create the workflow_nodes table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS workflow_nodes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			node_id VARCHAR(50) NOT NULL,
			node_type VARCHAR(50) NOT NULL,
			position_x FLOAT NOT NULL,
			position_y FLOAT NOT NULL,
			label VARCHAR(255) NOT NULL,
			description TEXT,
			metadata JSONB
		)
	`)
	assert.NoError(t, err)

	// Create the workflow_edges table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS workflow_edges (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			source_node_id VARCHAR(50) NOT NULL,
			target_node_id VARCHAR(50) NOT NULL,
			edge_id VARCHAR(50) NOT NULL,
			type VARCHAR(50) NOT NULL,
			animated BOOLEAN NOT NULL DEFAULT false,
			stroke_color VARCHAR(50) NOT NULL DEFAULT '#000000',
			stroke_width INTEGER NOT NULL DEFAULT 1,
			label VARCHAR(255) NOT NULL DEFAULT '',
			source_handle VARCHAR(50) NOT NULL DEFAULT '',
			label_style JSONB NOT NULL DEFAULT '{}'
		)
	`)
	assert.NoError(t, err)

	return pool
}

func TestWorkflowRepository_Create(t *testing.T) {
	pool := setupTestPgxDB(t)
	defer pool.Close()

	repo := NewWorkflowRepository(pool)
	ctx := context.Background()

	workflowID := uuid.New()
	workflow := &models.Workflow{
		ID:   workflowID.String(),
		Name: "Test Workflow",
	}

	// Test Create
	err := repo.Create(ctx, workflow)
	assert.NoError(t, err)
	assert.NotEmpty(t, workflow.ID)
	assert.NotEmpty(t, workflow.CreatedAt)
	assert.NotEmpty(t, workflow.UpdatedAt)

	// Verify the workflow was created
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM workflows WHERE id = $1", workflow.ID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWorkflowRepository_Get(t *testing.T) {
	pool := setupTestPgxDB(t)
	defer pool.Close()

	repo := NewWorkflowRepository(pool)
	ctx := context.Background()

	// Create a workflow
	workflowID := uuid.New().String()
	node1ID := uuid.New().String()
	edge1ID := uuid.New().String()
	
	workflow := &models.Workflow{
		ID:   workflowID,
		Name: "Test Workflow for Get",
		Nodes: []models.Node{
			{
				ID:     node1ID,
				NodeID: "node1",
				Type:   models.NodeTypeStart,
				Position: models.Position{
					X: 100,
					Y: 100,
				},
				Data: models.NodeData{
					Label:       "Start Node",
					Description: "This is a start node",
					Metadata:    map[string]interface{}{"key": "value"},
				},
			},
		},
		Edges: []models.Edge{
			{
				ID:          edge1ID,
				Source:      "node1",
				Target:      "node2",
				EdgeID:      "edge1",
				EdgeType:    "default",
				Animated:    true,
				SourceHandle: "handle1",
				Label:       "Test Edge",
				Style: models.EdgeStyle{
					Stroke:      "#ff0000",
					StrokeWidth: 2,
				},
				LabelStyle: &models.LabelStyle{},
			},
		},
	}

	err := repo.Create(ctx, workflow)
	assert.NoError(t, err)

	// Test Get
	fetchedWorkflow, err := repo.Get(ctx, workflowID)
	assert.NoError(t, err)
	assert.Equal(t, workflowID, fetchedWorkflow.ID)
	assert.Equal(t, workflow.Name, fetchedWorkflow.Name)
	assert.NotEmpty(t, fetchedWorkflow.CreatedAt)
	assert.NotEmpty(t, fetchedWorkflow.UpdatedAt)

	// Verify nodes were retrieved
	assert.Len(t, fetchedWorkflow.Nodes, 1)

	// Verify edges were retrieved
	assert.Len(t, fetchedWorkflow.Edges, 1)
	assert.Equal(t, "node1", fetchedWorkflow.Edges[0].Source)
	assert.Equal(t, "node2", fetchedWorkflow.Edges[0].Target)
	assert.Equal(t, "edge1", fetchedWorkflow.Edges[0].EdgeID)
}

func TestWorkflowRepositoryImpl_Update(t *testing.T) {
	pool := setupTestPgxDB(t)
	defer pool.Close()

	repo := NewWorkflowRepository(pool)
	ctx := context.Background()

	// Create a workflow
	workflowID := uuid.New().String()
	workflow := &models.Workflow{
		ID:   workflowID,
		Name: "Test Workflow for Update",
	}

	err := repo.Create(ctx, workflow)
	assert.NoError(t, err)

	// Update workflow
	workflow.Name = "Updated Workflow"
	updatedNodeID := uuid.New().String()
	workflow.Nodes = []models.Node{
		{
			ID:     updatedNodeID,
			NodeID: "updated-node",
			Type:   models.NodeTypeForm,
			Position: models.Position{
				X: 200,
				Y: 300,
			},
			Data: models.NodeData{
				Label:       "Updated Node",
				Description: "This is an updated node",
				Metadata:    map[string]interface{}{"updated": true},
			},
		},
	}

	err = repo.Update(ctx, workflow)
	assert.NoError(t, err)

	// Verify the workflow was updated
	fetchedWorkflow, err := repo.Get(ctx, workflowID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Workflow", fetchedWorkflow.Name)
	assert.Len(t, fetchedWorkflow.Nodes, 1)
}

func TestWorkflowRepositoryImpl_Delete(t *testing.T) {
	pool := setupTestPgxDB(t)
	defer pool.Close()

	repo := NewWorkflowRepository(pool)
	ctx := context.Background()

	// Create a workflow
	workflowID := uuid.New().String()
	workflow := &models.Workflow{
		ID:   workflowID,
		Name: "Test Workflow for Delete",
	}

	err := repo.Create(ctx, workflow)
	assert.NoError(t, err)

	// Delete the workflow
	err = repo.Delete(ctx, workflowID)
	assert.NoError(t, err)

	// Verify the workflow was deleted
	_, err = repo.Get(ctx, workflowID)
	assert.Error(t, err)
	assert.Equal(t, ErrWorkflowNotFound, err)
}
