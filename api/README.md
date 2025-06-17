# ⚡ Workflow API

A Go-based API for managing and executing workflow automations. Provides endpoints to retrieve workflow definitions and execute workflows, with PostgreSQL for persistent storage.

## 🛠️ Tech Stack

- Go 1.23+
- PostgreSQL
- Docker (for development and deployment)

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL
- Docker & Docker Compose (recommended for development)

### 1. Configure Database

Set the `DATABASE_URL` environment variable:

```
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable
```

Ensure PostgreSQL is running and accessible.

### 2. Run the API

- With Docker Compose (recommended):
  ```bash
  docker-compose up --build api
  ```
- Or run locally:
  ```bash
  go run main.go
  ```

## 📋 API Endpoints

| Method | Endpoint                         | Description                        |
| ------ | -------------------------------- | ---------------------------------- |
| GET    | `/api/v1/workflows/{id}`         | Load a workflow definition         |
| POST   | `/api/v1/workflows/{id}/execute` | Execute the workflow synchronously |

### Example Usage

#### GET workflow definition

```bash
curl http://localhost:8086/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000
```

#### POST execute workflow

```bash
curl -X POST http://localhost:8086/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute \
     -H "Content-Type: application/json" \
     -d '{"name":"John","email":"john@example.com","city":"Sydney","operator":"greater_than","threshold":25.0}'
```

You can also include a full workflow definition in the request to create or update the workflow at execution time:

```bash
curl -X POST http://localhost:8086/api/v1/workflows/550e8400-e29b-41d4-a716-446655440000/execute \
     -H "Content-Type: application/json" \
     -d '{
       "name":"John",
       "email":"john@example.com",
       "city":"Sydney",
       "operator":"greater_than",
       "threshold":25.0,
       "workflow":{
         "id":"550e8400-e29b-41d4-a716-446655440000",
         "name":"My Custom Workflow",
         "nodes":[
           {"id":"1","type":"start","position":{"x":250,"y":5},"data":{"label":"Start","description":"Starting point"}},
           {"id":"2","type":"form","position":{"x":250,"y":100},"data":{"label":"User Form","description":"Collect user data"}}
         ],
         "edges":[
           {"id":"e1-2","source":"1","target":"2","type":"smoothstep","animated":true,"style":{"stroke":"#AABBCC","strokeWidth":2}}
         ]
       }
     }'
```

When a workflow JSON is included in the request:
- If a workflow with the ID already exists, it will be updated with the new definition
- If no workflow exists with that ID, a new one will be created
- The updated or created workflow will then be executed with the provided input parameters

## 🗄️ Database

- The API uses `api/pkg/db.DefaultConfig()` and reads the URI from `DATABASE_URL`.
- For schema/configuration details, see the main project README or this file's comments.

## 📊 Database Schema

```mermaid
erDiagram
    WORKFLOWS {
        UUID id PK
        VARCHAR(255) name
        INTEGER version
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }
    
    WORKFLOW_NODES {
        UUID id PK
        UUID workflow_id FK
        VARCHAR(50) node_id
        VARCHAR(50) node_type
        FLOAT position_x
        FLOAT position_y
        VARCHAR(255) label
        TEXT description
        JSONB metadata
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }
    
    WORKFLOW_EDGES {
        UUID id PK
        UUID workflow_id FK
        VARCHAR(50) source_node_id
        VARCHAR(50) target_node_id
        VARCHAR(50) edge_id
        VARCHAR(50) type
        BOOLEAN animated
        VARCHAR(50) stroke_color
        INTEGER stroke_width
        VARCHAR(255) label
        VARCHAR(50) source_handle
        JSONB label_style
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }
    
    WORKFLOWS ||--o{ WORKFLOW_NODES : "contains"
    WORKFLOWS ||--o{ WORKFLOW_EDGES : "contains"
    WORKFLOW_NODES ||--o{ WORKFLOW_EDGES : "is source of"
    WORKFLOW_NODES ||--o{ WORKFLOW_EDGES : "is target of"
```

### Table Descriptions

#### WORKFLOWS
Stores the main workflow information including:
- **id**: UUID primary key
- **name**: Name of the workflow
- **version**: Version number of the workflow (increments on update)
- **created_at**: Timestamp when the workflow was created
- **updated_at**: Timestamp when the workflow was last updated

#### WORKFLOW_NODES
Stores individual nodes within workflows:
- **id**: UUID primary key
- **workflow_id**: Foreign key to the workflows table
- **node_id**: Identifier for the node within the workflow
- **node_type**: Type of node (start, form, integration, condition, email, end)
- **position_x/position_y**: Position coordinates for the node in the UI
- **label**: Display name for the node
- **description**: Longer text description of the node's purpose
- **metadata**: JSON data with node-specific configuration

#### WORKFLOW_EDGES
Stores connections between nodes:
- **id**: UUID primary key
- **workflow_id**: Foreign key to the workflows table
- **source_node_id**: ID of the starting node
- **target_node_id**: ID of the ending node
- **edge_id**: Identifier for the edge within the workflow
- **type**: Type of edge connection
- **animated**: Whether the edge should be animated in the UI
- **stroke_color/stroke_width**: Visual styling for the edge
- **label**: Text label for the edge
- **source_handle**: Connection point identifier on the source node
- **label_style**: JSON data with styling for the edge label

### Database Relationships
- A Workflow has many Nodes
- A Workflow has many Edges
- Nodes and Edges belong to a Workflow
- Each Edge connects two Nodes (source and target)

### Constraints and Indexes
- Unique constraint on (workflow_id, node_type)
- Indexes on workflow_id in both nodes and edges tables
- Indexes on source_node_id and target_node_id in edges table

## Project Structure
```
api/
├── .air.toml              # Air configuration for hot reloading
├── Dockerfile             # Docker configuration for containerization
├── README.md              # This documentation file
├── cmd/                   # Application entry points
│   └── api/               # Main API server
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── internal/              # Private application code
│   ├── api/               # API-specific code
│   ├── execution/         # Workflow execution engine
│   ├── handler/           # HTTP request handlers
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic layer
│   └── workflow/          # Workflow domain logic
├── migrations/            # Database migration scripts
├── pkg/                   # Reusable packages
│   ├── db/                # Database utilities
│   ├── log/               # Logging utilities
│   ├── mailer/            # Email sending functionality
│   ├── models/            # Shared data models
│   └── node/              # Node type implementations
│       ├── condition/     # Condition node logic
│       ├── email/         # Email node logic
│       ├── end/           # End node logic
│       ├── form/          # Form node logic
│       ├── integration/   # Integration node logic
│       │   └── weather/   # Weather API integration
│       └── start/         # Start node logic
├── scripts/               # Utility scripts
└── vendor/                # Vendored dependencies
```

## How It Works

### Workflow Engine Architecture

The workflow engine is built around a modular architecture that separates concerns and enables extensibility:

```
HTTP Request
   │
   ▼
Handler Layer
   │
   ▼
Service Layer
   │    │
   │    ▼
   │  Repository Layer
   │    │
   │    ▼
   │  Database
   │
   ▼
Execution Engine
   │
   ▼
Node Registry
   │
   ▼
Node Implementations
   │
   ▼
External Services
```

### Node Registry System

The Node Registry is a central component that manages all available node types:

```
Registry
│ +factories: map[NodeType]NodeFactory
│ +Register(nodeType, factory)
│ +Create(nodeModel) Node
│
├───► NodeFactory
│     │ +Create(model) Node
│     │
│     └───► Node (interface)
│           │ +Type() NodeType
│           │ +Execute(ctx, inputs) NodeOutputs
│           │ +Validate() error
│           │
│           ├─── StartNode
│           │    +Execute()
│           │
│           ├─── FormNode
│           │    +Execute()
│           │
│           ├─── IntegrationNode
│           │    +Execute()
│           │
│           ├─── ConditionNode
│           │    +Execute()
│           │
│           ├─── EmailNode
│           │    +Execute()
│           │
│           └─── EndNode
│                +Execute()
```
### Workflow Execution Process
![alt text](seqd.png)

The workflow execution process works as follows:

1. **Initialization**: The engine receives a workflow model and input data.

2. **Node and Edge Mapping**: It constructs an in-memory representation of the workflow by:
   - Creating node instances using the registry
   - Building an edge routing map for node connections
   - Identifying the start node

3. **Execution Loop**: Starting from the start node, it:
   - Executes the current node with appropriate inputs
   - Records the execution step
   - Determines the next node based on edge routing
   - Handles conditional routing for decision nodes
   - Continues until reaching an end node or error

4. **Result Collection**: It assembles execution results including:
   - Execution steps with timing information
   - Output data from each node
   - Overall workflow status and duration

The engine maintains state across nodes using a `priorOutputs` map, allowing nodes to access outputs from previously executed nodes.

# 🧩 Extending the Workflow System

### 📦 Adding New Node Types

The workflow system is designed to be easily extensible with new node types. Here's how to add a new type of node:

1. **Create a new node package**:
   
   Create a directory under `api/pkg/node/` for your new node type (e.g., `api/pkg/node/mynode/`).

2. **Implement the Node interface**:
   
   Create a Go file in your package (e.g., `mynode.go`) that implements the `node.Node` interface:

   ```go
   package mynode

   import (
       "context"
       "github.com/topu/workflow-challenge-v2/api/pkg/models"
       "github.com/topu/workflow-challenge-v2/api/pkg/node"
   )

   // Node implements a custom node type
   type Node struct {
       // Add any fields needed for your node
       baseInfo models.NodeBaseInfo
   }

   // NewNode creates a new instance of your node
   func NewNode(nodeModel models.Node) (*Node, error) {
       return &Node{
           baseInfo: models.NodeBaseInfo{
               ID:          nodeModel.ID,
               Label:       nodeModel.Label,
               Description: nodeModel.Description,
           },
       }, nil
   }

   // Execute implements the Node interface
   func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
       // Implement your node's custom logic here
       // Return outputs with appropriate data
   }

   // Type returns the node type
   func (n *Node) Type() models.NodeType {
       return "my_node_type" // Define a constant for this in models
   }

   // GetBaseInfo returns basic node info
   func (n *Node) GetBaseInfo() models.NodeBaseInfo {
       return n.baseInfo
   }
   ```

3. **Create a node factory function**:
   
   ```go
   // Factory creates a new node instance
   func Factory(nodeModel models.Node) (node.Node, error) {
       return NewNode(nodeModel)
   }
   ```

4. **Register your node type**:
   
   In `api/cmd/api/main.go`, add your node factory to the registry:

   ```go
   // Import your package
   import (
       "github.com/topu/workflow-challenge-v2/api/pkg/node/mynode"
   )

   // In the setup code where other nodes are registered:
   registry.Register(models.NodeTypeMyNode, mynode.Factory)
   ```

5. **Add node type constant**:
   
   In `api/pkg/models/types.go`, add a constant for your node type:

   ```go
   const (
       // ...existing node types...
       NodeTypeMyNode models.NodeType = "my_node_type"
   )
   ```


## Future Improvements

### Core Functionality
- **Bulk Operations**: Add support for bulk workflow creation and execution
- **Centralized Error Handling**: Create a consistent error handling system with custom error types to replace scattered `fmt.Errorf` calls.
- **Database Optimization**: Add indexes on frequently queried columns in the PostgreSQL database.
- **API Documentation**: Add comprehensive API documentation, possibly using Swagger/OpenAPI.

### Testing & Quality
- **Expanded Test Coverage**: More unit and integration tests
- **Performance Testing**: Benchmarks for workflow execution
- **Mocking Framework**: Better support for mocking external services in tests
- **Test Fixtures**: Reusable test data and workflow definitions
- **Workflow execution caching**
- **Concurrent execution of independent branches**
- **Adaptive batching of database operations**

## 📚 Key Libraries and Tools

### Core Libraries

#### UUID (github.com/google/uuid)
- **Why Chosen:** For generating unique identifiers
- **Benefits:**
  - Standard implementation of UUID generation
  - Provides globally unique identifiers without coordination
  - Useful for distributed systems where IDs need to be generated independently

#### gomail (gopkg.in/gomail.v2)(out of scope)
- **Why Chosen:** For sending emails from the email node
- **Benefits:**
  - Simple API for composing and sending emails
  - Support for attachments and HTML content
  - SMTP authentication and TLS support

#### Testify (github.com/stretchr/testify)
- **Why Chosen:** For unit and integration testing
- **Benefits:**
  - Rich assertion library
  - Mocking framework
  - Suite testing functionality

## 🔍 Key Design Decisions & Assumptions

### Workflow Structure
- **Single Start Node**: Each workflow must have exactly one start node
- **Directed Acyclic Graph**: Workflows cannot contain cycles (no cycle detection implemented)
- **Conditional Routing**: Only condition nodes can have multiple outgoing edges (true/false)

### Data Flow
- **Node Dependencies**: Weather conditions node assumes prior "weather-api" node output exists
- **State Management**: State is passed between nodes via outputs, with no global workflow state
- **Fixed Input Schema**: Weather integration expects specific parameters (lat/lon/city)

### External Services
- **Weather API**: Assumes reliable API availability with 10s default timeout
- **Email Service**: Email node assumes SMTP service availability

### Database Design
- **Cascading Deletion**: Deleting a workflow removes all associated nodes and edges
- **PostgreSQL-Specific**: Uses JSONB for flexible metadata storage

### Execution Model
- **Synchronous Processing**: Workflows execute in a blocking, synchronous manner
- **No Retry Logic**: Failed node execution fails the entire workflow
- **Pre-registered Nodes**: All node types must be registered before execution
