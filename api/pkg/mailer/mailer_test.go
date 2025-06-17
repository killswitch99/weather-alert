package mailer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareAndStubSendEmail(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		to          string
		variables   map[string]any
		template    EmailTemplate
		expectError bool
	}{
		{
			name: "Valid email preparation",
			to:   "test@example.com",
			variables: map[string]any{
				"name":        "John Doe",
				"city":        "New York",
				"temperature": 25.5,
			},
			template: EmailTemplate{
				Subject: "Weather Alert",
				Body:    "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
			},
			expectError: false,
		},
		{
			name: "Email with multiple variables",
			to:   "another@example.com",
			variables: map[string]any{
				"name":        "Jane Smith",
				"city":        "San Francisco",
				"temperature": 18.3,
				"humidity":    75,
				"condition":   "Cloudy",
			},
			template: EmailTemplate{
				Subject: "Weather Report for {{city}}",
				Body:    "Hello {{name}}, the weather in {{city}} is {{condition}} with {{temperature}}°C and {{humidity}}% humidity.",
			},
			expectError: false,
		},
		{
			name: "Email with missing template variables",
			to:   "missing@example.com",
			variables: map[string]any{
				"city": "London",
				// Missing temperature
			},
			template: EmailTemplate{
				Subject: "Weather Alert",
				Body:    "Weather alert for {{city}}! Temperature is {{temperature}}°C!",
			},
			expectError: false, // Should not error, just leave the placeholder
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			result, err := PrepareAndStubSendEmail(tc.to, tc.variables, tc.template)

			// Check error status
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Check that the returned map contains expected values
				assert.Equal(t, tc.to, result["to"])
				assert.Equal(t, "weather-alerts@checkbox.com", result["from"])
				
				// Check subject was processed correctly
				processedSubject := processTemplate(tc.template.Subject, tc.variables)
				assert.Equal(t, processedSubject, result["subject"])
				
				// Check body was processed correctly
				processedBody := processTemplate(tc.template.Body, tc.variables)
				assert.Equal(t, processedBody, result["body"])
				
				// Check variables were included
				assert.Equal(t, tc.variables, result["variables"])
				
				// Check if timestamp was included
				assert.Contains(t, result, "timestamp")
				
				// Additional specific checks for the test cases
				if tc.name == "Valid email preparation" {
					assert.Contains(t, result["body"], "Weather alert for New York")
					assert.Contains(t, result["body"], "Temperature is 25.5°C")
				} else if tc.name == "Email with multiple variables" {
					assert.Equal(t, "Weather Report for San Francisco", result["subject"])
					assert.Contains(t, result["body"], "Hello Jane Smith")
					assert.Contains(t, result["body"], "the weather in San Francisco is Cloudy")
				} else if tc.name == "Email with missing template variables" {
					assert.Contains(t, result["body"], "Temperature is {{temperature}}°C")
				}
			}
		})
	}
}

func TestProcessTemplate(t *testing.T) {
	testCases := []struct {
		name         string
		template     string
		variables    map[string]any
		expected     string
	}{
		{
			name:     "Simple text replacement",
			template: "Hello {{name}}!",
			variables: map[string]any{
				"name": "John",
			},
			expected: "Hello John!",
		},
		{
			name:     "Multiple replacements",
			template: "{{greeting}} {{name}}! The weather is {{temperature}}°C.",
			variables: map[string]any{
				"greeting":    "Hello",
				"name":        "Alice",
				"temperature": 22.5,
			},
			expected: "Hello Alice! The weather is 22.5°C.",
		},
		{
			name:     "Different variable types",
			template: "Count: {{count}}, Active: {{active}}, Rate: {{rate}}",
			variables: map[string]any{
				"count":  42,
				"active": true,
				"rate":   3.14,
			},
			expected: "Count: 42, Active: true, Rate: 3.1",
		},
		{
			name:     "Missing variables",
			template: "Hello {{name}}! Today is {{day}}.",
			variables: map[string]any{
				"name": "Bob",
				// day is missing
			},
			expected: "Hello Bob! Today is {{day}}.",
		},
		{
			name:     "No variables needed",
			template: "This is a plain text with no variables.",
			variables: map[string]any{},
			expected: "This is a plain text with no variables.",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processTemplate(tc.template, tc.variables)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper function to format temperature same way as in the main function
func getFormattedTemperature(temp float64) string {
	return getString(temp)
}

// getString formats float64 to string with one decimal place
func getString(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
