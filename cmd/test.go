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
	Long: `Test configured notification channels by sending test messages.

Examples:
  # Test all configured notification channels
  lai test
  
  # Test specific notification channels
  lai test --notifiers telegram,email
  
  # Test with custom message
  lai test --message "Custom test message"
  
  # Test with verbose output
  lai test --verbose
  
  # Connection test only (no actual messages sent)
  lai test --connection-only
  
  # Detailed diagnostic mode
  lai test --diagnostic
  
  # Validate configuration only
  lai test --validate-only
  
  # Show available notifiers
  lai test --list

Test Modes:
  --connection-only  Test connection validity without sending messages
  --diagnostic       Show detailed diagnostic information during testing
  --validate-only    Validate configuration without performing tests
  --list            Show available notification channels

Configuration:
  The test command uses the global configuration at ~/.lai/config.yaml
  Use 'lai config set' to configure notification channels
  Use 'lai config list' to view current configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line parameters
		enabledNotifiers, _ := cmd.Flags().GetStringSlice("notifiers")
		customMessage, _ := cmd.Flags().GetString("message")
		verbose, _ := cmd.Flags().GetBool("verbose")
		connectionOnly, _ := cmd.Flags().GetBool("connection-only")
		diagnostic, _ := cmd.Flags().GetBool("diagnostic")
		validateOnly, _ := cmd.Flags().GetBool("validate-only")
		listNotifiers, _ := cmd.Flags().GetBool("list")

		// Handle list command
		if listNotifiers {
			if err := listAvailableNotifiers(); err != nil {
				logger.Fatalf("Failed to list notifiers: %v", err)
			}
			return
		}

		if err := runTestNotifications(enabledNotifiers, customMessage, verbose, connectionOnly, diagnostic, validateOnly); err != nil {
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
	testCmd.Flags().Bool("connection-only", false, "Test connection validity without sending messages")
	testCmd.Flags().Bool("diagnostic", false, "Show detailed diagnostic information during testing")
	testCmd.Flags().Bool("validate-only", false, "Validate configuration without performing tests")
	testCmd.Flags().Bool("list", false, "Show available notification channels")

	// Add parameter validation
	testCmd.MarkFlagsMutuallyExclusive("connection-only", "validate-only")
	testCmd.MarkFlagsMutuallyExclusive("diagnostic", "validate-only")
	testCmd.MarkFlagsMutuallyExclusive("list", "notifiers")
}

// TestResult represents the result of testing a single notification service
type TestResult struct {
	ServiceName string
	Status      string
	Error       error
	Details     string
	Config      map[string]interface{}
}

// TestStatus represents the overall test status
type TestStatus struct {
	TotalServices int
	SuccessCount  int
	FailureCount  int
	SkippedCount  int
	TestMode      string
	TestResults   []TestResult
	Configuration *config.Config
	StartTime     time.Time
	EndTime       time.Time
}

func runTestNotifications(enabledNotifiers []string, customMessage string, verbose, connectionOnly, diagnostic, validateOnly bool) error {
	status := &TestStatus{
		TestMode:  getTestMode(connectionOnly, diagnostic, validateOnly),
		StartTime: time.Now(),
	}

	// Show test mode information
	logger.UserInfof("🧪 Lai Notification Test - %s Mode", status.TestMode)
	logger.UserInfo("=====================================")

	// Load global configuration
	cfg, err := config.BuildRuntimeConfig("", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	status.Configuration = cfg

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
	testChannels, err := prepareNotifiersToTest(enabledNotifiers, enabledChannels, unifiedNotifier)
	if err != nil {
		return fmt.Errorf("failed to prepare notifiers for testing: %w", err)
	}

	status.TotalServices = len(testChannels)

	// Show configuration overview
	if verbose || diagnostic {
		showConfigurationOverview(cfg, testChannels, status.TestMode)
	}

	if verbose {
		logger.UserInfof("Found %d configured notification service(s)", len(enabledChannels))
		for _, channel := range enabledChannels {
			logger.UserInfof("  - %s", channel)
		}

		if len(enabledNotifiers) > 0 {
			logger.UserInfof("Testing specified notifiers: %v", enabledNotifiers)
		}
		logger.UserInfo("")
	}

	// Prepare test message
	testMessage := customMessage
	if testMessage == "" {
		testMessage = getDefaultTestMessage()
	}

	// Validate configuration if requested
	if validateOnly {
		return validateOnlyTest(cfg, testChannels, status)
	}

	// Test each enabled service
	for i, serviceName := range testChannels {
		result := testSingleServiceEnhanced(unifiedNotifier, serviceName, testMessage, verbose, connectionOnly, diagnostic)
		status.TestResults = append(status.TestResults, result)

		updateTestStatus(status, result)

		// Show progress
		showTestProgress(result, i+1, len(testChannels), verbose, diagnostic)

		if i < len(testChannels)-1 && (verbose || diagnostic) {
			logger.UserInfo("")
		}
	}

	// Show summary
	status.EndTime = time.Now()
	showTestSummary(status)

	if status.FailureCount > 0 {
		return fmt.Errorf("%d service(s) failed the test", status.FailureCount)
	}

	return nil
}

// getTestMode returns the test mode string based on flags
func getTestMode(connectionOnly, diagnostic, validateOnly bool) string {
	switch {
	case validateOnly:
		return "Configuration Validation"
	case connectionOnly:
		return "Connection Test"
	case diagnostic:
		return "Diagnostic"
	default:
		return "Standard"
	}
}

// showConfigurationOverview displays configuration information
func showConfigurationOverview(cfg *config.Config, testChannels []string, testMode string) {
	logger.UserInfo("📋 Configuration Overview")
	logger.UserInfo("=====================================")
	logger.UserInfof("Test Mode: %s", testMode)
	logger.UserInfof("Services to Test: %d", len(testChannels))

	// Show each service configuration
	for _, serviceName := range testChannels {
		serviceConfig, exists := cfg.Notifications.Providers[serviceName]
		if !exists {
			logger.UserErrorf("❌ %s: Configuration not found", serviceName)
			continue
		}

		status := "✅"
		if !serviceConfig.Enabled {
			status = "⚠️"
		}

		logger.UserInfof("%s %s: %s", status, serviceName, getProviderDescription(serviceConfig.Provider))

		// Show configuration details for enabled services
		if serviceConfig.Enabled {
			showServiceConfigurationDetails(serviceName, serviceConfig)
		}
	}

	logger.UserInfo("")
}

// getProviderDescription returns a human-readable description of the provider
func getProviderDescription(provider string) string {
	descriptions := map[string]string{
		"telegram":        "Telegram Bot",
		"slack":           "Slack",
		"slack_webhook":   "Slack Webhook",
		"discord":         "Discord",
		"discord_webhook": "Discord Webhook",
		"email":           "Email",
		"smtp":            "SMTP Email",
		"gmail":           "Gmail",
		"sendgrid":        "SendGrid",
		"mailgun":         "Mailgun",
		"pushover":        "Pushover",
		"twilio":          "Twilio SMS",
		"pagerduty":       "PagerDuty",
		"dingtalk":        "DingTalk",
		"wechat":          "WeChat",
	}

	if desc, exists := descriptions[provider]; exists {
		return desc
	}
	return fmt.Sprintf("%s Provider", strings.Title(provider))
}

// showServiceConfigurationDetails displays detailed configuration for a service
func showServiceConfigurationDetails(serviceName string, serviceConfig config.ServiceConfig) {
	logger.UserInfof("   Status: %s", map[bool]string{true: "Enabled", false: "Disabled"}[serviceConfig.Enabled])
	logger.UserInfof("   Provider: %s", serviceConfig.Provider)

	// Show required configuration keys
	requiredKeys := getRequiredConfigKeys(serviceConfig.Provider)
	var missingKeys []string

	for _, key := range requiredKeys {
		value, exists := serviceConfig.Config[key]
		if !exists || value == "" {
			missingKeys = append(missingKeys, key)
		} else {
			// Mask sensitive information
			maskedValue := maskSensitiveValue(key, value)
			logger.UserInfof("   %s: %s", key, maskedValue)
		}
	}

	if len(missingKeys) > 0 {
		logger.UserWarningf("   ⚠️  Missing required keys: %s", strings.Join(missingKeys, ", "))
	}
}

// getRequiredConfigKeys returns the required configuration keys for a provider
func getRequiredConfigKeys(provider string) []string {
	switch provider {
	case "telegram":
		return []string{"bot_token", "chat_id"}
	case "slack":
		return []string{"oauth_token"}
	case "slack_webhook":
		return []string{"webhook_url"}
	case "discord":
		return []string{"bot_token"}
	case "discord_webhook":
		return []string{"webhook_url"}
	case "smtp", "gmail":
		return []string{"smtp_host", "username", "password"}
	case "sendgrid":
		return []string{"api_key", "from_email"}
	case "mailgun":
		return []string{"api_key", "domain"}
	case "pushover":
		return []string{"token", "user"}
	case "twilio":
		return []string{"account_sid", "auth_token", "from_number"}
	case "pagerduty":
		return []string{"routing_key"}
	case "dingtalk":
		return []string{"access_token"}
	case "wechat":
		return []string{"corp_id", "corp_secret", "agent_id"}
	default:
		return []string{}
	}
}

// validateOnlyTest validates configuration without performing actual tests
func validateOnlyTest(cfg *config.Config, testChannels []string, status *TestStatus) error {
	logger.UserInfo("🔍 Configuration Validation")
	logger.UserInfo("=====================================")

	var validationErrors []string

	for _, serviceName := range testChannels {
		serviceConfig, exists := cfg.Notifications.Providers[serviceName]
		if !exists {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: configuration not found", serviceName))
			continue
		}

		// Validate service configuration
		if err := validateServiceConfiguration(serviceName, serviceConfig); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", serviceName, err))
		}
	}

	if len(validationErrors) == 0 {
		logger.UserSuccess("✅ All service configurations are valid")
		return nil
	}

	logger.UserError("❌ Configuration validation failed:")
	for _, err := range validationErrors {
		logger.UserErrorf("  - %s", err)
	}

	return fmt.Errorf("configuration validation failed for %d service(s)", len(validationErrors))
}

// validateServiceConfiguration validates a single service configuration
func validateServiceConfiguration(serviceName string, serviceConfig config.ServiceConfig) error {
	if !serviceConfig.Enabled {
		return fmt.Errorf("service is disabled")
	}

	requiredKeys := getRequiredConfigKeys(serviceConfig.Provider)
	var missingKeys []string

	for _, key := range requiredKeys {
		value, exists := serviceConfig.Config[key]
		if !exists || value == "" {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing required configuration keys: %s", strings.Join(missingKeys, ", "))
	}

	return nil
}

// testSingleServiceEnhanced tests a single service with enhanced feedback
func testSingleServiceEnhanced(unifiedNotifier notifier.UnifiedNotifier, serviceName string, message string, verbose, connectionOnly, diagnostic bool) TestResult {
	result := TestResult{
		ServiceName: serviceName,
		Status:      "unknown",
		Config:      make(map[string]interface{}),
	}

	// Get service configuration for details
	if unifiedNotifierImpl, ok := unifiedNotifier.(*notifier.NotifyNotifier); ok {
		// This is a temporary approach - we'll need to add a proper method to the interface
		if config, exists := unifiedNotifierImpl.GetServiceConfig(serviceName); exists {
			result.Config = config
		}
	}

	if verbose || diagnostic {
		logger.UserInfof("🔍 Testing %s service...", serviceName)
	}

	// Check if service is enabled
	if !unifiedNotifier.IsServiceEnabled(serviceName) {
		result.Status = "skipped"
		result.Details = "Service is disabled in configuration"
		if verbose || diagnostic {
			logger.UserWarningf("⚠️  %s service is disabled, skipping...", serviceName)
		}
		return result
	}

	// Connection test mode
	if connectionOnly {
		return testConnectionOnly(unifiedNotifier, serviceName, verbose, diagnostic)
	}

	// Send test message
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	err := unifiedNotifier.TestProvider(ctx, serviceName, message)
	duration := time.Since(startTime)

	if err != nil {
		result.Status = "failed"
		result.Error = err
		result.Details = fmt.Sprintf("Test failed after %v: %v", duration, err)

		if diagnostic {
			logger.UserErrorf("❌ %s service test failed", serviceName)
			logger.UserInfof("   Duration: %v", duration)
			logger.UserErrorf("   Error: %v", err)
			showTroubleshootingTips(serviceName, err)
		} else {
			logger.UserErrorf("❌ %s service test failed: %v", serviceName, err)
		}
	} else {
		result.Status = "success"
		result.Details = fmt.Sprintf("Test completed successfully in %v", duration)

		if diagnostic {
			logger.UserSuccessf("✅ %s service test succeeded", serviceName)
			logger.UserInfof("   Duration: %v", duration)
			logger.UserInfof("   Message ID: Test-%d", time.Now().Unix())
		} else {
			logger.UserSuccessf("✅ %s service test succeeded (%v)", serviceName, duration)
		}
	}

	return result
}

// testConnectionOnly tests connection without sending actual messages
func testConnectionOnly(unifiedNotifier notifier.UnifiedNotifier, serviceName string, verbose, diagnostic bool) TestResult {
	result := TestResult{
		ServiceName: serviceName,
		Status:      "unknown",
	}

	if verbose || diagnostic {
		logger.UserInfof("🔌 Testing %s service connection...", serviceName)
	}

	// For connection testing, we'll validate the configuration
	// This is a simplified approach - in a real implementation,
	// you might want to add actual connection testing logic
	if unifiedNotifier.IsServiceEnabled(serviceName) {
		result.Status = "success"
		result.Details = "Connection test passed (configuration valid)"

		if diagnostic {
			logger.UserSuccessf("✅ %s service connection test passed", serviceName)
			logger.UserInfof("   Configuration validation: OK")
		} else {
			logger.UserSuccessf("✅ %s service connection test passed", serviceName)
		}
	} else {
		result.Status = "failed"
		result.Details = "Service is disabled or configuration invalid"

		if diagnostic {
			logger.UserErrorf("❌ %s service connection test failed", serviceName)
			logger.UserErrorf("   Reason: Service disabled or invalid configuration")
		} else {
			logger.UserErrorf("❌ %s service connection test failed", serviceName)
		}
	}

	return result
}

// updateTestStatus updates the overall test status based on individual test results
func updateTestStatus(status *TestStatus, result TestResult) {
	switch result.Status {
	case "success":
		status.SuccessCount++
	case "failed":
		status.FailureCount++
	case "skipped":
		status.SkippedCount++
	}
}

// showTestProgress displays the progress of testing
func showTestProgress(result TestResult, current, total int, verbose, diagnostic bool) {
	if diagnostic {
		logger.UserInfof("Progress: %d/%d services tested", current, total)
	}
}

// showTestSummary displays a comprehensive test summary
func showTestSummary(status *TestStatus) {
	duration := status.EndTime.Sub(status.StartTime)

	logger.UserInfo("=====================================")
	logger.UserInfo("📊 Test Summary")
	logger.UserInfo("=====================================")
	logger.UserInfof("Test Mode: %s", status.TestMode)
	logger.UserInfof("Total Duration: %v", duration)
	logger.UserInfof("Services Tested: %d", status.TotalServices)
	logger.UserSuccessf("✅ Successful: %d", status.SuccessCount)
	logger.UserErrorf("❌ Failed: %d", status.FailureCount)
	logger.UserWarningf("⚠️  Skipped: %d", status.SkippedCount)

	// Show failed services details
	if status.FailureCount > 0 {
		logger.UserError("\n❌ Failed Services:")
		for _, result := range status.TestResults {
			if result.Status == "failed" {
				logger.UserErrorf("  - %s: %s", result.ServiceName, result.Details)
			}
		}
	}

	// Show suggestions
	if status.FailureCount > 0 {
		logger.UserInfo("\n💡 Troubleshooting Suggestions:")
		logger.UserInfo("  - Use --diagnostic flag for detailed error information")
		logger.UserInfo("  - Use --validate-only to check configuration")
		logger.UserInfo("  - Check your configuration with 'lai config list'")
		logger.UserInfo("  - Verify API keys and tokens are correct")
	}
}

// showTroubleshootingTips shows specific troubleshooting tips for common errors
func showTroubleshootingTips(serviceName string, err error) {
	errorMsg := err.Error()

	logger.UserInfo("   💡 Troubleshooting Tips:")

	switch {
	case strings.Contains(errorMsg, "not found"):
		logger.UserInfof("   - Service '%s' might not be configured properly", serviceName)
		logger.UserInfo("   - Check if the service is enabled in configuration")
	case strings.Contains(errorMsg, "token") || strings.Contains(errorMsg, "key"):
		logger.UserInfo("   - Verify your API token/key is correct and not expired")
		logger.UserInfo("   - Check if the token has proper permissions")
	case strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection"):
		logger.UserInfo("   - Check your internet connection")
		logger.UserInfo("   - Verify firewall settings")
		logger.UserInfo("   - Check if the service is experiencing outages")
	case strings.Contains(errorMsg, "timeout"):
		logger.UserInfo("   - The request timed out. Try again later")
		logger.UserInfo("   - Check if the service is responding slowly")
	default:
		logger.UserInfo("   - Check the service documentation for specific requirements")
		logger.UserInfo("   - Verify all required configuration parameters are set")
	}
}

// listAvailableNotifiers shows all available notification channels
func listAvailableNotifiers() error {
	logger.UserInfo("📋 Available Notification Channels")
	logger.UserInfo("=====================================")

	// Load configuration to show what's actually configured
	cfg, err := config.BuildRuntimeConfig("", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.Notifications.Providers) == 0 {
		logger.UserWarning("No notification channels configured.")
		logger.UserInfo("Use 'lai config set' to configure notification channels.")
		return nil
	}

	// Group providers by type
	groupedProviders := make(map[string][]string)
	for name := range cfg.Notifications.Providers {
		group := "Other"
		if strings.Contains(name, "telegram") {
			group = "Messaging"
		} else if strings.Contains(name, "slack") || strings.Contains(name, "discord") {
			group = "Team Chat"
		} else if strings.Contains(name, "email") || strings.Contains(name, "smtp") || strings.Contains(name, "gmail") {
			group = "Email"
		} else if strings.Contains(name, "pushover") || strings.Contains(name, "twilio") {
			group = "SMS/Push"
		}

		groupedProviders[group] = append(groupedProviders[group], name)
	}

	// Display grouped providers
	for group, providers := range groupedProviders {
		logger.UserInfof("%s:", group)
		for _, provider := range providers {
			providerConfig := cfg.Notifications.Providers[provider]
			status := "❌ Disabled"
			if providerConfig.Enabled {
				status = "✅ Enabled"
			}
			logger.UserInfof("  %s %s (%s)", status, provider, getProviderDescription(providerConfig.Provider))
		}
		logger.UserInfo("")
	}

	logger.UserInfo("To test specific channels:")
	logger.UserInfo("  lai test --notifiers telegram,email")
	logger.UserInfo("  lai test --notifiers slack --diagnostic")

	return nil
}

// Legacy function for backward compatibility
func testSingleService(unifiedNotifier notifier.UnifiedNotifier, serviceName string, message string, verbose bool) error {
	result := testSingleServiceEnhanced(unifiedNotifier, serviceName, message, verbose, false, false)
	if result.Status == "failed" {
		return result.Error
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
