package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	tempDir             string
	cleanup             func()
	originalHome        string
	originalUserProfile string
}

func (s *ConfigTestSuite) SetupSuite() {
	var cleanup func()
	s.tempDir, cleanup = testutils.CreateTempDir(s.T())
	s.cleanup = cleanup

	s.originalHome = os.Getenv("HOME")
	s.originalUserProfile = os.Getenv("USERPROFILE")

	// Set both HOME and USERPROFILE for cross-platform compatibility
	os.Setenv("HOME", s.tempDir)
	os.Setenv("USERPROFILE", s.tempDir)
}

func (s *ConfigTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
	os.Setenv("HOME", s.originalHome)
	if s.originalUserProfile != "" {
		os.Setenv("USERPROFILE", s.originalUserProfile)
	} else {
		os.Unsetenv("USERPROFILE")
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestGetGlobalConfigPath() {
	path, err := GetGlobalConfigPath()

	assert.NoError(s.T(), err)
	expected := filepath.Join(s.tempDir, ".lai", "config.yaml")
	assert.Equal(s.T(), expected, path)
}

func (s *ConfigTestSuite) TestLoadConfig_Valid() {
	testLogPath := testutils.GetTestLogPath("test.log")
	// For YAML content, escape backslashes for Windows paths
	yamlLogPath := strings.ReplaceAll(testLogPath, "\\", "\\\\")
	configPath := testutils.CreateFileWithContent(s.T(), s.tempDir, "config.yaml", `
log_file: "`+yamlLogPath+`"
line_threshold: 15
check_interval: "45s"

openai:
  api_key: "sk-test-123"
  base_url: "https://custom.openai.com/v1"
  model: "gpt-4"

telegram:
  bot_token: "123:token"
  chat_id: "-100123"
`)

	config, err := LoadConfig(configPath)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testLogPath, config.LogFile)
	assert.Equal(s.T(), 15, config.LineThreshold)
	assert.Equal(s.T(), 45*time.Second, config.CheckInterval)
	assert.Equal(s.T(), "sk-test-123", config.OpenAI.APIKey)
	assert.Equal(s.T(), "https://custom.openai.com/v1", config.OpenAI.BaseURL)
	assert.Equal(s.T(), "gpt-4", config.OpenAI.Model)
	assert.Equal(s.T(), "123:token", config.Telegram.BotToken)
	assert.Equal(s.T(), "-100123", config.Telegram.ChatID)
}

func (s *ConfigTestSuite) TestLoadConfig_WithDefaults() {
	configPath := testutils.CreateFileWithContent(s.T(), s.tempDir, "minimal_config.yaml", `
log_file: "`+testutils.GetTestLogPath("test.log")+`"

openai:
  api_key: "sk-test-123"

telegram:
  bot_token: "123:token"
  chat_id: "-100123"
`)

	config, err := LoadConfig(configPath)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 10, config.LineThreshold)             // default
	assert.Equal(s.T(), 30*time.Second, config.CheckInterval) // default
}

func (s *ConfigTestSuite) TestLoadConfig_InvalidYAML() {
	configPath := testutils.CreateFileWithContent(s.T(), s.tempDir, "invalid.yaml", `
log_file: "`+testutils.GetTestLogPath("test.log")+`"
invalid_yaml: [unclosed
`)

	config, err := LoadConfig(configPath)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), config)
	assert.Contains(s.T(), err.Error(), "failed to parse config yaml")
}

func (s *ConfigTestSuite) TestLoadConfig_FileNotExists() {
	config, err := LoadConfig("/nonexistent/config.yaml")

	assert.Error(s.T(), err)
	assert.Nil(s.T(), config)
}

func (s *ConfigTestSuite) TestConfigValidate_Valid() {
	config := &Config{
		LogFile: testutils.GetTestLogPath("test.log"),
		OpenAI: OpenAIConfig{
			APIKey: "sk-test-123",
		},
		Telegram: TelegramConfig{
			BotToken: "123:token",
		},
		ChatID: "-100123",
	}

	err := config.Validate()
	assert.NoError(s.T(), err)
}

func (s *ConfigTestSuite) TestConfigValidate_MissingLogFile() {
	config := &Config{
		OpenAI: OpenAIConfig{
			APIKey: "sk-test-123",
		},
		Telegram: TelegramConfig{
			BotToken: "123:token",
		},
		ChatID: "-100123",
	}

	err := config.Validate()
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "either log_file or command is required")
}

func (s *ConfigTestSuite) TestConfigValidate_MissingOpenAIKey() {
	config := &Config{
		LogFile: testutils.GetTestLogPath("test.log"),
		Telegram: TelegramConfig{
			BotToken: "123:token",
		},
		ChatID: "-100123",
	}

	err := config.Validate()
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "openai.api_key is required")
}

func (s *ConfigTestSuite) TestConfigValidate_MissingTelegramToken() {
	config := &Config{
		LogFile: testutils.GetTestLogPath("test.log"),
		OpenAI: OpenAIConfig{
			APIKey: "sk-test-123",
		},
		ChatID: "-100123",
	}

	err := config.Validate()
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "telegram.bot_token is required")
}

func (s *ConfigTestSuite) TestConfigValidate_MissingChatID() {
	config := &Config{
		LogFile: testutils.GetTestLogPath("test.log"),
		OpenAI: OpenAIConfig{
			APIKey: "sk-test-123",
		},
		Telegram: TelegramConfig{
			BotToken: "123:token",
		},
	}

	err := config.Validate()
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "chat_id is required")
}

func (s *ConfigTestSuite) TestLoadGlobalConfig_Default() {
	// Remove any existing global config file first
	configPath, _ := GetGlobalConfigPath()
	os.Remove(configPath)
	os.RemoveAll(filepath.Dir(configPath))

	config, err := LoadGlobalConfig()

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), config)
	assert.Equal(s.T(), "https://api.openai.com/v1", config.Notifications.OpenAI.BaseURL)
	assert.Equal(s.T(), "gpt-3.5-turbo", config.Notifications.OpenAI.Model)
	assert.Equal(s.T(), 10, config.Defaults.LineThreshold)
	assert.Equal(s.T(), 30*time.Second, config.Defaults.CheckInterval)
}

func (s *ConfigTestSuite) TestSaveAndLoadGlobalConfig() {
	config := &GlobalConfig{
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				APIKey:  "sk-global-test-123",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-4",
			},
			Telegram: TelegramConfig{
				BotToken: "global:token",
				ChatID:   "-100global",
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 20,
			CheckInterval: 60 * time.Second,
		},
	}

	err := SaveGlobalConfig(config)
	assert.NoError(s.T(), err)

	loadedConfig, err := LoadGlobalConfig()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), config.Notifications.OpenAI.APIKey, loadedConfig.Notifications.OpenAI.APIKey)
	assert.Equal(s.T(), config.Defaults.LineThreshold, loadedConfig.Defaults.LineThreshold)
	assert.Equal(s.T(), config.Defaults.CheckInterval, loadedConfig.Defaults.CheckInterval)
}

func (s *ConfigTestSuite) TestEnsureGlobalConfig() {
	err := EnsureGlobalConfig()
	assert.NoError(s.T(), err)

	configPath, err := GetGlobalConfigPath()
	assert.NoError(s.T(), err)

	_, err = os.Stat(configPath)
	assert.NoError(s.T(), err, "Global config file should be created")
}

func (s *ConfigTestSuite) TestBuildRuntimeConfig_Default() {
	os.MkdirAll(filepath.Join(s.tempDir, ".lai"), 0755)
	globalConfigPath := filepath.Join(s.tempDir, ".lai", "config.yaml")

	globalConfigContent := `
notifications:
  openai:
    api_key: "sk-global-123"
    base_url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
  telegram:
    bot_token: "global:token"
    chat_id: "-100global"

defaults:
  line_threshold: 15
  check_interval: "45s"
  chat_id: "-100global"
`
	testutils.CreateFileWithContent(s.T(), filepath.Dir(globalConfigPath), "config.yaml", globalConfigContent)

	testLogPath := testutils.GetTestLogPath("test.log")
	config, err := BuildRuntimeConfig(testLogPath, nil, nil, nil)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), testLogPath, config.LogFile)
	assert.Equal(s.T(), 15, config.LineThreshold)
	assert.Equal(s.T(), 45*time.Second, config.CheckInterval)
	assert.Equal(s.T(), "-100global", config.ChatID)
	assert.Equal(s.T(), "sk-global-123", config.OpenAI.APIKey)
	assert.Equal(s.T(), "global:token", config.Telegram.BotToken)
}

func (s *ConfigTestSuite) TestBuildRuntimeConfig_WithOverrides() {
	os.MkdirAll(filepath.Join(s.tempDir, ".lai"), 0755)
	globalConfigPath := filepath.Join(s.tempDir, ".lai", "config.yaml")

	globalConfigContent := `
notifications:
  openai:
    api_key: "sk-global-123"
  telegram:
    bot_token: "global:token"
    chat_id: "-100global"

defaults:
  line_threshold: 15
  check_interval: "45s"
  chat_id: "-100global"
`
	testutils.CreateFileWithContent(s.T(), filepath.Dir(globalConfigPath), "config.yaml", globalConfigContent)

	lineThreshold := 25
	checkInterval := 120 * time.Second
	testLogPath := testutils.GetTestLogPath("test.log")

	config, err := BuildRuntimeConfig(testLogPath, &lineThreshold, &checkInterval, nil)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 25, config.LineThreshold)
	assert.Equal(s.T(), 120*time.Second, config.CheckInterval)
	assert.Equal(s.T(), "-100global", config.ChatID)
}

func (s *ConfigTestSuite) TestApplyGlobalDefaults() {
	config := &GlobalConfig{}

	applyGlobalDefaults(config)

	assert.Equal(s.T(), "https://api.openai.com/v1", config.Notifications.OpenAI.BaseURL)
	assert.Equal(s.T(), "gpt-3.5-turbo", config.Notifications.OpenAI.Model)
	assert.Equal(s.T(), 10, config.Defaults.LineThreshold)
	assert.Equal(s.T(), 30*time.Second, config.Defaults.CheckInterval)
}

func (s *ConfigTestSuite) TestGetDefaultGlobalConfig() {
	config := getDefaultGlobalConfig()

	assert.NotNil(s.T(), config)
	assert.Equal(s.T(), "https://api.openai.com/v1", config.Notifications.OpenAI.BaseURL)
	assert.Equal(s.T(), "gpt-3.5-turbo", config.Notifications.OpenAI.Model)
	assert.Equal(s.T(), 10, config.Defaults.LineThreshold)
	assert.Equal(s.T(), 30*time.Second, config.Defaults.CheckInterval)
	assert.Equal(s.T(), true, config.Defaults.FinalSummary)
	assert.Equal(s.T(), false, config.Defaults.FinalSummaryOnly)
	assert.Equal(s.T(), "English", config.Defaults.Language)
}
