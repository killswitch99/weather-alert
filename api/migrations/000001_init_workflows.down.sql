-- Drop triggers first
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;
DROP TRIGGER IF EXISTS update_workflow_nodes_updated_at ON workflow_nodes;
DROP TRIGGER IF EXISTS update_workflow_edges_updated_at ON workflow_edges;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_workflow_nodes_workflow_id;
DROP INDEX IF EXISTS idx_workflow_edges_workflow_id;
DROP INDEX IF EXISTS idx_workflow_edges_source_node_id;
DROP INDEX IF EXISTS idx_workflow_edges_target_node_id;

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS workflow_edges;
DROP TABLE IF EXISTS workflow_nodes;
DROP TABLE IF EXISTS workflows;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";