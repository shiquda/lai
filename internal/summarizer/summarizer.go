package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

// ErrorAnalysisResult represents the structured response for error detection
type ErrorAnalysisResult struct {
	HasError bool   `json:"has_error"`
	Severity string `json:"severity"` // "error", "warning", "info"
	Summary  string `json:"summary"`
}

func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o"
	}

	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (c *OpenAIClient) SetClient(client *http.Client) {
	c.client = client
}

func (c *OpenAIClient) Summarize(logContent string, language string) (string, error) {
	if language == "" {
		language = "English"
	}

	return c.SummarizeWithTemplate(logContent, language, "")
}

// SummarizeWithTemplate summarizes log content using a custom template
func (c *OpenAIClient) SummarizeWithTemplate(logContent, language, customTemplate string) (string, error) {
	if language == "" {
		language = "English"
	}

	// Determine which template to use
	var template string
	if customTemplate != "" {
		template = customTemplate
	} else {
		// Use built-in template
		template = `Please analyze the following log content and generate a summary in {{language}}:

1. Identify key events and errors
2. Count important metrics
3. Mark anomalies or issues that need attention
4. Provide a concise summary in {{language}}

Log content:
{{log_content}}`
	}

	// Create template engine and render template
	engine := NewTemplateEngine()
	engine.SetBuiltinVariable("language", language)

	variables := map[string]string{
		"log_content": logContent,
		"language":    language,
	}

	prompt, err := engine.RenderTemplate(template, variables)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt template: %w", err)
	}

	req := ChatCompletionRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// AnalyzeForErrors analyzes log content to determine if it contains errors or exceptions
func (c *OpenAIClient) AnalyzeForErrors(logContent string, language string) (*ErrorAnalysisResult, error) {
	if language == "" {
		language = "English"
	}

	return c.AnalyzeForErrorsWithTemplate(logContent, language, "")
}

// AnalyzeForErrorsWithTemplate analyzes log content using a custom template
func (c *OpenAIClient) AnalyzeForErrorsWithTemplate(logContent, language, customTemplate string) (*ErrorAnalysisResult, error) {
	if language == "" {
		language = "English"
	}

	// Determine which template to use
	var template string
	if customTemplate != "" {
		template = customTemplate
	} else {
		// Use built-in template
		template = `Please analyze the following log content and determine if it contains errors, exceptions, or warnings that require attention.

Respond with a valid JSON object in the following format:
{
  "has_error": true/false,
  "severity": "error"/"warning"/"info",
  "summary": "Brief description of the issue in {{language}} or 'No errors detected'"
}

Guidelines:
- Set "has_error" to true if the log contains:
  - Error messages or stack traces
  - Exception information
  - Critical warnings or failures
  - System crashes or timeouts
- Set "severity" to:
  - "error" for critical errors, exceptions, crashes
  - "warning" for warnings that need attention but don't break functionality
  - "info" for normal informational messages
- Provide a concise summary in {{language}}

Log content:
{{log_content}}`
	}

	// Create template engine and render template
	engine := NewTemplateEngine()
	engine.SetBuiltinVariable("language", language)

	variables := map[string]string{
		"log_content": logContent,
		"language":    language,
	}

	prompt, err := engine.RenderTemplate(template, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render error analysis template: %w", err)
	}

	req := ChatCompletionRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	// Parse the JSON response from LLM
	content := strings.TrimSpace(response.Choices[0].Message.Content)

	// Try to extract JSON from the response if it contains extra text
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonContent := content[startIdx : endIdx+1]

		var result ErrorAnalysisResult
		if err := json.Unmarshal([]byte(jsonContent), &result); err == nil {
			return &result, nil
		}
	}

	// If JSON extraction fails, try parsing the entire content
	var result ErrorAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fallback: if JSON parsing fails completely, treat as no error but log the issue
		return &ErrorAnalysisResult{
			HasError: false,
			Severity: "info",
			Summary:  fmt.Sprintf("Could not parse LLM response as JSON, treating as normal log. Response: %s", content),
		}, nil
	}

	return &result, nil
}
