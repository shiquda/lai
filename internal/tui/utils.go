package tui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shiquda/lai/internal/config"
)

// getFieldByPath gets a field value by dot notation path
func getFieldByPath(obj interface{}, path string) (string, error) {
	// Handle providers configuration first
	if strings.Contains(path, "providers") {
		return getProviderFieldValue(obj, path)
	}

	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj).Elem()

	for _, part := range parts {
		field := v.FieldByName(toCamelCase(part))
		if !field.IsValid() {
			return "", fmt.Errorf("field %s not found", part)
		}
		v = field
	}

	return fmt.Sprintf("%v", v.Interface()), nil
}

// setFieldByPath sets a field value by dot notation path
func setFieldByPath(obj interface{}, path, value string) error {
	// Handle providers configuration first
	if strings.Contains(path, "providers") {
		return setProviderFieldValue(obj, path, value)
	}

	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj).Elem()

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			field := v.FieldByName(toCamelCase(part))
			if !field.IsValid() {
				return fmt.Errorf("field %s not found", part)
			}
			if !field.CanSet() {
				return fmt.Errorf("field %s cannot be set", part)
			}

			// Convert string value to appropriate type
			switch field.Kind() {
			case reflect.String:
				field.SetString(value)
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(value); err == nil {
					field.SetBool(boolVal)
				} else {
					return fmt.Errorf("invalid bool value: %s", value)
				}
			case reflect.Int:
				if intVal, err := strconv.Atoi(value); err == nil {
					field.SetInt(int64(intVal))
				} else {
					return fmt.Errorf("invalid int value: %s", value)
				}
			case reflect.Int64:
				// Handle time.Duration (which is int64)
				if field.Type() == reflect.TypeOf(time.Duration(0)) {
					if duration, err := time.ParseDuration(value); err == nil {
						field.SetInt(int64(duration))
					} else {
						return fmt.Errorf("invalid duration value: %s", value)
					}
				} else {
					if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
						field.SetInt(intVal)
					} else {
						return fmt.Errorf("invalid int64 value: %s", value)
					}
				}
			case reflect.Slice:
				// Handle string slices (for to_emails, etc.)
				if field.Type().Elem().Kind() == reflect.String {
					var strSlice []string
					if value != "" {
						strSlice = strings.Split(value, ",")
						for i := range strSlice {
							strSlice[i] = strings.TrimSpace(strSlice[i])
						}
					}
					field.Set(reflect.ValueOf(strSlice))
				} else {
					return fmt.Errorf("unsupported slice type: %s", field.Type())
				}
			default:
				return fmt.Errorf("unsupported field type: %s", field.Kind())
			}
		} else {
			// Navigate to nested struct
			field := v.FieldByName(toCamelCase(part))
			if !field.IsValid() {
				return fmt.Errorf("field %s not found", part)
			}
			v = field
		}
	}

	return nil
}

// getProviderFieldValue handles getting values from the providers map
func getProviderFieldValue(obj interface{}, path string) (string, error) {
	globalConfig := obj.(*config.GlobalConfig)

	// Parse path like "notifications.providers.telegram.enabled"
	parts := strings.Split(path, ".")
	if len(parts) < 4 || parts[0] != "notifications" || parts[1] != "providers" {
		return "", fmt.Errorf("invalid provider path: %s", path)
	}

	providerName := parts[2]
	fieldName := strings.Join(parts[3:], ".")

	// Get the provider from the map
	provider, exists := globalConfig.Notifications.Providers[providerName]
	if !exists {
		// Return default value for non-existent providers
		return "", nil
	}

	// Handle different field paths
	switch {
	case fieldName == "enabled":
		return strconv.FormatBool(provider.Enabled), nil
	case strings.HasPrefix(fieldName, "config."):
		configKey := strings.TrimPrefix(fieldName, "config.")
		if val, ok := provider.Config[configKey]; ok {
			return fmt.Sprintf("%v", val), nil
		}
		return "", nil
	case strings.HasPrefix(fieldName, "defaults."):
		defaultKey := strings.TrimPrefix(fieldName, "defaults.")
		if val, ok := provider.Defaults[defaultKey]; ok {
			return fmt.Sprintf("%v", val), nil
		}
		return "", nil
	}

	return "", fmt.Errorf("unknown provider field: %s", fieldName)
}

// setProviderFieldValue handles setting values in the providers map
func setProviderFieldValue(obj interface{}, path, value string) error {
	globalConfig := obj.(*config.GlobalConfig)

	// Parse path like "notifications.providers.telegram.enabled"
	parts := strings.Split(path, ".")
	if len(parts) < 4 || parts[0] != "notifications" || parts[1] != "providers" {
		return fmt.Errorf("invalid provider path: %s", path)
	}

	providerName := parts[2]
	fieldName := strings.Join(parts[3:], ".")

	// Ensure providers map exists
	if globalConfig.Notifications.Providers == nil {
		globalConfig.Notifications.Providers = make(map[string]config.ServiceConfig)
	}

	// Get or create the provider
	provider, exists := globalConfig.Notifications.Providers[providerName]
	if !exists {
		provider = config.ServiceConfig{
			Enabled:  false,
			Provider: providerName,
			Config:   make(map[string]interface{}),
			Defaults: make(map[string]interface{}),
		}
	}

	// Handle different field paths
	switch {
	case fieldName == "enabled":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			provider.Enabled = boolVal
		} else {
			return fmt.Errorf("invalid bool value: %s", value)
		}
	case strings.HasPrefix(fieldName, "config."):
		configKey := strings.TrimPrefix(fieldName, "config.")
		if provider.Config == nil {
			provider.Config = make(map[string]interface{})
		}

		// Try to convert value to appropriate type
		provider.Config[configKey] = convertStringToInterface(value, configKey)

	case strings.HasPrefix(fieldName, "defaults."):
		defaultKey := strings.TrimPrefix(fieldName, "defaults.")
		if provider.Defaults == nil {
			provider.Defaults = make(map[string]interface{})
		}
		provider.Defaults[defaultKey] = convertStringToInterface(value, defaultKey)

	default:
		return fmt.Errorf("unknown provider field: %s", fieldName)
	}

	// Save the updated provider back to the map
	globalConfig.Notifications.Providers[providerName] = provider

	return nil
}

// convertStringToInterface converts string values to appropriate interface{} types
func convertStringToInterface(value, key string) interface{} {
	// Handle known field types
	switch key {
	case "smtp_port":
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	case "use_tls":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	case "to_emails":
		// Handle string arrays
		if value != "" {
			emails := strings.Split(value, ",")
			for i := range emails {
				emails[i] = strings.TrimSpace(emails[i])
			}
			return emails
		}
		return []string{}
	}

	// Default to string
	return value
}

// toCamelCase converts snake_case to CamelCase (with special handling for common acronyms)
func toCamelCase(s string) string {
	// Handle special compound words first
	switch s {
	case "open_ai", "openai":
		return "OpenAI"
	}

	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			lower := strings.ToLower(parts[i])
			// Special cases for common acronyms
			switch lower {
			case "id":
				parts[i] = "ID"
			case "api":
				parts[i] = "API"
			case "url":
				parts[i] = "URL"
			case "ai":
				parts[i] = "AI"
			default:
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
	}
	return strings.Join(parts, "")
}

// toSnakeCase converts CamelCase to snake_case (with proper handling of consecutive uppercase)
func toSnakeCase(s string) string {
	// Handle special compound words first
	switch s {
	case "OpenAI":
		return "openai"
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if this is part of an acronym (consecutive uppercase letters)
			prevIsUpper := i > 0 && runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			nextIsLower := i < len(runes)-1 && runes[i+1] >= 'a' && runes[i+1] <= 'z'

			// Add underscore if:
			// 1. Previous char is lowercase (normal CamelCase transition)
			// 2. This is the last uppercase in an acronym followed by lowercase
			if !prevIsUpper || nextIsLower {
				result.WriteByte('_')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
