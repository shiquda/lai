package collector

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shiquda/lai/internal/logger"
)

// StreamCollector monitors command output streams
type StreamCollector struct {
	command       string
	args          []string
	lineThreshold int
	lineCount     int
	checkInterval time.Duration
	onTrigger     func(newContent string) error
	finalSummary  bool

	cmd       *exec.Cmd
	lines     []string
	lineMutex sync.RWMutex
	stopCh    chan struct{}
	running   bool
	runMutex  sync.RWMutex
	startTime time.Time
}

// NewStreamCollector creates a new stream collector for command output
func NewStreamCollector(command string, args []string, lineThreshold int, checkInterval time.Duration, finalSummary bool) *StreamCollector {
	return &StreamCollector{
		command:       command,
		args:          args,
		lineThreshold: lineThreshold,
		checkInterval: checkInterval,
		finalSummary:  finalSummary,
		lines:         make([]string, 0),
		stopCh:        make(chan struct{}),
	}
}

// SetTriggerHandler sets the callback function for when threshold is reached
func (sc *StreamCollector) SetTriggerHandler(handler func(newContent string) error) {
	sc.onTrigger = handler
}

// Start begins monitoring the command output
func (sc *StreamCollector) Start() error {
	sc.runMutex.Lock()
	if sc.running {
		sc.runMutex.Unlock()
		return fmt.Errorf("stream collector is already running")
	}
	sc.running = true
	sc.startTime = time.Now() // Record start time
	sc.runMutex.Unlock()

	defer func() {
		sc.runMutex.Lock()
		sc.running = false
		sc.runMutex.Unlock()
	}()

	// Start the command
	sc.cmd = exec.Command(sc.command, sc.args...)

	stdout, err := sc.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := sc.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := sc.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	logger.Printf("Started command: %s %s (PID: %d)\n", sc.command, strings.Join(sc.args, " "), sc.cmd.Process.Pid)

	// Create wait groups for goroutines
	var wg sync.WaitGroup

	// Monitor stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		sc.monitorStream(stdout, "stdout")
	}()

	// Monitor stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		sc.monitorStream(stderr, "stderr")
	}()

	// Start threshold checker
	wg.Add(1)
	go func() {
		defer wg.Done()
		sc.runThresholdChecker()
	}()

	// Wait for command to finish or stop signal
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- sc.cmd.Wait()
	}()

	var commandError error
	select {
	case <-sc.stopCh:
		// Stop signal received, kill the command
		if sc.cmd.Process != nil {
			sc.cmd.Process.Kill()
		}
		<-cmdDone // Wait for command to actually exit
		logger.Printf("Command stopped by user\n")
	case err := <-cmdDone:
		// Command finished - signal threshold checker to stop
		close(sc.stopCh)
		commandError = err
		if err != nil {
			logger.Printf("Command finished with error: %v\n", err)
		} else {
			logger.Printf("Command finished successfully\n")
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Send final summary if enabled
	if sc.finalSummary && sc.onTrigger != nil {
		sc.sendFinalSummary(commandError)
	}

	return nil
}

// Stop stops the stream collector
func (sc *StreamCollector) Stop() {
	sc.runMutex.RLock()
	running := sc.running
	sc.runMutex.RUnlock()

	if running {
		select {
		case <-sc.stopCh:
			// Channel already closed
		default:
			close(sc.stopCh)
		}
	}
}

// monitorStream reads from a stream and adds lines to the buffer
func (sc *StreamCollector) monitorStream(stream io.ReadCloser, streamType string) {
	scanner := bufio.NewScanner(stream)
	defer stream.Close()

	for scanner.Scan() {
		line := scanner.Text()

		sc.lineMutex.Lock()
		sc.lines = append(sc.lines, fmt.Sprintf("[%s] %s", streamType, line))
		sc.lineCount++
		sc.lineMutex.Unlock()

		// Print to console for immediate feedback
		logger.Printf("[%s] %s\n", streamType, line)
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading from %s: %v", streamType, err)
	}
}

// runThresholdChecker periodically checks if threshold is reached
func (sc *StreamCollector) runThresholdChecker() {
	ticker := time.NewTicker(sc.checkInterval)
	defer ticker.Stop()

	lastProcessedCount := 0

	for {
		select {
		case <-sc.stopCh:
			// Before exiting, check if there are any unprocessed lines
			sc.processRemainingLines(lastProcessedCount)
			return
		case <-ticker.C:
			sc.lineMutex.RLock()
			currentCount := sc.lineCount
			sc.lineMutex.RUnlock()

			newLines := currentCount - lastProcessedCount

			if newLines >= sc.lineThreshold {
				// Get the new content
				sc.lineMutex.RLock()
				var newContent strings.Builder
				for i := lastProcessedCount; i < currentCount && i < len(sc.lines); i++ {
					newContent.WriteString(sc.lines[i])
					newContent.WriteString("\n")
				}
				contentStr := newContent.String()
				sc.lineMutex.RUnlock()

				// Call the trigger handler
				if sc.onTrigger != nil && contentStr != "" {
					if err := sc.onTrigger(contentStr); err != nil {
						logger.Errorf("Error in trigger handler: %v", err)
					}
				}

				lastProcessedCount = currentCount
			}
		}
	}
}

// processRemainingLines processes any unprocessed lines when stopping
func (sc *StreamCollector) processRemainingLines(lastProcessedCount int) {
	sc.lineMutex.RLock()
	currentCount := sc.lineCount
	
	// If there are unprocessed lines and we meet the threshold
	if currentCount > lastProcessedCount && (currentCount-lastProcessedCount) >= sc.lineThreshold {
		var newContent strings.Builder
		for i := lastProcessedCount; i < currentCount && i < len(sc.lines); i++ {
			newContent.WriteString(sc.lines[i])
			newContent.WriteString("\n")
		}
		contentStr := newContent.String()
		sc.lineMutex.RUnlock()

		// Call the trigger handler one final time
		if sc.onTrigger != nil && contentStr != "" {
			if err := sc.onTrigger(contentStr); err != nil {
				logger.Errorf("Error in final trigger handler: %v", err)
			}
		}
	} else {
		sc.lineMutex.RUnlock()
	}
}

// GetLineCount returns current line count (thread-safe)
func (sc *StreamCollector) GetLineCount() int {
	sc.lineMutex.RLock()
	defer sc.lineMutex.RUnlock()
	return sc.lineCount
}

// GetLines returns all collected lines (thread-safe)
func (sc *StreamCollector) GetLines() []string {
	sc.lineMutex.RLock()
	defer sc.lineMutex.RUnlock()

	result := make([]string, len(sc.lines))
	copy(result, sc.lines)
	return result
}

// sendFinalSummary generates and sends a final summary when the command exits
func (sc *StreamCollector) sendFinalSummary(commandError error) {
	sc.lineMutex.RLock()
	totalLines := sc.lineCount
	allLines := make([]string, len(sc.lines))
	copy(allLines, sc.lines)
	sc.lineMutex.RUnlock()

	duration := time.Since(sc.startTime)

	// Build final summary content
	var summaryBuilder strings.Builder
	summaryBuilder.WriteString("=== PROGRAM EXIT SUMMARY ===\n")
	summaryBuilder.WriteString(fmt.Sprintf("Command: %s %s\n", sc.command, strings.Join(sc.args, " ")))
	summaryBuilder.WriteString(fmt.Sprintf("Duration: %v\n", duration.Round(time.Second)))
	summaryBuilder.WriteString(fmt.Sprintf("Total lines processed: %d\n", totalLines))

	if commandError != nil {
		summaryBuilder.WriteString(fmt.Sprintf("Exit status: ERROR - %v\n", commandError))
	} else {
		summaryBuilder.WriteString("Exit status: SUCCESS\n")
	}

	// Include recent log content (last 50 lines or all if less)
	recentLineCount := 50
	if totalLines > 0 {
		summaryBuilder.WriteString("\n=== RECENT LOG CONTENT ===\n")
		startIdx := 0
		if len(allLines) > recentLineCount {
			startIdx = len(allLines) - recentLineCount
			summaryBuilder.WriteString(fmt.Sprintf("... (showing last %d of %d lines)\n", recentLineCount, totalLines))
		}

		for i := startIdx; i < len(allLines); i++ {
			summaryBuilder.WriteString(allLines[i])
			summaryBuilder.WriteString("\n")
		}
	} else {
		summaryBuilder.WriteString("\n=== NO LOG CONTENT CAPTURED ===\n")
	}

	finalContent := summaryBuilder.String()

	// Send the final summary
	logger.Printf("Generating final summary...\n")
	if err := sc.onTrigger(finalContent); err != nil {
		logger.Errorf("Failed to send final summary: %v", err)
	} else {
		logger.Printf("Final summary sent successfully\n")
	}
}
