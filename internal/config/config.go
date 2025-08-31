package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// GlobalConfig represents the global configuration structure
type GlobalConfig struct {
	Notifications NotificationsConfig `mapstructure:"notifications" yaml:"notifications"`
	Defaults      DefaultsConfig      `mapstructure:"defaults" yaml:"defaults"`
}

// NotificationsConfig contains notification channel configurations
type NotificationsConfig struct {
	OpenAI   OpenAIConfig   `mapstructure:"openai" yaml:"openai"`
	Telegram TelegramConfig `mapstructure:"telegram" yaml:"telegram"`
}

// DefaultsConfig contains default configuration values
type DefaultsConfig struct {
	LineThreshold    int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval    time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	ChatID           string        `mapstructure:"chat_id" yaml:"chat_id"`
	FinalSummary     bool          `mapstructure:"final_summary" yaml:"final_summary"`
	FinalSummaryOnly bool          `mapstructure:"final_summary_only" yaml:"final_summary_only"`
	ErrorOnlyMode    bool          `mapstructure:"error_only_mode" yaml:"error_only_mode"`
	Language         string        `mapstructure:"language" yaml:"language"`
}

// Config represents the runtime configuration (merged final configuration)
type Config struct {
	LogFile       string        `mapstructure:"log_file" yaml:"log_file"`
	LineThreshold int           `mapstructure:"line_threshold" yaml:"line_threshold"`
	CheckInterval time.Duration `mapstructure:"check_interval" yaml:"check_interval"`
	ChatID        string        `mapstructure:"chat_id" yaml:"chat_id"`
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

	OpenAI   OpenAIConfig   `mapstructure:"openai" yaml:"openai"`
	Telegram TelegramConfig `mapstructure:"telegram" yaml:"telegram"`
}

type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

type TelegramConfig struct {
	BotToken         string           `mapstructure:"bot_token" yaml:"bot_token"`
	ChatID           string           `mapstructure:"chat_id" yaml:"chat_id"`
	MessageTemplates MessageTemplates `mapstructure:"message_templates" yaml:"message_templates"`
}

// MessageTemplates contains customizable message templates
type MessageTemplates struct {
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
	}
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
		fmt.Printf("Created default config file: %s\n", configPath)
	}

	// Validate existing config
	_, err = LoadGlobalConfig()
	return err
}

// BuildRuntimeConfig builds runtime configuration by merging global config and command line arguments
func BuildRuntimeConfig(logFile string, lineThreshold *int, checkInterval *time.Duration, chatID *string, errorOnlyMode *bool) (*Config, error) {
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
		LogFile:       logFile,
		LineThreshold: globalConfig.Defaults.LineThreshold,
		CheckInterval: globalConfig.Defaults.CheckInterval,
		ChatID:        globalConfig.Defaults.ChatID,
		Language:      globalConfig.Defaults.Language,
		ErrorOnlyMode: globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:        globalConfig.Notifications.OpenAI,
		Telegram:      globalConfig.Notifications.Telegram,
	}

	// Apply command line parameter overrides
	if lineThreshold != nil {
		config.LineThreshold = *lineThreshold
	}
	if checkInterval != nil {
		config.CheckInterval = *checkInterval
	}
	if chatID != nil {
		config.ChatID = *chatID
	}
	if errorOnlyMode != nil {
		config.ErrorOnlyMode = *errorOnlyMode
	}

	// If no ChatID specified, use the default one
	if config.ChatID == "" {
		config.ChatID = config.Telegram.ChatID
	}

	return config, nil
}

// BuildStreamConfig builds runtime configuration for stream monitoring
func BuildStreamConfig(command string, args []string, lineThreshold *int, checkInterval *time.Duration, chatID *string, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool) (*Config, error) {
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
		ChatID:           globalConfig.Defaults.ChatID,
		Language:         globalConfig.Defaults.Language,
		FinalSummary:     globalConfig.Defaults.FinalSummary,
		FinalSummaryOnly: globalConfig.Defaults.FinalSummaryOnly,
		ErrorOnlyMode:    globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:           globalConfig.Notifications.OpenAI,
		Telegram:         globalConfig.Notifications.Telegram,
	}

	// Apply command line parameter overrides
	if lineThreshold != nil {
		config.LineThreshold = *lineThreshold
	}
	if checkInterval != nil {
		config.CheckInterval = *checkInterval
	}
	if chatID != nil {
		config.ChatID = *chatID
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

	// If no ChatID specified, use the default one
	if config.ChatID == "" {
		config.ChatID = config.Telegram.ChatID
	}

	return config, nil
}

// GetTemplateMap converts MessageTemplates struct to a map for notifier
func (mt *MessageTemplates) GetTemplateMap() map[string]string {
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
		return fmt.Errorf("chat_id is required (set via --chat-id or defaults.chat_id in global config)")
	}
	return nil
}

// IsStreamMode returns true if this config is for stream monitoring
func (c *Config) IsStreamMode() bool {
	return c.Command != ""
}
