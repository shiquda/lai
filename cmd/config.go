package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
	Long: `Manage global configuration settings for lai.

Available commands:
  interactive  Launch interactive TUI configuration interface (recommended)
  set          Set a configuration value via command line
  get          Get a configuration value
  list         List all configuration values
  reset        Reset configuration to defaults
  backup       Create a backup of current configuration
  restore      Restore configuration from a backup
  list-backups List all configuration backups

Examples:
  lai config interactive       # Launch interactive config interface
  lai config set notifications.openai.api_key "sk-your-key"
  lai config get notifications.openai.api_key
  lai config list
  lai config reset
  lai config backup            # Create backup
  lai config list-backups      # List backups
  lai config restore backup.yaml # Restore from backup`,
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
			logger.UserErrorf("Error setting config: %v", err)
			os.Exit(1)
		}

		logger.UserSuccessf("Set %s = %s", key, value)
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

		logger.UserInfof("%s = %s\n", key, value)
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

		logger.UserSuccess("Configuration reset to defaults")
	},
}

var configBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the current configuration",
	Long:  "Create a backup of the current global configuration with version information",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := config.GetGlobalConfigPath()
		if err != nil {
			logger.Errorf("Error getting config path: %v", err)
			os.Exit(1)
		}

		backupPath, err := config.BackupConfig(configPath)
		if err != nil {
			logger.Errorf("Error creating backup: %v", err)
			os.Exit(1)
		}

		logger.UserSuccessf("Configuration backup created: %s", backupPath)
	},
}

var configRestoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore configuration from a backup",
	Long:  "Restore the global configuration from a backup file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backupPath := args[0]
		configPath, err := config.GetGlobalConfigPath()
		if err != nil {
			logger.Errorf("Error getting config path: %v", err)
			os.Exit(1)
		}

		if err := config.RestoreConfig(backupPath, configPath); err != nil {
			logger.Errorf("Error restoring config: %v", err)
			os.Exit(1)
		}

		logger.UserSuccessf("Configuration restored from: %s", backupPath)
	},
}

var configListBackupsCmd = &cobra.Command{
	Use:   "list-backups",
	Short: "List all configuration backups",
	Long:  "List all available configuration backup files",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := config.GetGlobalConfigPath()
		if err != nil {
			logger.Errorf("Error getting config path: %v", err)
			os.Exit(1)
		}

		backups, err := config.ListBackups(configPath)
		if err != nil {
			logger.Errorf("Error listing backups: %v", err)
			os.Exit(1)
		}

		if len(backups) == 0 {
			logger.UserSuccess("No configuration backups found")
			return
		}

		logger.UserSuccessf("Found %d configuration backups:", len(backups))
		for i, backup := range backups {
			logger.UserInfof("%2d. %s", i+1, filepath.Base(backup))
		}
	},
}

var configInteractiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "Launch interactive configuration interface",
	Long:    "Launch an interactive TUI for managing configuration settings with a user-friendly interface",
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
	configCmd.AddCommand(configBackupCmd)
	configCmd.AddCommand(configRestoreCmd)
	configCmd.AddCommand(configListBackupsCmd)
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
			if v.Kind() == reflect.Map {
				// Handle map types
				if v.IsNil() {
					v.Set(reflect.MakeMap(v.Type()))
				}
				key := reflect.ValueOf(part)
				switch v.Type().Elem().Kind() {
				case reflect.String:
					v.SetMapIndex(key, reflect.ValueOf(value))
				case reflect.Bool:
					if boolVal, err := strconv.ParseBool(value); err == nil {
						v.SetMapIndex(key, reflect.ValueOf(boolVal))
					} else {
						return fmt.Errorf("invalid bool value: %s", value)
					}
				case reflect.Interface:
					// For interface{} maps, determine type from value
					if boolVal, err := strconv.ParseBool(value); err == nil {
						v.SetMapIndex(key, reflect.ValueOf(boolVal))
					} else if intVal, err := strconv.Atoi(value); err == nil {
						v.SetMapIndex(key, reflect.ValueOf(intVal))
					} else {
						v.SetMapIndex(key, reflect.ValueOf(value))
					}
				default:
					return fmt.Errorf("unsupported map value type: %s", v.Type().Elem().Kind())
				}
			} else {
				// Handle struct fields
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
			}
		} else {
			// Navigate to nested struct or map
			if v.Kind() == reflect.Map {
				// Handle map navigation
				key := reflect.ValueOf(part)
				mapValue := v.MapIndex(key)
				if !mapValue.IsValid() {
					// Create nested map or struct if it doesn't exist
					elemType := v.Type().Elem()
					if elemType.Kind() == reflect.Interface {
						// Create a new map[string]interface{}
						newMap := make(map[string]interface{})
						v.SetMapIndex(key, reflect.ValueOf(newMap))
						mapValue = v.MapIndex(key)
					} else {
						return fmt.Errorf("cannot create nested value for path: %s", strings.Join(parts[:i+1], "."))
					}
				}
				v = mapValue
				if v.Kind() == reflect.Interface {
					v = v.Elem()
				}
			} else {
				// Handle struct navigation
				field := v.FieldByName(toCamelCase(part))
				if !field.IsValid() {
					return fmt.Errorf("field %s not found", part)
				}
				v = field
			}
		}
	}

	return nil
}

// getFieldByPath gets a field value by dot notation path
func getFieldByPath(obj interface{}, path string) (string, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj).Elem()

	for _, part := range parts {
		if v.Kind() == reflect.Map {
			// Handle map navigation
			key := reflect.ValueOf(part)
			mapValue := v.MapIndex(key)
			if !mapValue.IsValid() {
				return "", fmt.Errorf("key %s not found in map", part)
			}
			v = mapValue
			if v.Kind() == reflect.Interface {
				v = v.Elem()
			}
		} else {
			// Handle struct navigation
			field := v.FieldByName(toCamelCase(part))
			if !field.IsValid() {
				return "", fmt.Errorf("field %s not found", part)
			}
			v = field
		}
	}

	return fmt.Sprintf("%v", v.Interface()), nil
}

// printConfig recursively prints all configuration values with improved formatting
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
			printConfigValue(fullPath, field.Interface())
		}
	}
}

// printConfigValue prints a single configuration value with improved formatting for complex types
func printConfigValue(path string, value interface{}) {
	// Add import for config package types
	switch v := value.(type) {
	case map[string]config.ServiceConfig:
		printServiceConfigMap(path, v)
	case config.ServiceConfig:
		printServiceConfig(path, v)
	case map[string]interface{}:
		printNestedMap(path, v, 0)
	case []interface{}:
		printSlice(path, v, 0)
	case []string:
		printStringSlice(path, v, 0)
	case *config.FallbackConfig:
		if v != nil {
			printFallbackConfig(path, *v)
		} else {
			logger.UserPrintf("%s = <nil>\n", path)
		}
	case *interface{}:
		if v != nil {
			printConfigValue(path, *v)
		} else {
			logger.UserPrintf("%s = <nil>\n", path)
		}
	default:
		// Handle sensitive information masking
		maskedValue := maskSensitiveValue(path, value)
		logger.UserPrintf("%s = %s\n", path, maskedValue)
	}
}

// printServiceConfigMap prints a map of service configurations
func printServiceConfigMap(basePath string, serviceMap map[string]config.ServiceConfig) {
	if len(serviceMap) == 0 {
		logger.UserPrintf("%s = {}\n", basePath)
		return
	}

	for serviceName, serviceConfig := range serviceMap {
		fullPath := basePath + "." + serviceName
		printServiceConfig(fullPath, serviceConfig)
	}
}

// printServiceConfig prints a single service configuration with proper structure
func printServiceConfig(path string, service config.ServiceConfig) {
	logger.UserPrintf("%s.enabled = %t\n", path, service.Enabled)
	logger.UserPrintf("%s.provider = %s\n", path, service.Provider)

	// Print config map
	if len(service.Config) > 0 {
		configPath := path + ".config"
		printNestedMap(configPath, service.Config, 0)
	} else {
		logger.UserPrintf("%s.config = {}\n", path)
	}

	// Print defaults map
	if len(service.Defaults) > 0 {
		defaultsPath := path + ".defaults"
		printNestedMap(defaultsPath, service.Defaults, 0)
	} else {
		logger.UserPrintf("%s.defaults = {}\n", path)
	}
}

// printFallbackConfig prints fallback configuration
func printFallbackConfig(path string, fallback config.FallbackConfig) {
	logger.UserPrintf("%s.enabled = %t\n", path, fallback.Enabled)
	logger.UserPrintf("%s.provider = %s\n", path, fallback.Provider)

	if len(fallback.Config) > 0 {
		configPath := path + ".config"
		printNestedMap(configPath, fallback.Config, 0)
	} else {
		logger.UserPrintf("%s.config = {}\n", path)
	}
}

// printNestedMap prints a map with proper indentation and structure
func printNestedMap(basePath string, m map[string]interface{}, depth int) {
	for key, value := range m {
		fullPath := basePath + "." + key

		switch v := value.(type) {
		case map[string]interface{}:
			printNestedMap(fullPath, v, depth+1)
		case []interface{}:
			printSlice(fullPath, v, depth+1)
		case []string:
			printStringSlice(fullPath, v, depth+1)
		default:
			maskedValue := maskSensitiveValue(fullPath, value)
			logger.UserPrintf("%s = %s\n", fullPath, maskedValue)
		}
	}
}

// printSlice prints a slice with proper formatting
func printSlice(path string, slice []interface{}, depth int) {
	if len(slice) == 0 {
		logger.UserPrintf("%s = []\n", path)
		return
	}

	logger.UserPrintf("%s:", path)
	indent := strings.Repeat("  ", depth+1)
	for i, item := range slice {
		switch v := item.(type) {
		case map[string]interface{}:
			logger.UserPrintf("%s[%d]:", indent, i)
			printNestedMap(path, v, depth+2)
		default:
			logger.UserPrintf("%s- %v", indent, item)
		}
	}
}

// printStringSlice prints a string slice with proper formatting
func printStringSlice(path string, slice []string, depth int) {
	if len(slice) == 0 {
		logger.UserPrintf("%s = []\n", path)
		return
	}

	logger.UserPrintf("%s:", path)
	indent := strings.Repeat("  ", depth+1)
	for _, item := range slice {
		logger.UserPrintf("%s- %s", indent, item)
	}
}

// maskSensitiveValue masks sensitive configuration values
func maskSensitiveValue(key string, value interface{}) string {
	sensitiveKeys := []string{"token", "password", "api_key", "secret", "auth_token", "bot_token"}

	for _, sensitiveKey := range sensitiveKeys {
		if strings.Contains(strings.ToLower(key), sensitiveKey) {
			if str, ok := value.(string); ok && len(str) > 4 {
				return str[:4] + "****" + str[len(str)-4:]
			}
			return "****"
		}
	}

	return fmt.Sprintf("%v", value)
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

	logger.UserSuccess("Interactive configuration completed successfully")
	return nil
}
