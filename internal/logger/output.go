package logger

import (
	"fmt"
	"io"
	"os"
)

// UserOutputLevel defines user output level types
type UserOutputLevel string

const (
	InfoOutput    UserOutputLevel = "info"
	SuccessOutput UserOutputLevel = "success"
	WarningOutput UserOutputLevel = "warning"
	ErrorOutput   UserOutputLevel = "error"
)

// UserOutput interface for user-facing output
type UserOutput interface {
	// Basic output methods
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Success(args ...interface{})
	Successf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Printf(format string, args ...interface{})

	}

// ConsoleOutput implements UserOutput for console output
type ConsoleOutput struct {
	writer io.Writer
	// Future extension: colorEnabled bool, colors map[UserOutputLevel]string
}

// NewConsoleOutput creates a new console output instance
func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{
		writer: os.Stdout,
	}
}

// NewConsoleOutputWithWriter creates a new console output instance with custom writer
func NewConsoleOutputWithWriter(writer io.Writer) *ConsoleOutput {
	return &ConsoleOutput{
		writer: writer,
	}
}

// Info outputs info-level user message
func (c *ConsoleOutput) Info(args ...interface{}) {
	fmt.Fprint(c.writer, args...)
}

// Infof outputs formatted info-level user message
func (c *ConsoleOutput) Infof(format string, args ...interface{}) {
	fmt.Fprintf(c.writer, format, args...)
}

// Success outputs success-level user message
func (c *ConsoleOutput) Success(args ...interface{}) {
	fmt.Fprint(c.writer, args...)
}

// Successf outputs formatted success-level user message
func (c *ConsoleOutput) Successf(format string, args ...interface{}) {
	fmt.Fprintf(c.writer, format, args...)
}

// Warning outputs warning-level user message
func (c *ConsoleOutput) Warning(args ...interface{}) {
	fmt.Fprint(c.writer, args...)
}

// Warningf outputs formatted warning-level user message
func (c *ConsoleOutput) Warningf(format string, args ...interface{}) {
	fmt.Fprintf(c.writer, format, args...)
}

// Error outputs error-level user message
func (c *ConsoleOutput) Error(args ...interface{}) {
	fmt.Fprint(c.writer, args...)
}

// Errorf outputs formatted error-level user message
func (c *ConsoleOutput) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(c.writer, format, args...)
}

// Printf outputs formatted text for special cases like command output
func (c *ConsoleOutput) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.writer, format, args...)
}


// Global user output instance
var defaultUserOutput UserOutput

// InitDefaultUserOutput initializes default user output instance
func InitDefaultUserOutput() {
	defaultUserOutput = NewConsoleOutput()
}

// GetDefaultUserOutput gets default user output instance
func GetDefaultUserOutput() UserOutput {
	if defaultUserOutput == nil {
		InitDefaultUserOutput()
	}
	return defaultUserOutput
}

// SetDefaultUserOutput sets custom user output instance for testing
func SetDefaultUserOutput(output UserOutput) {
	defaultUserOutput = output
}

// Global user output convenience functions
func UserInfo(args ...interface{}) {
	GetDefaultUserOutput().Info(args...)
}

func UserInfof(format string, args ...interface{}) {
	GetDefaultUserOutput().Infof(format, args...)
}

func UserSuccess(args ...interface{}) {
	GetDefaultUserOutput().Success(args...)
}

func UserSuccessf(format string, args ...interface{}) {
	GetDefaultUserOutput().Successf(format, args...)
}

func UserWarning(args ...interface{}) {
	GetDefaultUserOutput().Warning(args...)
}

func UserWarningf(format string, args ...interface{}) {
	GetDefaultUserOutput().Warningf(format, args...)
}

func UserError(args ...interface{}) {
	GetDefaultUserOutput().Error(args...)
}

func UserErrorf(format string, args ...interface{}) {
	GetDefaultUserOutput().Errorf(format, args...)
}

// UserPrintf outputs formatted text for special cases like config listing and command output
func UserPrintf(format string, args ...interface{}) {
	GetDefaultUserOutput().Printf(format, args...)
}