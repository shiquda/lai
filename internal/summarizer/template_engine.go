package summarizer

import (
	"fmt"
	"regexp"
	"strings"
)

// TemplateEngine handles template variable substitution
type TemplateEngine struct {
	// Built-in variables that are always available
	builtinVariables map[string]string
}

// NewTemplateEngine creates a new template engine with built-in variables
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		builtinVariables: map[string]string{
			// Language and output format variables
			"language": "English",

			// System information variables
			"system": "Lai Log Monitor",
			"version": "latest",

			// Common formatting variables
			"timestamp": "{{.Timestamp}}", // This will be replaced with actual timestamp at runtime
		},
	}
}

// SetBuiltinVariable sets a built-in variable value
func (e *TemplateEngine) SetBuiltinVariable(key, value string) {
	e.builtinVariables[key] = value
}

// RenderTemplate renders a template with the given variables
func (e *TemplateEngine) RenderTemplate(template string, variables map[string]string) (string, error) {
	if template == "" {
		return "", fmt.Errorf("template cannot be empty")
	}

	// Merge custom variables with built-in variables (custom take precedence)
	mergedVariables := make(map[string]string)
	for k, v := range e.builtinVariables {
		mergedVariables[k] = v
	}
	for k, v := range variables {
		mergedVariables[k] = v
	}

	// Replace variables in the template
	result := template
	var err error

	// Replace {{variable}} format
	result, err = e.replaceVariables(result, mergedVariables, "{{", "}}")
	if err != nil {
		return "", fmt.Errorf("failed to replace variables: %w", err)
	}

	// Replace ${variable} format (alternative syntax)
	result, err = e.replaceVariables(result, mergedVariables, "${", "}")
	if err != nil {
		return "", fmt.Errorf("failed to replace variables: %w", err)
	}

	// Replace $variable format (simple syntax)
	result, err = e.replaceSimpleVariables(result, mergedVariables)
	if err != nil {
		return "", fmt.Errorf("failed to replace simple variables: %w", err)
	}

	return result, nil
}

// replaceVariables replaces variables in the format prefix + variable + suffix
func (e *TemplateEngine) replaceVariables(template string, variables map[string]string, prefix, suffix string) (string, error) {
	result := template

	// Find all variables in the format prefix + variable + suffix
	pattern := regexp.MustCompile(regexp.QuoteMeta(prefix) + `([^` + regexp.QuoteMeta(suffix) + `]+)` + regexp.QuoteMeta(suffix))
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMatch := match[0]
		variableName := strings.TrimSpace(match[1])

		// Skip if variable name is empty
		if variableName == "" {
			continue
		}

		// Get variable value
		value, exists := variables[variableName]
		if !exists {
			// Variable not found, leave as-is or handle error
			// For now, leave as-is to be backward compatible
			continue
		}

		// Replace the variable with its value
		result = strings.ReplaceAll(result, fullMatch, value)
	}

	return result, nil
}

// replaceSimpleVariables replaces simple $variable format
func (e *TemplateEngine) replaceSimpleVariables(template string, variables map[string]string) (string, error) {
	result := template

	// This is more complex because we need to handle variable names with alphanumeric and underscore
	// We'll use a regex to find $ followed by valid variable name characters
	pattern := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMatch := match[0]
		variableName := match[1]

		// Get variable value
		value, exists := variables[variableName]
		if !exists {
			// Variable not found, leave as-is
			continue
		}

		// Replace the variable with its value
		result = strings.ReplaceAll(result, fullMatch, value)
	}

	return result, nil
}

// ValidateTemplate validates a template by checking for undefined variables
func (e *TemplateEngine) ValidateTemplate(template string, allowedVariables map[string]bool) error {
	if template == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Check for undefined variables in all formats
	var undefinedVars []string

	// Check {{variable}} format
	undefinedVars = append(undefinedVars, e.findUndefinedVariables(template, "{{", "}}", allowedVariables)...)

	// Check ${variable} format
	undefinedVars = append(undefinedVars, e.findUndefinedVariables(template, "${", "}", allowedVariables)...)

	// Check $variable format
	undefinedVars = append(undefinedVars, e.findUndefinedSimpleVariables(template, allowedVariables)...)

	if len(undefinedVars) > 0 {
		return fmt.Errorf("undefined variables found in template: %v", undefinedVars)
	}

	return nil
}

// findUndefinedVariables finds undefined variables in the format prefix + variable + suffix
func (e *TemplateEngine) findUndefinedVariables(template string, prefix, suffix string, allowedVariables map[string]bool) []string {
	pattern := regexp.MustCompile(regexp.QuoteMeta(prefix) + `([^` + regexp.QuoteMeta(suffix) + `]+)` + regexp.QuoteMeta(suffix))
	matches := pattern.FindAllStringSubmatch(template, -1)

	var undefinedVars []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		variableName := strings.TrimSpace(match[1])
		if variableName != "" && !allowedVariables[variableName] && !e.isBuiltinVariable(variableName) {
			undefinedVars = append(undefinedVars, variableName)
		}
	}

	return undefinedVars
}

// findUndefinedSimpleVariables finds undefined simple $variable format
func (e *TemplateEngine) findUndefinedSimpleVariables(template string, allowedVariables map[string]bool) []string {
	pattern := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := pattern.FindAllStringSubmatch(template, -1)

	var undefinedVars []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		variableName := match[1]
		if !allowedVariables[variableName] && !e.isBuiltinVariable(variableName) {
			undefinedVars = append(undefinedVars, variableName)
		}
	}

	return undefinedVars
}

// isBuiltinVariable checks if a variable is a built-in variable
func (e *TemplateEngine) isBuiltinVariable(name string) bool {
	_, exists := e.builtinVariables[name]
	return exists
}

// GetBuiltinVariables returns all built-in variables
func (e *TemplateEngine) GetBuiltinVariables() map[string]string {
	result := make(map[string]string)
	for k, v := range e.builtinVariables {
		result[k] = v
	}
	return result
}

// GetSupportedVariableFormats returns the supported variable formats
func (e *TemplateEngine) GetSupportedVariableFormats() []string {
	return []string{
		"{{variable}}",  // Double brace format
		"${variable}",   // Dollar brace format
		"$variable",     // Simple dollar format
	}
}