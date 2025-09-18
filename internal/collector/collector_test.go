package collector

import (
	"os"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CollectorTestSuite struct {
	suite.Suite
	tempDir string
	cleanup func()
}

func (s *CollectorTestSuite) SetupSuite() {
	var cleanup func()
	s.tempDir, cleanup = testutils.CreateTempDir(s.T())
	s.cleanup = cleanup
}

func (s *CollectorTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func TestCollectorSuite(t *testing.T) {
	suite.Run(t, new(CollectorTestSuite))
}

func TestNew(t *testing.T) {
	filePath := "/test/file.log"
	lineThreshold := 10
	checkInterval := 30 * time.Second

	collector := New(filePath, lineThreshold, checkInterval)

	assert.NotNil(t, collector)
	assert.Equal(t, filePath, collector.filePath)
	assert.Equal(t, lineThreshold, collector.lineThreshold)
	assert.Equal(t, checkInterval, collector.checkInterval)
	assert.Equal(t, 0, collector.lastLineCount)
	assert.Nil(t, collector.onTrigger)
	assert.NotNil(t, collector.stopCh)
}

func TestSetTriggerHandler(t *testing.T) {
	collector := New("test.log", 5, time.Second)
	handlerCalled := false

	handler := func(content string) error {
		handlerCalled = true
		return nil
	}

	collector.SetTriggerHandler(handler)
	assert.NotNil(t, collector.onTrigger)

	collector.onTrigger("test")
	assert.True(t, handlerCalled)
}

func (s *CollectorTestSuite) TestCountLines() {
	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{"empty file", "", 0},
		{"single line", "hello", 1},
		{"single line with newline", "hello\n", 1},
		{"multiple lines", "line1\nline2\nline3", 3},
		{"multiple lines with trailing newline", "line1\nline2\nline3\n", 3},
		{"lines with spaces", "  line 1  \n  line 2  ", 2},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", tt.content)
			collector := New(filePath, 5, time.Second)

			count, err := collector.countLines()

			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expectedCount, count)
		})
	}
}

func (s *CollectorTestSuite) TestCountLines_FileNotExist() {
	collector := New("/nonexistent/file.log", 5, time.Second)

	count, err := collector.countLines()

	assert.Error(s.T(), err)
	assert.True(s.T(), os.IsNotExist(err))
	assert.Equal(s.T(), 0, count)
}

func (s *CollectorTestSuite) TestReadNewLines() {
	content := "line1\nline2\nline3\nline4\nline5\n"
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", content)
	collector := New(filePath, 5, time.Second)

	tests := []struct {
		name     string
		fromLine int
		toLine   int
		expected string
	}{
		{"read first line", 0, 1, "line1\n"},
		{"read middle lines", 1, 3, "line2\nline3\n"},
		{"read last lines", 3, 5, "line4\nline5\n"},
		{"read all lines", 0, 5, "line1\nline2\nline3\nline4\nline5\n"},
		{"read beyond file", 3, 10, "line4\nline5\n"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := collector.readNewLines(tt.fromLine, tt.toLine)

			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expected, result)
		})
	}
}

func (s *CollectorTestSuite) TestReadNewLines_FileNotExist() {
	collector := New("/nonexistent/file.log", 5, time.Second)

	result, err := collector.readNewLines(0, 5)

	assert.Error(s.T(), err)
	assert.True(s.T(), os.IsNotExist(err))
	assert.Empty(s.T(), result)
}

func (s *CollectorTestSuite) TestInitLastLineCount() {
	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{"empty file", "", 0},
		{"file with content", "line1\nline2\nline3", 3},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", tt.content)
			collector := New(filePath, 5, time.Second)

			err := collector.initLastLineCount()

			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expectedCount, collector.lastLineCount)
		})
	}
}

func (s *CollectorTestSuite) TestInitLastLineCount_FileNotExist() {
	collector := New("/nonexistent/file.log", 5, time.Second)

	err := collector.initLastLineCount()

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, collector.lastLineCount)
}

func (s *CollectorTestSuite) TestCheckAndTrigger_NoChanges() {
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", "line1\nline2")
	collector := New(filePath, 5, time.Second)
	collector.initLastLineCount()

	handlerCalled := false
	collector.SetTriggerHandler(func(content string) error {
		handlerCalled = true
		return nil
	})

	err := collector.checkAndTrigger()

	assert.NoError(s.T(), err)
	assert.False(s.T(), handlerCalled)
}

func (s *CollectorTestSuite) TestCheckAndTrigger_BelowThreshold() {
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", "line1\nline2")
	collector := New(filePath, 5, time.Second)
	collector.initLastLineCount()

	testutils.AppendToFile(s.T(), filePath, "\nline3\nline4")

	handlerCalled := false
	collector.SetTriggerHandler(func(content string) error {
		handlerCalled = true
		return nil
	})

	err := collector.checkAndTrigger()

	assert.NoError(s.T(), err)
	assert.False(s.T(), handlerCalled)
}

func (s *CollectorTestSuite) TestCheckAndTrigger_ThresholdReached() {
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", "line1\nline2")
	collector := New(filePath, 3, time.Second)
	collector.initLastLineCount()

	testutils.AppendToFile(s.T(), filePath, "\nline3\nline4\nline5")

	var triggeredContent string
	handlerCalled := false
	collector.SetTriggerHandler(func(content string) error {
		handlerCalled = true
		triggeredContent = content
		return nil
	})

	err := collector.checkAndTrigger()

	assert.NoError(s.T(), err)
	assert.True(s.T(), handlerCalled)
	assert.Equal(s.T(), "line3\nline4\nline5\n", triggeredContent)
	assert.Equal(s.T(), 5, collector.lastLineCount)
}

func (s *CollectorTestSuite) TestCheckAndTrigger_HandlerError() {
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", "line1")
	collector := New(filePath, 1, time.Second)
	collector.initLastLineCount()

	testutils.AppendToFile(s.T(), filePath, "\nline2")

	expectedErr := assert.AnError
	collector.SetTriggerHandler(func(content string) error {
		return expectedErr
	})

	err := collector.checkAndTrigger()

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "trigger handler failed")
}

func (s *CollectorTestSuite) TestCheckAndTrigger_FileDisappears() {
	filePath := testutils.CreateFileWithContent(s.T(), s.tempDir, "test.log", "line1")
	collector := New(filePath, 1, time.Second)
	collector.initLastLineCount()

	os.Remove(filePath)

	err := collector.checkAndTrigger()

	assert.NoError(s.T(), err)
}

func TestCollector_Integration(t *testing.T) {
	tempDir, cleanup := testutils.CreateTempDir(t)
	defer cleanup()

	filePath := testutils.CreateFileWithContent(t, tempDir, "test.log", "initial content\n")

	collector := New(filePath, 2, 50*time.Millisecond)

	var receivedContent []string
	collector.SetTriggerHandler(func(content string) error {
		receivedContent = append(receivedContent, content)
		return nil
	})

	done := make(chan error, 1)
	go func() {
		done <- collector.Start()
	}()

	time.Sleep(10 * time.Millisecond)

	testutils.AppendToFile(t, filePath, "new line 1\nnew line 2\n")

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, receivedContent, 1)
	assert.Equal(t, "new line 1\nnew line 2\n", receivedContent[0])

	collector.Stop()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("collector failed to stop")
	}
}
