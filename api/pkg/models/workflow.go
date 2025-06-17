package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// NodeType represents the type of a workflow node
type NodeType string

// Valid node types
const (
	NodeTypeStart       NodeType = "start"
	NodeTypeForm        NodeType = "form"
	NodeTypeIntegration NodeType = "integration"
	NodeTypeCondition   NodeType = "condition"
	NodeTypeEmail       NodeType = "email"
	NodeTypeEnd         NodeType = "end"
)

// ValidNodeTypes is a map of valid node types
var ValidNodeTypes = map[NodeType]bool{
	NodeTypeStart:       true,
	NodeTypeForm:        true,
	NodeTypeIntegration: true,
	NodeTypeCondition:   true,
	NodeTypeEmail:       true,
	NodeTypeEnd:         true,
}

// Operator represents the type of comparison operator
type Operator string

// Valid operators for condition evaluation
const (
	OperatorGreaterThan        Operator = "greater_than"
	OperatorLessThan          Operator = "less_than"
	OperatorEquals            Operator = "equals"
	OperatorGreaterThanOrEqual Operator = "greater_than_or_equal"
	OperatorLessThanOrEqual   Operator = "less_than_or_equal"
)

// ValidOperators is a map of valid operators
var ValidOperators = map[Operator]bool{
	OperatorGreaterThan:        true,
	OperatorLessThan:          true,
	OperatorEquals:            true,
	OperatorGreaterThanOrEqual: true,
	OperatorLessThanOrEqual:   true,
}

// Status represents the status of a workflow execution or step
type Status string

// Valid status values
const (
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRunning   Status = "running"
)

// ValidStatuses is a map of valid status values
var ValidStatuses = map[Status]bool{
	StatusCompleted: true,
	StatusFailed:    true,
	StatusRunning:   true,
}

// Workflow represents a workflow definition in the database
type Workflow struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Version    int       `json:"version" db:"version"`
	Nodes      []Node    `json:"nodes"`
	Edges      []Edge    `json:"edges"`
	CreatedAt  time.Time `json:"-" db:"created_at"`
	UpdatedAt  time.Time `json:"-" db:"updated_at"`
}

// WorkflowExecution represents the execution of a workflow
type WorkflowExecution struct {
	ID            string         `json:"id" db:"id"`
	WorkflowID    string         `json:"-" db:"workflow_id"`
	Status        Status         `json:"status" db:"status"` // 'completed', 'failed', or 'cancelled'
	StartTime     string         `json:"startTime" db:"start_time"`
	EndTime       string         `json:"endTime" db:"end_time"`
	TotalDuration int64          `json:"totalDuration,omitempty" db:"total_duration"`
	Steps         []ExecutionStep `json:"steps" db:"-"`
	Metadata      JSONB          `json:"metadata,omitempty" db:"metadata"`
	ExecutedAt    time.Time      `json:"-" db:"executed_at"` // Kept for internal use
}

// ExecutionStep represents a single step in the workflow execution
type ExecutionStep struct {
	NodeID      string    `json:"-" db:"node_id"`
	StepNumber  int       `json:"stepNumber" db:"step_number"`
	NodeType    NodeType  `json:"nodeType" db:"node_type"`  // Changed from Type
	Status      Status    `json:"status" db:"status"`       // 'completed', 'failed', or 'cancelled'
	Label       string    `json:"-" db:"label"`             // Hidden in frontend
	Description string    `json:"-" db:"description"`       // Hidden in frontend
	Duration    int64     `json:"duration" db:"duration"`   // Duration in milliseconds
	Output      JSONB     `json:"output" db:"output"`       // Contains message, details, and other specific fields
	Timestamp   string    `json:"timestamp" db:"timestamp"` // Single timestamp for frontend
	Error       string    `json:"error,omitempty" db:"error"`
	StartedAt   string    `json:"-" db:"-"`                 // Used internally
	EndedAt     string    `json:"-" db:"-"`                 // Used internally
}

// WorkflowInput represents the input data for workflow execution
type WorkflowInput struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	City      string   `json:"city"`
	Threshold float64  `json:"threshold"`
	Operator  Operator `json:"operator"`
	Workflow  JSONB    `json:"workflow"`
}

// Validate validates the workflow input
func (w *WorkflowInput) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("name is required")
	}
	if w.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Basic email validation
	if !strings.Contains(w.Email, "@") || !strings.Contains(w.Email, ".") {
		return fmt.Errorf("invalid email format")
	}
	if w.City == "" {
		return fmt.Errorf("city is required")
	}
	if !ValidOperators[w.Operator] {
		return fmt.Errorf("invalid operator: %s", w.Operator)
	}
	if w.Threshold < 0 {
		return fmt.Errorf("temperature cannot be negative")
	}
	if w.Threshold > 100 {
		return fmt.Errorf("temperature must be below 100°C")
	}
	return nil
}

// JSONB is a custom type for handling JSONB data
type JSONB map[string]any

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

// IsValid checks if the NodeType is valid
func (t NodeType) IsValid() bool {
	_, ok := ValidNodeTypes[t]
	return ok
}

// IsValid checks if the Operator is valid
func (o Operator) IsValid() bool {
	_, ok := ValidOperators[o]
	return ok
}

// NodeData represents the data associated with a node
type NodeData struct {
	Label       string         `json:"label"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
}

// NodeID represents a node identifier in the workflow
type NodeID string

// Valid node IDs
const (
	NodeIDStart       NodeID = "start"
	NodeIDForm        NodeID = "form"
	NodeIDWeatherAPI  NodeID = "weather-api"
	NodeIDCondition   NodeID = "condition"
	NodeIDEmail       NodeID = "email"
	NodeIDEnd         NodeID = "end"
)

// ValidNodeIDs is a map of valid node IDs
var ValidNodeIDs = map[NodeID]bool{
	NodeIDStart:       true,
	NodeIDForm:        true,
	NodeIDWeatherAPI:  true,
	NodeIDCondition:   true,
	NodeIDEmail:       true,
	NodeIDEnd:         true,
}

// OutputKey represents a key in the node output
type OutputKey string

// Valid output keys
const (
	OutputKeyName         OutputKey = "name"
	OutputKeyEmail        OutputKey = "email"
	OutputKeyCity         OutputKey = "city"
	OutputKeyTemperature  OutputKey = "temperature"
	OutputKeyLocation     OutputKey = "location"
	OutputKeyConditionMet OutputKey = "conditionMet"
	OutputKeyError        OutputKey = "error"
)

// ValidOutputKeys is a map of valid output keys
var ValidOutputKeys = map[OutputKey]bool{
	OutputKeyName:         true,
	OutputKeyEmail:        true,
	OutputKeyCity:         true,
	OutputKeyTemperature:  true,
	OutputKeyLocation:     true,
	OutputKeyConditionMet: true,
	OutputKeyError:        true,
}