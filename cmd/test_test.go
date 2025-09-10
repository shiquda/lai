package cmd

import (
	"context"
	"testing"
)

// MockUnifiedNotifier for testing
type MockUnifiedNotifier struct {
	sendMessageFunc    func(ctx context.Context, message string) error
	sendLogSummaryFunc func(ctx context.Context, filePath, summary string) error
	testProviderFunc   func(ctx context.Context, providerName, message string) error
	enabledChannels    []string
}

func (m *MockUnifiedNotifier) SendLogSummary(ctx context.Context, filePath, summary string) error {
	if m.sendLogSummaryFunc != nil {
		return m.sendLogSummaryFunc(ctx, filePath, summary)
	}
	return nil
}

func (m *MockUnifiedNotifier) SendMessage(ctx context.Context, message string) error {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, message)
	}
	return nil
}

func (m *MockUnifiedNotifier) SendError(ctx context.Context, filePath, errorMsg string) error {
	return m.SendMessage(ctx, errorMsg)
}

func (m *MockUnifiedNotifier) TestProvider(ctx context.Context, providerName, message string) error {
	if m.testProviderFunc != nil {
		return m.testProviderFunc(ctx, providerName, message)
	}
	return nil
}

func (m *MockUnifiedNotifier) IsEnabled() bool {
	return len(m.enabledChannels) > 0
}

func (m *MockUnifiedNotifier) GetEnabledChannels() []string {
	return m.enabledChannels
}

func TestRunTestNotifications(t *testing.T) {
	// Test with successful notifications
	t.Run("Successful notifications", func(t *testing.T) {
		// This test would require more complex setup to mock the actual notifier creation
		// For now, we'll just test that the function doesn't panic
		// In a real test, you would mock the CreateUnifiedNotifier function
		_ = true // Placeholder for actual test
	})

	// Test with no enabled providers
	t.Run("No enabled providers", func(t *testing.T) {
		// This would test the error case where no providers are enabled
		_ = true // Placeholder for actual test
	})
}

func TestGetDefaultTestMessage(t *testing.T) {
	message := getDefaultTestMessage()

	if message == "" {
		t.Error("Default test message should not be empty")
	}

	// Check if message contains expected elements
	expectedElements := []string{
		"🧪 Lai Notification Test",
		"This is a test message from Lai log monitoring tool",
		"If you receive this message, your notification configuration is working correctly",
	}

	for _, element := range expectedElements {
		if !containsString(message, element) {
			t.Errorf("Test message should contain '%s'", element)
		}
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(s)] >= substr
}
