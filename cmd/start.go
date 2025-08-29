package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/notifier"
	"github.com/shiquda/lai/internal/summarizer"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [log-file]",
	Short: "Start log monitoring",
	Long:  "Start log monitoring service for the specified log file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logFile := args[0]
		
		// Get command line parameters
		lineThreshold, _ := cmd.Flags().GetInt("line-threshold")
		intervalStr, _ := cmd.Flags().GetString("interval")
		chatID, _ := cmd.Flags().GetString("chat-id")
		
		// Parse time interval
		var checkInterval *time.Duration
		if intervalStr != "" {
			if duration, err := time.ParseDuration(intervalStr); err != nil {
				log.Fatalf("Invalid interval format: %v", err)
			} else {
				checkInterval = &duration
			}
		}
		
		// Handle line-threshold parameter
		var lineThresholdPtr *int
		if cmd.Flags().Changed("line-threshold") {
			lineThresholdPtr = &lineThreshold
		}
		
		// Handle chat-id parameter
		var chatIDPtr *string
		if cmd.Flags().Changed("chat-id") {
			chatIDPtr = &chatID
		}
		
		if err := runMonitor(logFile, lineThresholdPtr, checkInterval, chatIDPtr); err != nil {
			log.Fatalf("Monitor failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	
	// Add command line parameters
	startCmd.Flags().IntP("line-threshold", "l", 0, "Number of lines to trigger summary (overrides global config)")
	startCmd.Flags().StringP("interval", "i", "", "Check interval (e.g., 30s, 1m) (overrides global config)")
	startCmd.Flags().StringP("chat-id", "c", "", "Telegram chat ID (overrides global config)")
}

func runMonitor(logFile string, lineThreshold *int, checkInterval *time.Duration, chatID *string) error {
	// Build runtime configuration
	cfg, err := config.BuildRuntimeConfig(logFile, lineThreshold, checkInterval, chatID)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Create service instances
	openaiClient := summarizer.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.Model)
	telegramNotifier := notifier.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.ChatID)
	logCollector := collector.New(cfg.LogFile, cfg.LineThreshold, cfg.CheckInterval)

	// Set trigger handler
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

	// Display startup information
	fmt.Printf("Starting log monitoring: %s\n", cfg.LogFile)
	fmt.Printf("Line threshold: %d lines\n", cfg.LineThreshold)
	fmt.Printf("Check interval: %v\n", cfg.CheckInterval)
	fmt.Printf("Chat ID: %s\n", cfg.ChatID)

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