package collector

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type Collector struct {
	filePath      string
	lineThreshold int
	lastLineCount int
	checkInterval time.Duration
	onTrigger     func(newContent string) error
}

func New(filePath string, lineThreshold int, checkInterval time.Duration) *Collector {
	return &Collector{
		filePath:      filePath,
		lineThreshold: lineThreshold,
		checkInterval: checkInterval,
	}
}

func (c *Collector) SetTriggerHandler(handler func(newContent string) error) {
	c.onTrigger = handler
}

func (c *Collector) Start() error {
	if err := c.initLastLineCount(); err != nil {
		return fmt.Errorf("failed to initialize last line count: %w", err)
	}

	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.checkAndTrigger(); err != nil {
			fmt.Printf("Error checking file: %v\n", err)
		}
	}

	return nil
}

func (c *Collector) initLastLineCount() error {
	lineCount, err := c.countLines()
	if err != nil {
		if os.IsNotExist(err) {
			c.lastLineCount = 0
			return nil
		}
		return err
	}
	c.lastLineCount = lineCount
	return nil
}

func (c *Collector) checkAndTrigger() error {
	currentLineCount, err := c.countLines()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lineDiff := currentLineCount - c.lastLineCount

	if lineDiff >= c.lineThreshold {
		newContent, err := c.readNewLines(c.lastLineCount, currentLineCount)
		if err != nil {
			return fmt.Errorf("failed to read new lines: %w", err)
		}

		if c.onTrigger != nil {
			if err := c.onTrigger(newContent); err != nil {
				return fmt.Errorf("trigger handler failed: %w", err)
			}
		}

		c.lastLineCount = currentLineCount
	}

	return nil
}

func (c *Collector) countLines() (int, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	return lineCount, scanner.Err()
}

func (c *Collector) readNewLines(fromLine, toLine int) (string, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var newContent []string
	currentLine := 0

	for scanner.Scan() {
		if currentLine >= fromLine && currentLine < toLine {
			newContent = append(newContent, scanner.Text())
		}
		currentLine++
		if currentLine >= toLine {
			break
		}
	}

	result := ""
	for _, line := range newContent {
		result += line + "\n"
	}

	return result, scanner.Err()
}
