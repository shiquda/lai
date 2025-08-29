package summarizer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SummarizerTestSuite struct {
	suite.Suite
	server *httptest.Server
	client *OpenAIClient
}

func (s *SummarizerTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(s.mockOpenAIHandler))

	s.client = NewOpenAIClient("test-api-key", s.server.URL, "gpt-4o")
	s.client.client = s.server.Client()
}

func (s *SummarizerTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *SummarizerTestSuite) mockOpenAIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if strings.Contains(authHeader, "error-key") {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Model == "" || len(req.Messages) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if req.Model == "error-model" {
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	// Generate mock summary based on log content
	logContent := ""
	if len(req.Messages) > 0 {
		logContent = req.Messages[0].Content
	}

	var summary string
	if strings.Contains(logContent, "ERROR") {
		summary = "Log Analysis Summary:\n- Found 1 ERROR event that needs attention\n- System appears to have encountered an issue\n- Recommend immediate investigation"
	} else if strings.Contains(logContent, "INFO") {
		summary = "Log Analysis Summary:\n- Normal operation detected\n- 2 INFO events logged\n- No immediate action required"
	} else {
		summary = "Log Analysis Summary:\n- Analyzed log content\n- No critical issues found\n- System operating normally"
	}

	response := ChatCompletionResponse{
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: summary,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func TestSummarizerSuite(t *testing.T) {
	suite.Run(t, new(SummarizerTestSuite))
}

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		baseURL       string
		model         string
		expectedURL   string
		expectedModel string
	}{
		{
			name:          "with all parameters",
			apiKey:        "sk-test-123",
			baseURL:       "https://custom.openai.com/v1",
			model:         "gpt-4",
			expectedURL:   "https://custom.openai.com/v1",
			expectedModel: "gpt-4",
		},
		{
			name:          "with empty baseURL",
			apiKey:        "sk-test-123",
			baseURL:       "",
			model:         "gpt-3.5-turbo",
			expectedURL:   "https://api.openai.com/v1",
			expectedModel: "gpt-3.5-turbo",
		},
		{
			name:          "with empty model",
			apiKey:        "sk-test-123",
			baseURL:       "https://custom.openai.com/v1",
			model:         "",
			expectedURL:   "https://custom.openai.com/v1",
			expectedModel: "gpt-4o",
		},
		{
			name:          "with all defaults",
			apiKey:        "sk-test-123",
			baseURL:       "",
			model:         "",
			expectedURL:   "https://api.openai.com/v1",
			expectedModel: "gpt-4o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL, tt.model)

			assert.Equal(t, tt.apiKey, client.apiKey)
			assert.Equal(t, tt.expectedURL, client.baseURL)
			assert.Equal(t, tt.expectedModel, client.model)
			assert.NotNil(t, client.client)
		})
	}
}

func (s *SummarizerTestSuite) TestSummarize_Success() {
	logContent := `2024-01-15 10:30:00 INFO Application started successfully
2024-01-15 10:30:01 INFO Database connection established
2024-01-15 10:30:02 INFO Ready to serve requests`

	summary, err := s.client.Summarize(logContent, "English")

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), summary)
	assert.Contains(s.T(), summary, "Log Analysis Summary")
	assert.Contains(s.T(), summary, "INFO events")
}

func (s *SummarizerTestSuite) TestSummarize_WithErrors() {
	logContent := `2024-01-15 10:30:00 INFO Application started
2024-01-15 10:30:01 ERROR Failed to connect to database
2024-01-15 10:30:02 INFO Retrying connection`

	summary, err := s.client.Summarize(logContent, "English")

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), summary)
	assert.Contains(s.T(), summary, "ERROR event")
	assert.Contains(s.T(), summary, "attention")
}

func (s *SummarizerTestSuite) TestSummarize_UnauthorizedAPI() {
	client := NewOpenAIClient("error-key", s.server.URL, "gpt-4o")
	client.client = s.server.Client()

	summary, err := client.Summarize("test log content", "English")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "API request failed")
	assert.Empty(s.T(), summary)
}

func (s *SummarizerTestSuite) TestSummarize_InvalidModel() {
	client := NewOpenAIClient("test-api-key", s.server.URL, "error-model")
	client.client = s.server.Client()

	summary, err := client.Summarize("test log content", "English")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "API request failed")
	assert.Empty(s.T(), summary)
}

func (s *SummarizerTestSuite) TestSummarize_NetworkError() {
	// Create client with invalid URL to simulate network error
	client := NewOpenAIClient("test-api-key", "http://invalid-url-that-does-not-exist", "gpt-4o")

	summary, err := client.Summarize("test log content", "English")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to send request")
	assert.Empty(s.T(), summary)
}

func (s *SummarizerTestSuite) TestChatCompletionRequest_Marshall() {
	req := ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Test message",
			},
		},
	}

	jsonData, err := json.Marshal(req)
	assert.NoError(s.T(), err)

	var unmarshaled ChatCompletionRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), req.Model, unmarshaled.Model)
	assert.Len(s.T(), unmarshaled.Messages, 1)
	assert.Equal(s.T(), req.Messages[0].Role, unmarshaled.Messages[0].Role)
	assert.Equal(s.T(), req.Messages[0].Content, unmarshaled.Messages[0].Content)
}

func (s *SummarizerTestSuite) TestChatCompletionResponse_Unmarshall() {
	responseJSON := `{
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "This is a test response"
				}
			}
		]
	}`

	var response ChatCompletionResponse
	err := json.Unmarshal([]byte(responseJSON), &response)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), response.Choices, 1)
	assert.Equal(s.T(), "assistant", response.Choices[0].Message.Role)
	assert.Equal(s.T(), "This is a test response", response.Choices[0].Message.Content)
}

func (s *SummarizerTestSuite) TestSummarize_EmptyResponse() {
	// Create a server that returns empty choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatCompletionResponse{
			Choices: []Choice{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-api-key", server.URL, "gpt-4o")
	client.client = server.Client()

	summary, err := client.Summarize("test log content", "English")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no response choices returned")
	assert.Empty(s.T(), summary)
}

func (s *SummarizerTestSuite) TestSummarize_PromptGeneration() {
	logContent := "test log line 1\ntest log line 2"

	// Create server to capture the request
	var capturedRequest ChatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		response := ChatCompletionResponse{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Test summary",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-api-key", server.URL, "gpt-4")
	client.client = server.Client()

	_, err := client.Summarize(logContent, "English")

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "gpt-4", capturedRequest.Model)
	assert.Len(s.T(), capturedRequest.Messages, 1)
	assert.Equal(s.T(), "user", capturedRequest.Messages[0].Role)
	assert.Contains(s.T(), capturedRequest.Messages[0].Content, "analyze the following log content")
	assert.Contains(s.T(), capturedRequest.Messages[0].Content, logContent)
}
