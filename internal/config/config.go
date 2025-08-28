package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogFile       string        `yaml:"log_file"`
	LineThreshold int           `yaml:"line_threshold"` // 行数
	CheckInterval time.Duration `yaml:"check_interval"`
	
	OpenAI OpenAIConfig `yaml:"openai"`
	Telegram TelegramConfig `yaml:"telegram"`
}

type OpenAIConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.LineThreshold == 0 {
		config.LineThreshold = 10 // 默认10行
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.LogFile == "" {
		return fmt.Errorf("log_file is required")
	}
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required")
	}
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram.bot_token is required")
	}
	if c.Telegram.ChatID == "" {
		return fmt.Errorf("telegram.chat_id is required")
	}
	return nil
}