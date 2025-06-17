package email

import (
	"context"
	"fmt"
	"time"
	"workflow-code-test/api/pkg/mailer"
	"workflow-code-test/api/pkg/models"
	"workflow-code-test/api/pkg/node"
)

// Node implements an email node
type Node struct {
	node.BaseNode
	InputVariables []string            `json:"inputVariables"`
	EmailTemplate  mailer.EmailTemplate `json:"emailTemplate"`
}

// NewNode creates an email node from a model
func NewNode(model models.Node) (node.Node, error) {
	emailNode := &Node{
		BaseNode: node.BaseNode{
			ID:          model.ID,
			Label:       model.Data.Label,
			Description: model.Data.Description,
		},
	}
	
	// Extract metadata fields if available
	if meta, ok := model.Data.Metadata["inputVariables"]; ok {
		// Get input variables
		if inputVars, ok := meta.([]any); ok {
			for _, v := range inputVars {
				if strVar, ok := v.(string); ok {
					emailNode.InputVariables = append(emailNode.InputVariables, strVar)
				}
			}
		}
		
		// Get email template
		if templateData, ok := model.Data.Metadata["emailTemplate"]; ok {
			if template, ok := templateData.(map[string]any); ok {
				if subject, ok := template["subject"].(string); ok {
					emailNode.EmailTemplate.Subject = subject
				}
				if body, ok := template["body"].(string); ok {
					emailNode.EmailTemplate.Body = body
				}
			}
		}
	}
	
	return emailNode, nil
}

// Type returns the node type
func (n *Node) Type() models.NodeType {
	return models.NodeTypeEmail
}

// GetBaseInfo returns the base node information
func (n *Node) GetBaseInfo() node.BaseNode {
	return n.BaseNode
}

// Execute implements the email sending logic
func (n *Node) Execute(ctx context.Context, inputs node.NodeInputs) (node.NodeOutputs, error) {
	started := time.Now()
	outputs := node.NodeOutputs{
		Data:      make(map[string]any),
		Status:    models.StatusRunning,
		StartedAt: started.Format(time.RFC3339),
	}
	
	// Check if condition was met from prior condition node
	conditionNodeOutput, ok := inputs.PriorOutputs[string(models.NodeIDCondition)]
	if !ok {
		outputs.Status = models.StatusFailed
		outputs.Data["message"] = "Failed to process email"
		outputs.Data["error"] = "Failed to get condition result"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("failed to get condition result")
	}
	
	// Get the condition result from the new structure
	conditionResult, ok := conditionNodeOutput.Data["conditionResult"].(map[string]any)
	if !ok {
		outputs.Status = models.StatusFailed
		outputs.Data["message"] = "Failed to process email"
		outputs.Data["error"] = "Failed to get condition result"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("invalid condition result format")
	}
	
	conditionMet, ok := conditionResult["result"].(bool)
	if !ok {
		outputs.Status = models.StatusFailed
		outputs.Data["message"] = "Failed to process email"
		outputs.Data["error"] = "Failed to get condition result"
		outputs.EndedAt = time.Now().Format(time.RFC3339)
		return outputs, fmt.Errorf("invalid condition result format")
	}
	
	if conditionMet {
		// Get required info from form outputs
		formOutput, ok := inputs.PriorOutputs[string(models.NodeIDForm)]
		if !ok {
			outputs.Status = models.StatusFailed
			outputs.Data["message"] = "Failed to process email"
			outputs.Data["error"] = "Failed to get form data"
			outputs.EndedAt = time.Now().Format(time.RFC3339)
			return outputs, fmt.Errorf("missing form data")
		}
		
		// Get email recipient
		email, ok := formOutput.Data["email"].(string)
		if !ok {
			outputs.Status = models.StatusFailed
			outputs.Data["message"] = "Failed to process email"
			outputs.Data["error"] = "Failed to get email from form output"
			outputs.EndedAt = time.Now().Format(time.RFC3339)
			return outputs, fmt.Errorf("missing email")
		}
		
		// Collect all template variables from various node outputs
		templateVars := make(map[string]any)
		
		// Collect all required input variables from prior outputs
		for _, varName := range n.InputVariables {
			// For each input variable, check in all prior outputs
			found := false
			
			for _, output := range inputs.PriorOutputs {
				if value, ok := output.Data[varName]; ok {
					templateVars[varName] = value
					found = true
					break
				}
			}
			
			if !found {
				outputs.Status = models.StatusFailed
				outputs.Data["message"] = "Failed to process email"
				outputs.Data["error"] = fmt.Sprintf("Missing required variable: %s", varName)
				outputs.EndedAt = time.Now().Format(time.RFC3339)
				return outputs, fmt.Errorf("missing required variable: %s", varName)
			}
		}
		
		// Use the mailer with template support
		emailPayload, err := mailer.PrepareAndStubSendEmail(email, templateVars, n.EmailTemplate)
		if err != nil {
			outputs.Status = models.StatusFailed
			outputs.Data["error"] = fmt.Sprintf("Failed to send email: %v", err)
			outputs.EndedAt = time.Now().Format(time.RFC3339)
			return outputs, fmt.Errorf("email sending failed: %w", err)
		}
		
		// Prepare output data in the format expected by the frontend
		subject, _ := emailPayload["subject"].(string)
		body, _ := emailPayload["body"].(string)
		timestamp := time.Now().Format(time.RFC3339)
		
		// Set the output data using the response from the mailer to match frontend expectations
		outputs.Data = map[string]any{
			"message": "Email sent successfully",
			"details": map[string]any{
				"outputVariables": []string{"emailSent"},
			},
			"emailContent": map[string]any{
				"to":        email,
				"subject":   subject,
				"body":      body,
				"timestamp": timestamp,
			},
		}
	} else {
		outputs.Data = map[string]any{
			"message": "Email not sent - condition not met",
			"details": map[string]any{
				"reason": "Condition not met",
			},
		}
	}
	
	outputs.Status = models.StatusCompleted
	outputs.EndedAt = time.Now().Format(time.RFC3339)
	return outputs, nil
}

// Validate ensures the node is properly configured
func (n *Node) Validate() error {
	// Ensure we have at least some input variables and a template
	if len(n.InputVariables) == 0 {
		return fmt.Errorf("email node requires at least one input variable")
	}
	
	if n.EmailTemplate.Subject == "" || n.EmailTemplate.Body == "" {
		return fmt.Errorf("email node requires both subject and body templates")
	}
	
	return nil
}