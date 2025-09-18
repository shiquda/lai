package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"

	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/summarizer"
	"github.com/shiquda/lai/internal/version"
)

// GlobalConfig represents the global configuration structure
type GlobalConfig struct {
	Version         string                `mapstructure:"version" yaml:"version"`
	Notifications   NotificationsConfig   `mapstructure:"notifications" yaml:"notifications"`
	Defaults        DefaultsConfig        `mapstructure:"defaults" yaml:"defaults"`
	PromptTemplates PromptTemplatesConfig `mapstructure:"prompt_templates" yaml:"prompt_templates"`
	Logging         LoggingConfig         `mapstructure:"logging" yaml:"logging"`
	Display         DisplayConfig         `mapstructure:"display" yaml:"display"`
}

// NotificationsConfig contains the new unified notification configuration
type NotificationsConfig struct {
	OpenAI    OpenAIConfig             `mapstructure:"openai" yaml:"openai"`
	Providers map[string]ServiceConfig `mapstructure:"providers" yaml:"providers"`
	Fallback  *FallbackConfig          `mapstructure:"fallback" yaml:"fallback"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

// DisplayConfig contains display and output formatting configuration
type DisplayConfig struct {
	Colors ColorsConfig `mapstructure:"colors" yaml:"colors"`
}

// ColorsConfig contains color output configuration
type ColorsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Stdout  string `mapstructure:"stdout" yaml:"stdout"`
	Stderr  string `mapstructure:"stderr" yaml:"stderr"`
}

// DefaultsConfig contains default configuration values
type DefaultsConfig struct {
	LineThreshold    int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval    time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	FinalSummary     bool          `mapstructure:"final_summary" yaml:"final_summary"`
	FinalSummaryOnly bool          `mapstructure:"final_summary_only" yaml:"final_summary_only"`
	ErrorOnlyMode    bool          `mapstructure:"error_only_mode" yaml:"error_only_mode"`
	Language         string        `mapstructure:"language" yaml:"language"`
}

// PromptTemplatesConfig contains custom prompt templates for AI summarization
type PromptTemplatesConfig struct {
	// SummarizeTemplate is used for general log summarization
	SummarizeTemplate string `mapstructure:"summarize_template" yaml:"summarize_template"`

	// ErrorAnalysisTemplate is used for error detection and analysis
	ErrorAnalysisTemplate string `mapstructure:"error_analysis_template" yaml:"error_analysis_template"`

	// CustomVariables contains user-defined variables that can be used in templates
	CustomVariables map[string]string `mapstructure:"custom_variables" yaml:"custom_variables"`
}

// Config represents the runtime configuration (merged final configuration)
type Config struct {
	LogFile       string        `mapstructure:"log_file" yaml:"log_file"`
	LineThreshold int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	Language      string        `mapstructure:"language" yaml:"language"`

	// Command execution parameters (for stream monitoring)
	Command     string   `mapstructure:"command" yaml:"command"`
	CommandArgs []string `mapstructure:"command_args" yaml:"command_args"`
	WorkingDir  string   `mapstructure:"working_dir" yaml:"working_dir"`

	// Exit handling options
	FinalSummary     bool `mapstructure:"final_summary" yaml:"final_summary"`
	FinalSummaryOnly bool `mapstructure:"final_summary_only" yaml:"final_summary_only"`

	// Error detection options
	ErrorOnlyMode bool `mapstructure:"error_only_mode" yaml:"error_only_mode"`

	OpenAI          OpenAIConfig          `mapstructure:"openai" yaml:"openai"`
	Notifications   NotificationsConfig   `mapstructure:"notifications" yaml:"notifications"`
	PromptTemplates PromptTemplatesConfig `mapstructure:"prompt_templates" yaml:"prompt_templates"`
	Display         DisplayConfig         `mapstructure:"display" yaml:"display"`
}

type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

// ServiceConfig represents a single notification service configuration
type ServiceConfig struct {
	Enabled  bool                   `mapstructure:"enabled" yaml:"enabled"`
	Provider string                 `mapstructure:"provider" yaml:"provider"`
	Config   map[string]interface{} `mapstructure:"config" yaml:"config"`
	Defaults map[string]interface{} `mapstructure:"defaults" yaml:"defaults"`
}

// FallbackConfig represents fallback notification configuration
type FallbackConfig struct {
	Enabled  bool                   `mapstructure:"enabled" yaml:"enabled"`
	Provider string                 `mapstructure:"provider" yaml:"provider"`
	Config   map[string]interface{} `mapstructure:"config" yaml:"config"`
}

// Legacy configurations for backward compatibility during migration
type TelegramConfig struct {
	BotToken         string                   `mapstructure:"bot_token" yaml:"bot_token"`
	ChatID           string                   `mapstructure:"chat_id" yaml:"chat_id"`
	MessageTemplates TelegramMessageTemplates `mapstructure:"message_templates" yaml:"message_templates"`
}

type EmailConfig struct {
	SMTPHost         string                `mapstructure:"smtp_host" yaml:"smtp_host"`
	SMTPPort         int                   `mapstructure:"smtp_port" yaml:"smtp_port"`
	Username         string                `mapstructure:"username" yaml:"username"`
	Password         string                `mapstructure:"password" yaml:"password"`
	FromEmail        string                `mapstructure:"from_email" yaml:"from_email"`
	ToEmails         []string              `mapstructure:"to_emails" yaml:"to_emails"`
	Subject          string                `mapstructure:"subject" yaml:"subject"`
	UseTLS           bool                  `mapstructure:"use_tls" yaml:"use_tls"`
	MessageTemplates EmailMessageTemplates `mapstructure:"message_templates" yaml:"message_templates"`
}

type DiscordConfig struct {
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url"`
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
}

type SlackConfig struct {
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url"`
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
}

// Legacy message template types
type TelegramMessageTemplates struct {
	LogSummary string `mapstructure:"log_summary" yaml:"log_summary"`
}

type EmailMessageTemplates struct {
	LogSummary string `mapstructure:"log_summary" yaml:"log_summary"`
}

// GetGlobalConfigPath returns the path to the global configuration file
func GetGlobalConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".lai", "config.yaml"), nil
}

// LoadGlobalConfig loads the global configuration
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default configuration
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	// Parse to map first
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse global config yaml: %w", err)
	}

	// Use mapstructure to map to struct with custom decoder
	var config GlobalConfig
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
		Result: &config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(rawConfig); err != nil {
		return nil, fmt.Errorf("failed to decode global config: %w", err)
	}

	// Check if migration is needed and perform it silently
	if needsMigration(&config, rawConfig) {
		if err := migrateConfigSilently(&config, rawConfig); err != nil {
			return nil, fmt.Errorf("failed to migrate config: %w", err)
		}
	}

	// Apply default values
	applyGlobalDefaults(&config)
	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func SaveGlobalConfig(config *GlobalConfig) error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getDefaultGlobalConfig returns the default global configuration
func getDefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Version: version.Version,
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
			},
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  false, // Disabled by default until configured
					Provider: "telegram",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"parse_mode": "markdown",
					},
				},
				"email": {
					Enabled:  false, // Disabled by default until configured
					Provider: "smtp",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"subject": "Log Summary Notification",
					},
				},
				"slack": {
					Enabled:  false, // Disabled by default until configured
					Provider: "slack",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"username":   "Lai Bot",
						"icon_emoji": ":robot_face:",
					},
				},
				"discord": {
					Enabled:  false, // Disabled by default until configured
					Provider: "discord",
					Config:   map[string]interface{}{},
				},
				"discord_webhook": {
					Enabled:  false, // Disabled by default until configured
					Provider: "discord_webhook",
					Config:   map[string]interface{}{},
				},
			},
			Fallback: &FallbackConfig{
				Enabled:  false,
				Provider: "email",
				Config:   map[string]interface{}{},
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 10,
			CheckInterval: 30 * time.Second,
			FinalSummary:  true,      // Default to sending final summary
			Language:      "English", // Default language for AI responses
		},
		PromptTemplates: PromptTemplatesConfig{
			// Default summarize template (empty means use built-in template)
			SummarizeTemplate: "",
			// Default error analysis template (empty means use built-in template)
			ErrorAnalysisTemplate: "",
			// No custom variables by default
			CustomVariables: make(map[string]string),
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Display: DisplayConfig{
			Colors: ColorsConfig{
				Enabled: true,   // Enable colors by default
				Stdout:  "gray", // Gray for stdout
				Stderr:  "red",  // Red for stderr
			},
		},
	}
}

// GetDefaultGlobalConfig returns the default global configuration (exported)
func GetDefaultGlobalConfig() *GlobalConfig {
	return getDefaultGlobalConfig()
}

// applyGlobalDefaults applies global default values
// Only applies defaults for fields that are truly missing/empty
func applyGlobalDefaults(config *GlobalConfig) {
	if config.Notifications.OpenAI.BaseURL == "" {
		config.Notifications.OpenAI.BaseURL = "https://api.openai.com/v1"
	}
	if config.Notifications.OpenAI.Model == "" {
		config.Notifications.OpenAI.Model = "gpt-3.5-turbo"
	}
	if config.Defaults.LineThreshold == 0 {
		config.Defaults.LineThreshold = 10
	}
	if config.Defaults.CheckInterval == 0 {
		config.Defaults.CheckInterval = 30 * time.Second
	}
	if config.Defaults.Language == "" {
		config.Defaults.Language = "English"
	}

	// Note: We no longer force FinalSummary to true here
	// The migration logic should handle this properly

	// Apply logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
}

// LoadConfig loads legacy configuration file (backward compatible)
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse to map first
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config yaml: %w", err)
	}

	// Use mapstructure to map to struct with custom decoder
	var config Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
		Result: &config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(rawConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Set default values
	if config.LineThreshold == 0 {
		config.LineThreshold = 10 // Default 10 lines
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}

	return &config, nil
}

// EnsureGlobalConfig ensures the global config file exists and is valid
func EnsureGlobalConfig() error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := getDefaultGlobalConfig()
		if err := SaveGlobalConfig(defaultConfig); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		logger.Infof("Created default config file: %s", configPath)
	}

	// Validate existing config
	_, err = LoadGlobalConfig()
	return err
}

// InitLogger initializes the logging system with global config
func InitLogger() error {
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config for logger initialization: %w", err)
	}

	// Convert LoggingConfig to logger.LogConfig
	logConfig := &logger.LogConfig{
		Level: globalConfig.Logging.Level,
	}

	return logger.InitDefaultLogger(logConfig)
}

// getKeys returns the keys of a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// needsMigration checks if the configuration needs migration
func needsMigration(config *GlobalConfig, rawConfig map[string]interface{}) bool {
	defaultConfig := getDefaultGlobalConfig()

	// If config file doesn't have version field, always migrate
	if _, hasVersion := rawConfig["version"]; !hasVersion {
		return true
	}

	// Handle dev versions more intelligently
	// Only migrate if default version is not dev and is actually different
	if config.Version == "dev" {
		// If default version is also dev, no migration needed
		if defaultConfig.Version == "dev" {
			return false
		}
		// If default version is a real version, migrate from dev to real version
		return true
	}

	// If default config version is dev, but current is real version, no migration needed
	if defaultConfig.Version == "dev" {
		return false
	}

	// If current config version is less than default config version, migrate
	return compareVersions(config.Version, defaultConfig.Version) < 0
}

// migrateConfigSilently migrates the configuration without user interaction
func migrateConfigSilently(config *GlobalConfig, rawConfig map[string]interface{}) error {
	// Create backup of existing config using the new backup function
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	backupPath, err := BackupConfig(configPath)
	if err != nil {
		logger.Warnf("Failed to create backup: %v", err)
		// Continue with migration even if backup fails
	} else {
		logger.Infof("Created config backup: %s", backupPath)
	}

	// Get default config
	defaultConfig := getDefaultGlobalConfig()

	// Use intelligent configuration merging
	mergedConfig := smartMergeConfigs(config, defaultConfig)

	// Update the provided config pointer so callers immediately get the latest values
	*config = *mergedConfig

	// Save merged config
	if err := SaveGlobalConfig(config); err != nil {
		return fmt.Errorf("failed to save merged config: %w", err)
	}

	// Clean up old backups (keep only last 5)
	if err := CleanupOldBackups(configPath, 5); err != nil {
		logger.Warnf("Failed to clean up old backups: %v", err)
	}

	return nil
}

// MergeConfigsWithReflection merges two configurations using reflection
// Exported version for testing and external use
func MergeConfigsWithReflection(existing, defaults *GlobalConfig) *GlobalConfig {
	return mergeConfigsWithReflection(existing, defaults)
}

// mergeConfigsWithReflection merges two configurations using reflection
// Now uses existing config as base and only merges missing default fields
func mergeConfigsWithReflection(existing, defaults *GlobalConfig) *GlobalConfig {
	merged := *existing // Start with existing config as base

	// Use reflection to iterate through all fields
	existingValue := reflect.ValueOf(existing).Elem()
	defaultsValue := reflect.ValueOf(defaults).Elem()
	mergedValue := reflect.ValueOf(&merged).Elem()

	for i := 0; i < existingValue.NumField(); i++ {
		existingField := existingValue.Field(i)
		defaultsField := defaultsValue.Field(i)
		mergedField := mergedValue.Field(i)

		// Always update version to default version
		if i == 0 { // Version is the first field
			mergedField.SetString(defaultsField.String())
			continue
		}

		// For each field, merge based on type
		mergeFieldPreserveExisting(existingField, defaultsField, mergedField)
	}

	return &merged
}

// smartMergeConfigs performs intelligent configuration merging
// This is the new improved merge logic that handles complex nested structures properly
func smartMergeConfigs(existing, defaults *GlobalConfig) *GlobalConfig {
	merged := *existing // Start with existing config as base

	// Always update version
	merged.Version = defaults.Version

	// Smart merge for Notifications section
	merged.Notifications = smartMergeNotifications(existing.Notifications, defaults.Notifications)

	// Smart merge for Defaults section
	merged.Defaults = smartMergeDefaults(existing.Defaults, defaults.Defaults)

	// Smart merge for Logging section
	merged.Logging = smartMergeLogging(existing.Logging, defaults.Logging)

	// Smart merge for Display section
	merged.Display = smartMergeDisplay(existing.Display, defaults.Display)

	// Smart merge for PromptTemplates section
	merged.PromptTemplates = smartMergePromptTemplates(existing.PromptTemplates, defaults.PromptTemplates)

	return &merged
}

// smartMergeNotifications intelligently merges notification configurations
func smartMergeNotifications(existing, defaults NotificationsConfig) NotificationsConfig {
	merged := existing

	// Merge OpenAI config (preserve user settings, fill missing from defaults)
	merged.OpenAI = smartMergeOpenAI(existing.OpenAI, defaults.OpenAI)

	// Smart merge providers - this is the key fix!
	merged.Providers = smartMergeProviders(existing.Providers, defaults.Providers)

	// Merge fallback config
	merged.Fallback = smartMergeFallback(existing.Fallback, defaults.Fallback)

	return merged
}

// smartMergeProviders intelligently merges provider configurations
// This is the core fix for the configuration loss issue
func smartMergeProviders(existing, defaults map[string]ServiceConfig) map[string]ServiceConfig {
	if existing == nil {
		return defaults
	}

	merged := make(map[string]ServiceConfig)

	// First, copy all existing providers (preserve user configurations)
	for name, config := range existing {
		merged[name] = config
	}

	// Then, add missing providers from defaults (don't override existing ones)
	for name, defaultConfig := range defaults {
		if _, exists := merged[name]; !exists {
			merged[name] = defaultConfig
		}
	}

	return merged
}

// smartMergeOpenAI merges OpenAI configuration preserving user settings
func smartMergeOpenAI(existing, defaults OpenAIConfig) OpenAIConfig {
	merged := existing

	// Only fill missing values from defaults
	if merged.APIKey == "" {
		merged.APIKey = defaults.APIKey
	}
	if merged.BaseURL == "" {
		merged.BaseURL = defaults.BaseURL
	}
	if merged.Model == "" {
		merged.Model = defaults.Model
	}

	return merged
}

// smartMergeDefaults merges default configurations preserving user settings
func smartMergeDefaults(existing, defaults DefaultsConfig) DefaultsConfig {
	merged := existing

	// Only fill missing/zero values from defaults
	if merged.LineThreshold == 0 {
		merged.LineThreshold = defaults.LineThreshold
	}
	if merged.CheckInterval == 0 {
		merged.CheckInterval = defaults.CheckInterval
	}
	if merged.Language == "" {
		merged.Language = defaults.Language
	}

	// For boolean values, we need to be careful - false is a valid user choice
	// Only override if the existing value is the zero value (false for bools)
	// But we should handle this based on context - some bools should always preserve user choice

	return merged
}

// smartMergeLogging merges logging configurations
func smartMergeLogging(existing, defaults LoggingConfig) LoggingConfig {
	merged := existing

	if merged.Level == "" {
		merged.Level = defaults.Level
	}

	return merged
}

// smartMergeDisplay merges display configurations
func smartMergeDisplay(existing, defaults DisplayConfig) DisplayConfig {
	merged := existing

	// Merge colors config
	merged.Colors = smartMergeColors(existing.Colors, defaults.Colors)

	return merged
}

// smartMergeColors merges color configurations
func smartMergeColors(existing, defaults ColorsConfig) ColorsConfig {
	merged := existing

	// Preserve user's color choices, only fill missing
	if !merged.Enabled && defaults.Enabled {
		// If user disabled colors, respect that choice
		merged.Enabled = existing.Enabled
	}
	if merged.Stdout == "" {
		merged.Stdout = defaults.Stdout
	}
	if merged.Stderr == "" {
		merged.Stderr = defaults.Stderr
	}

	return merged
}

// smartMergeFallback merges fallback configurations
func smartMergeFallback(existing, defaults *FallbackConfig) *FallbackConfig {
	if existing == nil {
		return defaults
	}

	merged := *existing

	// Only fill missing values from defaults
	if !merged.Enabled && defaults != nil && defaults.Enabled {
		merged.Enabled = defaults.Enabled
	}
	if merged.Provider == "" && defaults != nil {
		merged.Provider = defaults.Provider
	}

	// Smart merge config maps
	if merged.Config == nil && defaults != nil {
		merged.Config = defaults.Config
	} else if defaults != nil && defaults.Config != nil {
		// Merge config maps, preserving existing keys
		for key, value := range defaults.Config {
			if _, exists := merged.Config[key]; !exists {
				if merged.Config == nil {
					merged.Config = make(map[string]interface{})
				}
				merged.Config[key] = value
			}
		}
	}

	return &merged
}

// smartMergePromptTemplates merges prompt template configurations
func smartMergePromptTemplates(existing, defaults PromptTemplatesConfig) PromptTemplatesConfig {
	merged := existing

	// Only fill missing template values from defaults
	// Empty strings mean "use built-in template", so we preserve that
	if merged.SummarizeTemplate == "" && defaults.SummarizeTemplate != "" {
		merged.SummarizeTemplate = defaults.SummarizeTemplate
	}
	if merged.ErrorAnalysisTemplate == "" && defaults.ErrorAnalysisTemplate != "" {
		merged.ErrorAnalysisTemplate = defaults.ErrorAnalysisTemplate
	}

	// Merge custom variables - preserve existing variables, add missing defaults
	if merged.CustomVariables == nil {
		merged.CustomVariables = make(map[string]string)
	}
	if defaults.CustomVariables != nil {
		for key, value := range defaults.CustomVariables {
			if _, exists := merged.CustomVariables[key]; !exists {
				merged.CustomVariables[key] = value
			}
		}
	}

	return merged
}

// mergeFieldPreserveExisting merges field values preserving existing configuration
// Only fills in missing fields from defaults
func mergeFieldPreserveExisting(existing, defaults, merged reflect.Value) {
	switch existing.Kind() {
	case reflect.Struct:
		// Recursively merge struct fields
		for i := 0; i < existing.NumField(); i++ {
			mergeFieldPreserveExisting(existing.Field(i), defaults.Field(i), merged.Field(i))
		}
	case reflect.String:
		// Only use default if existing is empty
		if existing.String() == "" {
			merged.SetString(defaults.String())
		}
		// Otherwise keep existing value (already set in merged)
	case reflect.Int, reflect.Int64:
		// Only use default if existing is 0
		if existing.Int() == 0 {
			merged.SetInt(defaults.Int())
		}
		// Otherwise keep existing value (already set in merged)
	case reflect.Bool:
		// For bool, the existing value should always be preserved
		// because false is a valid value that users might explicitly set
		// The merged value already contains the existing value, so we don't need to do anything
		// Only if we want to implement special logic for certain boolean fields, we should handle them individually
	case reflect.Slice:
		// Only use default if existing is empty
		if existing.Len() == 0 {
			merged.Set(defaults)
		}
		// Otherwise keep existing value (already set in merged)
	case reflect.Ptr:
		// Only use default if existing is nil
		if existing.IsNil() {
			merged.Set(defaults)
		}
		// Otherwise keep existing value (already set in merged)
	case reflect.Map:
		// Handle map merging specially
		if existing.IsNil() || existing.Len() == 0 {
			// If existing map is empty, use default map
			merged.Set(defaults)
		} else {
			// If existing map has values, keep it but merge missing keys from defaults
			if defaults.Len() > 0 {
				// Create a new map that combines existing and defaults
				mergedMap := reflect.MakeMap(existing.Type())

				// Copy all existing values
				iter := existing.MapRange()
				for iter.Next() {
					mergedMap.SetMapIndex(iter.Key(), iter.Value())
				}

				// Add default values for keys not in existing
				iter = defaults.MapRange()
				for iter.Next() {
					key := iter.Key()
					if !existing.MapIndex(key).IsValid() {
						mergedMap.SetMapIndex(key, iter.Value())
					}
				}

				merged.Set(mergedMap)
			}
			// Otherwise keep existing value (already set in merged)
		}
	default:
		// For other types, only use default if existing is zero
		if existing.IsZero() {
			merged.Set(defaults)
		}
		// Otherwise keep existing value (already set in merged)
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// BackupConfig creates a backup of the current configuration with version information
func BackupConfig(configPath string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Read current config to get version information
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config for backup: %w", err)
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return "", fmt.Errorf("failed to parse config for backup: %w", err)
	}

	// Get version from config or use "unknown"
	configVersion := "unknown"
	if version, exists := rawConfig["version"]; exists {
		if versionStr, ok := version.(string); ok {
			configVersion = versionStr
		}
	}

	// Create backup filename with timestamp and version
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("config.v%s.%s.backup.yaml", configVersion, timestamp)

	configDir := filepath.Dir(configPath)
	backupPath := filepath.Join(configDir, backupFilename)

	// Copy the file
	if err := copyFile(configPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// RestoreConfig restores configuration from a backup file
func RestoreConfig(backupPath, configPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Validate that backup is a valid config file
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var testConfig GlobalConfig
	if err := yaml.Unmarshal(data, &testConfig); err != nil {
		return fmt.Errorf("backup file is not a valid config: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Copy backup to config location
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}

// ListBackups lists all available backup files for a config
func ListBackups(configPath string) ([]string, error) {
	configDir := filepath.Dir(configPath)

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".backup.yaml") {
			backups = append(backups, filepath.Join(configDir, entry.Name()))
		}
	}

	// Sort backups by name (which includes timestamp)
	sort.Strings(backups)
	return backups, nil
}

// CleanupOldBackups removes old backup files, keeping only the most recent ones
func CleanupOldBackups(configPath string, keepCount int) error {
	backups, err := ListBackups(configPath)
	if err != nil {
		return err
	}

	if len(backups) <= keepCount {
		return nil // Nothing to clean up
	}

	// Remove oldest backups
	for i := 0; i < len(backups)-keepCount; i++ {
		if err := os.Remove(backups[i]); err != nil {
			logger.Warnf("Failed to remove old backup %s: %v", backups[i], err)
		}
	}

	return nil
}

// compareVersions compares two version strings (e.g., "1.0.0")
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	if v1 == "dev" {
		return -1 // Treat dev as older
	}
	if v2 == "dev" {
		return 1 // Treat dev as older
	}

	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int

		if i < len(v1Parts) {
			v1Num, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			v2Num, _ = strconv.Atoi(v2Parts[i])
		}

		if v1Num < v2Num {
			return -1
		} else if v1Num > v2Num {
			return 1
		}
	}

	return 0
}

// BuildRuntimeConfig builds runtime configuration by merging global config and command line arguments
func BuildRuntimeConfig(logFile string, lineThreshold *int, checkInterval *time.Duration, errorOnlyMode *bool) (*Config, error) {
	// Ensure global config exists
	if err := EnsureGlobalConfig(); err != nil {
		return nil, fmt.Errorf("failed to ensure global config: %w", err)
	}

	// Load global configuration
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Build runtime configuration
	config := &Config{
		LogFile:         logFile,
		LineThreshold:   globalConfig.Defaults.LineThreshold,
		CheckInterval:   globalConfig.Defaults.CheckInterval,
		Language:        globalConfig.Defaults.Language,
		ErrorOnlyMode:   globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:          globalConfig.Notifications.OpenAI,
		Notifications:   globalConfig.Notifications,
		PromptTemplates: globalConfig.PromptTemplates,
		Display:         globalConfig.Display,
	}

	// Apply command line parameter overrides
	if lineThreshold != nil {
		config.LineThreshold = *lineThreshold
	}
	if checkInterval != nil {
		config.CheckInterval = *checkInterval
	}
	if errorOnlyMode != nil {
		config.ErrorOnlyMode = *errorOnlyMode
	}

	return config, nil
}

// BuildStreamConfig builds runtime configuration for stream monitoring
func BuildStreamConfig(command string, args []string, lineThreshold *int, checkInterval *time.Duration, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool) (*Config, error) {
	// Ensure global config exists
	if err := EnsureGlobalConfig(); err != nil {
		return nil, fmt.Errorf("failed to ensure global config: %w", err)
	}

	// Load global configuration
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Build runtime configuration for stream monitoring
	config := &Config{
		Command:          command,
		CommandArgs:      args,
		WorkingDir:       workingDir,
		LineThreshold:    globalConfig.Defaults.LineThreshold,
		CheckInterval:    globalConfig.Defaults.CheckInterval,
		Language:         globalConfig.Defaults.Language,
		FinalSummary:     globalConfig.Defaults.FinalSummary,
		FinalSummaryOnly: globalConfig.Defaults.FinalSummaryOnly,
		ErrorOnlyMode:    globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:           globalConfig.Notifications.OpenAI,
		Notifications:    globalConfig.Notifications,
		PromptTemplates:  globalConfig.PromptTemplates,
		Display:          globalConfig.Display,
	}

	// Apply command line parameter overrides
	if lineThreshold != nil {
		config.LineThreshold = *lineThreshold
	}
	if checkInterval != nil {
		config.CheckInterval = *checkInterval
	}
	if finalSummary != nil {
		config.FinalSummary = *finalSummary
	}
	if finalSummaryOnly != nil {
		config.FinalSummaryOnly = *finalSummaryOnly
	}
	if errorOnlyMode != nil {
		config.ErrorOnlyMode = *errorOnlyMode
	}

	return config, nil
}

// MigrateToNewProviderConfig migrates legacy configuration to new provider format
// This function should be called when the old configuration format is detected
func MigrateToNewProviderConfig(oldConfig *GlobalConfig) *GlobalConfig {
	newConfig := *oldConfig

	// Initialize providers map if it doesn't exist
	if newConfig.Notifications.Providers == nil {
		newConfig.Notifications.Providers = make(map[string]ServiceConfig)
	}

	// Check if we have legacy configuration by looking for direct field access
	// This is a simplified migration that assumes the old config might have these fields

	// Try to migrate from mapstructure data if available
	if oldConfig.Notifications.OpenAI.APIKey == "" {
		// No OpenAI key configured, likely no legacy configuration to migrate
		return &newConfig
	}

	// Set up common providers with default configurations
	// These will be used if the user hasn't configured them yet

	// Default Telegram provider (disabled by default)
	if _, exists := newConfig.Notifications.Providers["telegram"]; !exists {
		newConfig.Notifications.Providers["telegram"] = ServiceConfig{
			Enabled:  false,
			Provider: "telegram",
			Config:   make(map[string]interface{}),
			Defaults: map[string]interface{}{
				"parse_mode": "markdown",
			},
		}
	}

	// Default Email provider (disabled by default)
	if _, exists := newConfig.Notifications.Providers["email"]; !exists {
		newConfig.Notifications.Providers["email"] = ServiceConfig{
			Enabled:  false,
			Provider: "smtp",
			Config:   make(map[string]interface{}),
			Defaults: map[string]interface{}{
				"subject": "🚨 Log Summary Notification",
			},
		}
	}

	// Default Slack provider (disabled by default)
	if _, exists := newConfig.Notifications.Providers["slack"]; !exists {
		newConfig.Notifications.Providers["slack"] = ServiceConfig{
			Enabled:  false,
			Provider: "slack",
			Config:   make(map[string]interface{}),
			Defaults: map[string]interface{}{
				"username":   "Lai Bot",
				"icon_emoji": ":robot_face:",
			},
		}
	}

	// Default Discord provider (disabled by default)
	if _, exists := newConfig.Notifications.Providers["discord"]; !exists {
		newConfig.Notifications.Providers["discord"] = ServiceConfig{
			Enabled:  false,
			Provider: "discord",
			Config:   make(map[string]interface{}),
			Defaults: map[string]interface{}{
				"username": "Lai Bot",
			},
		}
	}

	// Default Discord Webhook provider (disabled by default)
	if _, exists := newConfig.Notifications.Providers["discord_webhook"]; !exists {
		newConfig.Notifications.Providers["discord_webhook"] = ServiceConfig{
			Enabled:  false,
			Provider: "discord_webhook",
			Config:   make(map[string]interface{}),
			Defaults: map[string]interface{}{
				"username": "Lai Bot",
			},
		}
	}

	return &newConfig
}

func (c *Config) Validate() error {
	// Check if it's file monitoring or stream monitoring
	if c.LogFile == "" && c.Command == "" {
		return fmt.Errorf("either log_file or command is required")
	}

	if c.LogFile != "" && c.Command != "" {
		return fmt.Errorf("cannot specify both log_file and command")
	}

	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required")
	}

	// Check if at least one notification provider is configured
	if len(c.Notifications.Providers) == 0 {
		return fmt.Errorf("at least one notification provider must be configured")
	}

	// Validate that at least one provider is enabled
	var providerEnabled bool
	for _, provider := range c.Notifications.Providers {
		if provider.Enabled {
			providerEnabled = true
			break
		}
	}
	if !providerEnabled {
		return fmt.Errorf("at least one notification provider must be enabled")
	}

	// Validate prompt templates if they are provided
	if err := c.validatePromptTemplates(); err != nil {
		return fmt.Errorf("prompt template validation failed: %w", err)
	}

	return nil
}

// validatePromptTemplates validates the prompt templates configuration
func (c *Config) validatePromptTemplates() error {
	// Create a template engine for validation
	engine := summarizer.NewTemplateEngine()

	// Define allowed variables for templates
	allowedVariables := map[string]bool{
		"log_content": true,
		"language":    true,
		"system":      true,
		"version":     true,
		"timestamp":   true,
	}

	// Add custom variables to allowed list
	for key := range c.PromptTemplates.CustomVariables {
		allowedVariables[key] = true
	}

	// Validate summarize template if provided
	if c.PromptTemplates.SummarizeTemplate != "" {
		if err := engine.ValidateTemplate(c.PromptTemplates.SummarizeTemplate, allowedVariables); err != nil {
			return fmt.Errorf("summarize template validation failed: %w", err)
		}
	}

	// Validate error analysis template if provided
	if c.PromptTemplates.ErrorAnalysisTemplate != "" {
		if err := engine.ValidateTemplate(c.PromptTemplates.ErrorAnalysisTemplate, allowedVariables); err != nil {
			return fmt.Errorf("error analysis template validation failed: %w", err)
		}
	}

	return nil
}

// IsStreamMode returns true if this config is for stream monitoring
func (c *Config) IsStreamMode() bool {
	return c.Command != ""
}
