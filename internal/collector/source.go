package collector

import (
	"bufio"
	"os"
)

// MonitorType represents the type of monitoring source
type MonitorType string

const (
	MonitorTypeFile    MonitorType = "file"
	MonitorTypeCommand MonitorType = "command"
)

// MonitorSource represents a source that can be monitored for changes
type MonitorSource interface {
	GetType() MonitorType
	GetIdentifier() string
	GetContent() (string, error)
	GetLineCount() (int, error)
	ReadNewLines(from, to int) (string, error)
	GetLastPosition() int64
	SetLastPosition(pos int64)
	Start() error
	Stop()
}

// FileSource represents a file-based monitoring source
type FileSource struct {
	filePath     string
	lastPosition int64
}

func NewFileSource(filePath string) *FileSource {
	return &FileSource{
		filePath:     filePath,
		lastPosition: 0,
	}
}

func (f *FileSource) GetType() MonitorType {
	return MonitorTypeFile
}

func (f *FileSource) GetIdentifier() string {
	return f.filePath
}

func (f *FileSource) GetContent() (string, error) {
	file, err := os.Open(f.filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content []string
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	result := ""
	for _, line := range content {
		result += line + "\n"
	}

	return result, scanner.Err()
}

func (f *FileSource) GetLineCount() (int, error) {
	file, err := os.Open(f.filePath)
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

func (f *FileSource) ReadNewLines(fromLine, toLine int) (string, error) {
	file, err := os.Open(f.filePath)
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

func (f *FileSource) GetLastPosition() int64 {
	return f.lastPosition
}

func (f *FileSource) SetLastPosition(pos int64) {
	f.lastPosition = pos
}

func (f *FileSource) Start() error {
	// File source doesn't need explicit start
	return nil
}

func (f *FileSource) Stop() {
	// File source doesn't need explicit stop
}
