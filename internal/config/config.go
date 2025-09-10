package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"

	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/version"
)

// GlobalConfig represents the global configuration structure
type GlobalConfig struct {
	Version       string              `mapstructure:"version" yaml:"version"`
	Notifications NotificationsConfig `mapstructure:"notifications" yaml:"notifications"`
	Defaults      DefaultsConfig      `mapstructure:"defaults" yaml:"defaults"`
	Logging       LoggingConfig       `mapstructure:"logging" yaml:"logging"`
}

// NotificationsConfig contains notification channel configurations
type NotificationsConfig struct {
	OpenAI   OpenAIConfig   `mapstructure:"openai" yaml:"openai"`
	Telegram TelegramConfig `mapstructure:"telegram" yaml:"telegram"`
	Email    EmailConfig    `mapstructure:"email" yaml:"email"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

// DefaultsConfig contains default configuration values
type DefaultsConfig struct {
	LineThreshold    int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval    time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	FinalSummary     bool          `mapstructure:"final_summary" yaml:"final_summary"`
	FinalSummaryOnly bool          `mapstructure:"final_summary_only" yaml:"final_summary_only"`
	ErrorOnlyMode    bool          `mapstructure:"error_only_mode" yaml:"error_only_mode"`
	Language         string        `mapstructure:"language" yaml:"language"`
	EnabledNotifiers []string      `mapstructure:"enabled_notifiers" yaml:"enabled_notifiers"`
}

// Config represents the runtime configuration (merged final configuration)
type Config struct {
	LogFile          string        `mapstructure:"log_file" yaml:"log_file"`
	LineThreshold    int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval    time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	ChatID           string        `mapstructure:"chat_id" yaml:"chat_id"`
	Language         string        `mapstructure:"language" yaml:"language"`
	EnabledNotifiers []string      `mapstructure:"enabled_notifiers" yaml:"enabled_notifiers"`

	// Command execution parameters (for stream monitoring)
	Command     string   `mapstructure:"command" yaml:"command"`
	CommandArgs []string `mapstructure:"command_args" yaml:"command_args"`
	WorkingDir  string   `mapstructure:"working_dir" yaml:"working_dir"`

	// Exit handling options
	FinalSummary     bool `mapstructure:"final_summary" yaml:"final_summary"`
	FinalSummaryOnly bool `mapstructure:"final_summary_only" yaml:"final_summary_only"`

	// Error detection options
	ErrorOnlyMode bool `mapstructure:"error_only_mode" yaml:"error_only_mode"`

	OpenAI   OpenAIConfig   `mapstructure:"openai" yaml:"openai"`
	Telegram TelegramConfig `mapstructure:"telegram" yaml:"telegram"`
	Email    EmailConfig    `mapstructure:"email" yaml:"email"`
}

type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

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
		},
		Defaults: DefaultsConfig{
			LineThreshold: 10,
			CheckInterval: 30 * time.Second,
			FinalSummary:  true,      // Default to sending final summary
			Language:      "English", // Default language for AI responses
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

// GetDefaultGlobalConfig returns the default global configuration (exported)
func GetDefaultGlobalConfig() *GlobalConfig {
	return getDefaultGlobalConfig()
}

// applyGlobalDefaults applies global default values
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
	// Set FinalSummary to true by default
	// Since boolean false is the zero-value, we need to check if config was loaded from file
	// If no explicit setting exists, apply the default
	if !config.Defaults.FinalSummary {
		config.Defaults.FinalSummary = true
	}

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
		logger.Printf("Created default config file: %s\n", configPath)
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

	// If current config version is dev, always migrate
	if config.Version == "dev" {
		return true
	}

	// If default config version is dev, always migrate
	if defaultConfig.Version == "dev" {
		return true
	}

	// If current config version is less than default config version, migrate
	return compareVersions(config.Version, defaultConfig.Version) < 0
}

// migrateConfigSilently migrates the configuration without user interaction
func migrateConfigSilently(config *GlobalConfig, rawConfig map[string]interface{}) error {
	// Create backup of existing config
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102-150405")
	if err := copyFile(configPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Get default config
	defaultConfig := getDefaultGlobalConfig()

	// Use reflection to merge configurations elegantly
	mergedConfig := mergeConfigsWithReflection(config, defaultConfig)

	// Save merged config
	if err := SaveGlobalConfig(mergedConfig); err != nil {
		return fmt.Errorf("failed to save merged config: %w", err)
	}

	return nil
}

// mergeConfigsWithReflection merges two configurations using reflection
func mergeConfigsWithReflection(existing, defaults *GlobalConfig) *GlobalConfig {
	merged := *defaults

	// Use reflection to iterate through all fields
	existingValue := reflect.ValueOf(existing).Elem()
	defaultsValue := reflect.ValueOf(defaults).Elem()
	mergedValue := reflect.ValueOf(&merged).Elem()

	for i := 0; i < existingValue.NumField(); i++ {
		existingField := existingValue.Field(i)
		defaultsField := defaultsValue.Field(i)
		mergedField := mergedValue.Field(i)

		// Skip version field - always use default version
		if i == 0 { // Version is the first field
			continue
		}

		// For each field, merge based on type
		mergeField(existingField, defaultsField, mergedField)
	}

	return &merged
}

// mergeField recursively merges field values
func mergeField(existing, defaults, merged reflect.Value) {
	switch existing.Kind() {
	case reflect.Struct:
		// Recursively merge struct fields
		for i := 0; i < existing.NumField(); i++ {
			mergeField(existing.Field(i), defaults.Field(i), merged.Field(i))
		}
	case reflect.String:
		if existing.String() != "" {
			merged.SetString(existing.String())
		} else {
			merged.SetString(defaults.String())
		}
	case reflect.Int, reflect.Int64:
		if existing.Int() != 0 {
			merged.SetInt(existing.Int())
		} else {
			merged.SetInt(defaults.Int())
		}
	case reflect.Bool:
		// For bool, we need to check if it was explicitly set
		// This is tricky because false is a valid value
		// We'll use the existing value if it's different from the default
		if existing.Bool() != defaults.Bool() {
			merged.SetBool(existing.Bool())
		} else {
			merged.SetBool(defaults.Bool())
		}
	case reflect.Slice:
		if existing.Len() > 0 {
			merged.Set(existing)
		} else {
			merged.Set(defaults)
		}
	case reflect.Ptr:
		if !existing.IsNil() {
			merged.Set(existing)
		} else {
			merged.Set(defaults)
		}
	default:
		// For other types, use existing if not zero, otherwise use default
		if !existing.IsZero() {
			merged.Set(existing)
		} else {
			merged.Set(defaults)
		}
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
		LogFile:          logFile,
		LineThreshold:    globalConfig.Defaults.LineThreshold,
		CheckInterval:    globalConfig.Defaults.CheckInterval,
		ChatID:           globalConfig.Notifications.Telegram.ChatID,
		Language:         globalConfig.Defaults.Language,
		ErrorOnlyMode:    globalConfig.Defaults.ErrorOnlyMode,
		EnabledNotifiers: globalConfig.Defaults.EnabledNotifiers,
		OpenAI:           globalConfig.Notifications.OpenAI,
		Telegram:         globalConfig.Notifications.Telegram,
		Email:            globalConfig.Notifications.Email,
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
		ChatID:           globalConfig.Notifications.Telegram.ChatID,
		Language:         globalConfig.Defaults.Language,
		EnabledNotifiers: globalConfig.Defaults.EnabledNotifiers,
		FinalSummary:     globalConfig.Defaults.FinalSummary,
		FinalSummaryOnly: globalConfig.Defaults.FinalSummaryOnly,
		ErrorOnlyMode:    globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:           globalConfig.Notifications.OpenAI,
		Telegram:         globalConfig.Notifications.Telegram,
		Email:            globalConfig.Notifications.Email,
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

// GetTemplateMap converts TelegramMessageTemplates struct to a map for notifier
func (mt *TelegramMessageTemplates) GetTemplateMap() map[string]string {
	templates := make(map[string]string)
	if mt.LogSummary != "" {
		templates["log_summary"] = mt.LogSummary
	}
	return templates
}

// GetTemplateMap converts EmailMessageTemplates struct to a map for notifier
func (mt *EmailMessageTemplates) GetTemplateMap() map[string]string {
	templates := make(map[string]string)
	if mt.LogSummary != "" {
		templates["log_summary"] = mt.LogSummary
	}
	return templates
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
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram.bot_token is required")
	}
	if c.ChatID == "" {
		return fmt.Errorf("telegram.chat_id is required (set via notifications.telegram.chat_id in global config)")
	}
	return nil
}

// IsStreamMode returns true if this config is for stream monitoring
func (c *Config) IsStreamMode() bool {
	return c.Command != ""
}
