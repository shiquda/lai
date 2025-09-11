package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global configuration",
	Long:  `Manage global configuration settings for lai.

Available commands:
  interactive  Launch interactive TUI configuration interface (recommended)
  set          Set a configuration value via command line
  get          Get a configuration value
  list         List all configuration values
  reset        Reset configuration to defaults

Examples:
  lai config interactive       # Launch interactive config interface
  lai config set notifications.openai.api_key "sk-your-key"
  lai config get notifications.openai.api_key
  lai config list
  lai config reset`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  "Set a configuration value in the global config file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debugf("Config set command called with args: %v", args)
		key := args[0]
		value := args[1]

		if err := setConfigValue(key, value); err != nil {
			logger.Printf("Error setting config: %v\n", err)
			os.Exit(1)
		}

		logger.Printf("Set %s = %s\n", key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  "Get a configuration value from the global config file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		value, err := getConfigValue(key)
		if err != nil {
			logger.Errorf("Error getting config: %v", err)
			os.Exit(1)
		}

		logger.Printf("%s = %s\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  "List all configuration values in the global config file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listConfigValues(); err != nil {
			logger.Errorf("Error listing config: %v", err)
			os.Exit(1)
		}
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long:  "Reset the global configuration to default values",
	Run: func(cmd *cobra.Command, args []string) {
		if err := resetConfigValues(); err != nil {
			logger.Errorf("Error resetting config: %v", err)
			os.Exit(1)
		}

		logger.Println("Configuration reset to defaults")
	},
}

var configInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive configuration interface",
	Long:  "Launch an interactive TUI for managing configuration settings with a user-friendly interface",
	Aliases: []string{"i", "tui", "ui"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInteractiveConfig(); err != nil {
			logger.Errorf("Failed to run interactive config: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configInteractiveCmd)
}

func setConfigValue(key, value string) error {
	// Load current global config
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Set the value using reflection
	if err := setFieldByPath(cfg, key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Save the updated config
	if err := config.SaveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func getConfigValue(key string) (string, error) {
	// Load global config
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load global config: %w", err)
	}

	// Get the value using reflection
	value, err := getFieldByPath(cfg, key)
	if err != nil {
		return "", fmt.Errorf("failed to get config value: %w", err)
	}

	return value, nil
}

func listConfigValues() error {
	// Load global config
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Print all config values
	printConfig(cfg, "")
	return nil
}

func resetConfigValues() error {
	// Get default config from config package
	defaultConfig := config.GetDefaultGlobalConfig()

	// Save default config
	if err := config.SaveGlobalConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to save default config: %w", err)
	}

	return nil
}

// setFieldByPath sets a field value by dot notation path (e.g., "notifications.openai.api_key")
func setFieldByPath(obj interface{}, path, value string) error {
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

// getFieldByPath gets a field value by dot notation path
func getFieldByPath(obj interface{}, path string) (string, error) {
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

// printConfig recursively prints all configuration values
func printConfig(obj interface{}, prefix string) {
	v := reflect.ValueOf(obj).Elem()
	t := reflect.TypeOf(obj).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := toSnakeCase(fieldType.Name)

		fullPath := fieldName
		if prefix != "" {
			fullPath = prefix + "." + fieldName
		}

		if field.Kind() == reflect.Struct {
			printConfig(field.Addr().Interface(), fullPath)
		} else {
			logger.Printf("%s = %v\n", fullPath, field.Interface())
		}
	}
}

// toCamelCase converts snake_case to CamelCase (with special handling for common acronyms)
func toCamelCase(s string) string {
	// Handle special compound words first
	switch s {
	case "open_ai":
		return "OpenAI"
	case "openai": // Also support the simplified form
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

// runInteractiveConfig launches the interactive TUI configuration interface
func runInteractiveConfig() error {
	// Ensure global config exists before starting TUI
	if err := config.EnsureGlobalConfig(); err != nil {
		return fmt.Errorf("failed to ensure global config: %w", err)
	}
	
	// Create the TUI model
	model, err := tui.NewConfigModel()
	if err != nil {
		return fmt.Errorf("failed to create config model: %w", err)
	}
	
	// Create and run the Bubble Tea program
	program := tea.NewProgram(model, tea.WithAltScreen())
	
	// Run the program
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run interactive config: %w", err)
	}
	
	// Check if there were any final errors
	if configModel, ok := finalModel.(*tui.ConfigModel); ok {
		if configModel.HasError() {
			return configModel.GetError()
		}
	}
	
	logger.Println("Interactive configuration completed successfully")
	return nil
}
