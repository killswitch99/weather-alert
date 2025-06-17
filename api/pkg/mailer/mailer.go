package mailer

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	mail "gopkg.in/gomail.v2"
)

// EmailTemplate represents a template for email content
type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// PrepareAndStubSendEmail prepares an email using gomail and logs the payload (does not send).
func PrepareAndStubSendEmail(to string, variables map[string]any, template EmailTemplate) (map[string]any, error) {
	m := mail.NewMessage()
	m.SetHeader("From", "weather-alerts@checkbox.com")
	m.SetHeader("To", to)

	// Process subject and body using provided variables
	subject := processTemplate(template.Subject, variables)
	body := processTemplate(template.Body, variables)

	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	slog.Debug(fmt.Sprintf("[STUB EMAIL] Would send: To=%s, Subject=%s", to, subject))

	return map[string]any{
		"to":        to,
		"from":      "weather-alerts@checkbox.com",
		"subject":   subject,
		"body":      body,
		"variables": variables,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

// processTemplate replaces template placeholders {{variable}} with actual values
func processTemplate(template string, variables map[string]any) string {
	result := template

	// Replace each variable in the template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)

		// Convert value to string based on type
		var stringValue string
		switch v := value.(type) {
		case float64:
			stringValue = fmt.Sprintf("%.1f", v)
		case int:
			stringValue = fmt.Sprintf("%d", v)
		case string:
			stringValue = v
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		result = strings.Replace(result, placeholder, stringValue, -1)
	}

	return result
}
