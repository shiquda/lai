package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/notifier"
	"github.com/shiquda/lai/internal/summarizer"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start log monitoring",
	Long:  "Start log monitoring service to monitor specified log files",
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("config")
		if configPath == "" {
			configPath = "config.yaml"
		}

		if err := runMonitor(configPath); err != nil {
			log.Fatalf("Monitor failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("config", "c", "", "Config file path (default: config.yaml)")
}

func runMonitor(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	openaiClient := summarizer.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.Model)
	telegramNotifier := notifier.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)

	logCollector := collector.New(cfg.LogFile, cfg.LineThreshold, cfg.CheckInterval)

	logCollector.SetTriggerHandler(func(newContent string) error {
		fmt.Printf("Log changes detected, generating summary...\n")

		summary, err := openaiClient.Summarize(newContent)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		if err := telegramNotifier.SendLogSummary(cfg.LogFile, summary); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		fmt.Printf("Summary sent to Telegram\n")
		return nil
	})

	fmt.Printf("Starting log monitoring: %s\n", cfg.LogFile)
	fmt.Printf("Line threshold: %d lines\n", cfg.LineThreshold)
	fmt.Printf("Check interval: %v\n", cfg.CheckInterval)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run collector in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- logCollector.Start()
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived stop signal, shutting down...")
		return nil
	case err := <-errChan:
		return err
	}
}