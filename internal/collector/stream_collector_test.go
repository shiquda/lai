package collector

import (
	"strings"
	"testing"
	"time"
)

func TestNewStreamCollector(t *testing.T) {
	command := "echo"
	args := []string{"hello", "world"}
	lineThreshold := 5
	checkInterval := 1 * time.Second
	finalSummary := true

	sc := NewStreamCollector(command, args, lineThreshold, checkInterval, finalSummary)

	if sc.command != command {
		t.Errorf("Expected command %s, got %s", command, sc.command)
	}
	if len(sc.args) != len(args) {
		t.Errorf("Expected %d args, got %d", len(args), len(sc.args))
	}
	if sc.lineThreshold != lineThreshold {
		t.Errorf("Expected lineThreshold %d, got %d", lineThreshold, sc.lineThreshold)
	}
	if sc.checkInterval != checkInterval {
		t.Errorf("Expected checkInterval %v, got %v", checkInterval, sc.checkInterval)
	}
	if sc.finalSummary != finalSummary {
		t.Errorf("Expected finalSummary %v, got %v", finalSummary, sc.finalSummary)
	}
}

func TestStreamCollectorSimpleCommand(t *testing.T) {
	// Test with a simple echo command
	sc := NewStreamCollector("echo", []string{"line1\nline2\nline3"}, 2, 100*time.Millisecond, false)

	sc.SetTriggerHandler(func(content string) error {
		// Just capture that trigger was called
		_ = content
		return nil
	})

	// Run the collector for a short time
	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	// Wait a bit for the command to execute and trigger
	time.Sleep(500 * time.Millisecond)

	// Stop the collector
	sc.Stop()

	// Wait for Start() to return
	<-done

	// Verify that lines were captured
	if sc.GetLineCount() == 0 {
		t.Error("Expected some lines to be captured")
	}

	lines := sc.GetLines()
	if len(lines) == 0 {
		t.Error("Expected some lines to be stored")
	}

	// Check that lines contain stdout prefix
	foundStdout := false
	for _, line := range lines {
		if strings.Contains(line, "[stdout]") {
			foundStdout = true
			break
		}
	}
	if !foundStdout {
		t.Error("Expected lines to contain [stdout] prefix")
	}
}

func TestStreamCollectorMultiLineCommand(t *testing.T) {
	// Test with a command that outputs multiple lines with delay to allow threshold checking
	sc := NewStreamCollector("sh", []string{"-c", "echo line1; sleep 0.1; echo line2; sleep 0.1; echo line3; sleep 0.1; echo line4; sleep 0.1; echo line5"}, 3, 50*time.Millisecond, false)

	triggerCount := 0
	sc.SetTriggerHandler(func(content string) error {
		triggerCount++
		return nil
	})

	// Run the collector
	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	// Wait for command to complete and give time for processing
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out - command didn't finish")
		sc.Stop()
		return
	}

	// Should have at least some lines
	if sc.GetLineCount() < 3 {
		t.Errorf("Expected at least 3 lines, got %d", sc.GetLineCount())
	}

	// Should have triggered at least once since we have 5 lines and threshold is 3
	if triggerCount == 0 {
		t.Error("Expected at least one trigger event")
	}
}

func TestStreamCollectorInvalidCommand(t *testing.T) {
	sc := NewStreamCollector("nonexistentcommand", []string{}, 1, 100*time.Millisecond, false)

	err := sc.Start()
	if err == nil {
		t.Error("Expected error when running non-existent command")
	}
}

func TestStreamCollectorStop(t *testing.T) {
	// Test with a long-running command
	sc := NewStreamCollector("sh", []string{"-c", "for i in $(seq 1 100); do echo line$i; sleep 0.1; done"}, 10, 100*time.Millisecond, false)

	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	// Stop it
	sc.Stop()

	// Should finish quickly after stop
	select {
	case <-done:
		// Good, it stopped
	case <-time.After(2 * time.Second):
		t.Error("Collector did not stop within reasonable time")
	}
}

func TestStreamCollectorTriggerHandler(t *testing.T) {
	// Use a slower command to allow threshold checking
	sc := NewStreamCollector("sh", []string{"-c", "echo test content; sleep 0.1"}, 1, 50*time.Millisecond, false)

	triggerCalled := false
	var receivedContent string

	sc.SetTriggerHandler(func(content string) error {
		triggerCalled = true
		receivedContent = content
		return nil
	})

	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	// Wait for the command to finish naturally
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out")
		sc.Stop()
		return
	}

	if !triggerCalled {
		t.Error("Expected trigger handler to be called")
	}

	if receivedContent == "" {
		t.Error("Expected non-empty content in trigger handler")
	}

	if !strings.Contains(receivedContent, "test content") {
		t.Errorf("Expected content to contain 'test content', got: %s", receivedContent)
	}
}

func TestStreamCollectorGetters(t *testing.T) {
	sc := NewStreamCollector("echo", []string{"test"}, 1, 100*time.Millisecond, false)

	// Initially should have no lines
	if sc.GetLineCount() != 0 {
		t.Errorf("Expected 0 lines initially, got %d", sc.GetLineCount())
	}

	if len(sc.GetLines()) != 0 {
		t.Errorf("Expected empty lines initially, got %d", len(sc.GetLines()))
	}

	// Run the command
	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	time.Sleep(200 * time.Millisecond)
	sc.Stop()
	<-done

	// Should have some lines now
	if sc.GetLineCount() == 0 {
		t.Error("Expected some lines after running command")
	}

	lines := sc.GetLines()
	if len(lines) != sc.GetLineCount() {
		t.Errorf("Lines slice length (%d) should match line count (%d)", len(lines), sc.GetLineCount())
	}
}

func TestStreamCollectorFinalSummary(t *testing.T) {
	sc := NewStreamCollector("echo", []string{"test final summary"}, 10, 50*time.Millisecond, true)

	finalSummaryCalled := false
	var finalContent string

	sc.SetTriggerHandler(func(content string) error {
		finalSummaryCalled = true
		finalContent = content
		return nil
	})

	// Run the command
	done := make(chan error, 1)
	go func() {
		done <- sc.Start()
	}()

	time.Sleep(200 * time.Millisecond)
	<-done

	// Should have called final summary
	if !finalSummaryCalled {
		t.Error("Expected final summary to be called")
	}

	// Check final summary content
	if !strings.Contains(finalContent, "PROGRAM EXIT SUMMARY") {
		t.Errorf("Expected final summary content, got: %s", finalContent)
	}

	if !strings.Contains(finalContent, "test final summary") {
		t.Errorf("Expected final summary to contain command output, got: %s", finalContent)
	}
}
