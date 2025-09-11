package cmd

import (
	"context"
	"fmt"
	"strings"
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

	// Validate and prepare the list of notifiers to test
	notifiersToTest, err := prepareNotifiersToTest(enabledNotifiers, enabledChannels, unifiedNotifier)
	if err != nil {
		return fmt.Errorf("failed to prepare notifiers for testing: %w", err)
	}

	if verbose {
		logger.Printf("Found %d configured notification service(s)\n", len(enabledChannels))
		for _, channel := range enabledChannels {
			logger.Printf("  - %s\n", channel)
		}
		
		if len(enabledNotifiers) > 0 {
			logger.Printf("Testing specified notifiers: %v\n", enabledNotifiers)
		}
		logger.Println()
	}

	// Prepare test message
	testMessage := customMessage
	if testMessage == "" {
		testMessage = getDefaultTestMessage()
	}

	// Test each selected service
	var successCount, failureCount int
	for i, serviceName := range notifiersToTest {
		if verbose {
			logger.Printf("Testing %s service (%d/%d)...\n", serviceName, i+1, len(notifiersToTest))
		}

		if err := testSingleService(unifiedNotifier, serviceName, testMessage, verbose); err != nil {
			failureCount++
			logger.Printf("❌ %s service test failed: %v\n", serviceName, err)
		} else {
			successCount++
			logger.Printf("✅ %s service test succeeded\n", serviceName)
		}

		if verbose && i < len(notifiersToTest)-1 {
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

// prepareNotifiersToTest prepares the list of notifiers to test based on user input and enabled channels
func prepareNotifiersToTest(userNotifiers []string, enabledChannels []string, unifiedNotifier notifier.UnifiedNotifier) ([]string, error) {
	// If user specified specific notifiers, validate and use them
	if len(userNotifiers) > 0 {
		// Create a map of enabled channels for case-insensitive lookup
		enabledChannelsMap := make(map[string]string) // lowercase -> original case
		for _, channel := range enabledChannels {
			enabledChannelsMap[strings.ToLower(channel)] = channel
		}

		// Validate user-specified notifiers and collect valid ones in original case
		var validNotifiers []string
		var invalidNotifiers []string
		
		for _, notifier := range userNotifiers {
			lowerNotifier := strings.ToLower(notifier)
			if originalCase, exists := enabledChannelsMap[lowerNotifier]; exists {
				validNotifiers = append(validNotifiers, originalCase)
			} else {
				invalidNotifiers = append(invalidNotifiers, notifier)
			}
		}

		// If there are invalid notifiers, provide detailed error
		if len(invalidNotifiers) > 0 {
			if len(validNotifiers) > 0 {
				return nil, fmt.Errorf("the following notifiers are not configured or enabled: %v\nAvailable notifiers: %v", invalidNotifiers, enabledChannels)
			} else {
				return nil, fmt.Errorf("the following notifiers are not configured or enabled: %v\nNo valid notifiers specified. Available notifiers: %v", invalidNotifiers, enabledChannels)
			}
		}

		// If all specified notifiers are invalid
		if len(validNotifiers) == 0 {
			return nil, fmt.Errorf("no valid notifiers specified. Available notifiers: %v", enabledChannels)
		}

		return validNotifiers, nil
	}

	// If no user notifiers specified, test all enabled channels
	return enabledChannels, nil
}

func getDefaultTestMessage() string {
	return fmt.Sprintf("🧪 Lai Notification Test\n\nThis is a test message from Lai log monitoring tool.\n\nTime: %s\n\nIf you receive this message, your notification configuration is working correctly!", time.Now().Format("2006-01-02 15:04:05"))
}
