package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/shiquda/lai/internal/config"
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
			log.Fatalf("Test failed: %v", err)
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
		fmt.Println("Starting notification channel test...")
		fmt.Println("=====================================")
	}

	// Load global configuration
	cfg, err := config.BuildRuntimeConfig("", nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create notifiers
	notifiers, err := notifier.CreateNotifiers(cfg, enabledNotifiers)
	if err != nil {
		return fmt.Errorf("failed to create notifiers: %w", err)
	}

	if len(notifiers) == 0 {
		return fmt.Errorf("no valid notifiers configured")
	}

	if verbose {
		fmt.Printf("Found %d configured notifier(s)\n", len(notifiers))
		fmt.Println()
	}

	// Prepare test message
	testMessage := customMessage
	if testMessage == "" {
		testMessage = getDefaultTestMessage()
	}

	// Test each notifier
	var successCount, failureCount int
	for i, n := range notifiers {
		notifierType := getNotifierType(n)
		if verbose {
			fmt.Printf("Testing %s notifier (%d/%d)...\n", notifierType, i+1, len(notifiers))
		}

		if err := testSingleNotifier(n, notifierType, testMessage, verbose); err != nil {
			failureCount++
			fmt.Printf("❌ %s notification test failed: %v\n", notifierType, err)
		} else {
			successCount++
			fmt.Printf("✅ %s notification test succeeded\n", notifierType)
		}

		if verbose && i < len(notifiers)-1 {
			fmt.Println()
		}
	}

	// Show summary
	fmt.Println("=====================================")
	fmt.Printf("Test completed: %d succeeded, %d failed\n", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("%d notifier(s) failed the test", failureCount)
	}

	return nil
}

func testSingleNotifier(n notifier.Notifier, notifierType string, message string, verbose bool) error {
	if verbose {
		fmt.Printf("  - Preparing test message...\n")
		fmt.Printf("  - Message: %s\n", message)
	}

	// Send test message using SendMessage interface
	if err := n.SendMessage(message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if verbose {
		fmt.Printf("  - Message sent successfully\n")
	}

	return nil
}

func getNotifierType(n notifier.Notifier) string {
	// Use type assertion to determine notifier type
	switch n.(type) {
	case *notifier.TelegramNotifier:
		return "Telegram"
	case *notifier.EmailNotifier:
		return "Email"
	default:
		return "Unknown"
	}
}

func getDefaultTestMessage() string {
	return fmt.Sprintf("🧪 Lai Notification Test\n\nThis is a test message from Lai log monitoring tool.\n\nTime: %s\n\nIf you receive this message, your notification configuration is working correctly!", time.Now().Format("2006-01-02 15:04:05"))
}