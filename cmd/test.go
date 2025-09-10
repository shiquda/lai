package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/notifier"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test notification channels",
	Long:  "Test configured notification channels by sending test messages",
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line parameters
		enabledNotifiers, _ := cmd.Flags().GetStringSlice("notifiers")
		customMessage, _ := cmd.Flags().GetString("message")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if err := runTestNotifications(enabledNotifiers, customMessage, verbose); err != nil {
			logger.Fatalf("Test failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Add command line parameters
	testCmd.Flags().StringSlice("notifiers", []string{}, "Test specific notifiers (comma-separated: telegram,email)")
	testCmd.Flags().String("message", "", "Custom test message")
	testCmd.Flags().BoolP("verbose", "v", false, "Show detailed test process")
}

func runTestNotifications(enabledNotifiers []string, customMessage string, verbose bool) error {
	if verbose {
		logger.Println("Starting notification channel test...")
		logger.Println("=====================================")
	}

	// Load global configuration
	cfg, err := config.BuildRuntimeConfig("", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create unified notifier for detailed service testing
	unifiedNotifier, err := notifier.CreateUnifiedNotifier(cfg)
	if err != nil {
		return fmt.Errorf("failed to create unified notifier: %w", err)
	}

	if !unifiedNotifier.IsEnabled() {
		return fmt.Errorf("no notification services enabled")
	}

	// Get enabled channels
	enabledChannels := unifiedNotifier.GetEnabledChannels()
	if len(enabledChannels) == 0 {
		return fmt.Errorf("no notification channels configured")
	}

	if verbose {
		logger.Printf("Found %d configured notification service(s)\n", len(enabledChannels))
		for _, channel := range enabledChannels {
			logger.Printf("  - %s\n", channel)
		}
		logger.Println()
	}

	// Prepare test message
	testMessage := customMessage
	if testMessage == "" {
		testMessage = getDefaultTestMessage()
	}

	// Test each enabled service
	var successCount, failureCount int
	for i, serviceName := range enabledChannels {
		if verbose {
			logger.Printf("Testing %s service (%d/%d)...\n", serviceName, i+1, len(enabledChannels))
		}

		if err := testSingleService(unifiedNotifier, serviceName, testMessage, verbose); err != nil {
			failureCount++
			logger.Printf("❌ %s service test failed: %v\n", serviceName, err)
		} else {
			successCount++
			logger.Printf("✅ %s service test succeeded\n", serviceName)
		}

		if verbose && i < len(enabledChannels)-1 {
			logger.Println()
		}
	}

	// Show summary
	logger.Println("=====================================")
	logger.Printf("Test completed: %d succeeded, %d failed\n", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("%d service(s) failed the test", failureCount)
	}

	return nil
}

func testSingleService(unifiedNotifier notifier.UnifiedNotifier, serviceName string, message string, verbose bool) error {
	if verbose {
		logger.Printf("  - Preparing test message for %s...\n", serviceName)
		logger.Printf("  - Message: %s\n", message)
	}

	// Send test message using TestProvider method
	ctx := context.Background()
	if err := unifiedNotifier.TestProvider(ctx, serviceName, message); err != nil {
		return fmt.Errorf("failed to send message via %s: %w", serviceName, err)
	}

	if verbose {
		logger.Printf("  - Message sent successfully via %s\n", serviceName)
	}

	return nil
}

func getDefaultTestMessage() string {
	return fmt.Sprintf("🧪 Lai Notification Test\n\nThis is a test message from Lai log monitoring tool.\n\nTime: %s\n\nIf you receive this message, your notification configuration is working correctly!", time.Now().Format("2006-01-02 15:04:05"))
}
