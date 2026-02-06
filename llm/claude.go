package llm

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	claudeModel      = "claude-sonnet-4@20250514"
	anthropicVersion = "vertex-2023-10-16"
)

//go:embed system_prompt.md
var systemPrompt string

// ThreadAnalysis represents the LLM analysis result for a discussion thread
type ThreadAnalysis struct {
	DeveloperFeedbackScore int    `json:"developer_feedback_score"`
	Summary                string `json:"summary"`
}

// Client handles communication with the Claude API
type Client struct {
	modelAPI   string // Base API URL
	userKey    string // Bearer token
	httpClient *http.Client
}

// NewClient creates a new Claude API client with configurable endpoint
func NewClient(modelAPI, userKey string) *Client {
	return &Client{
		modelAPI: modelAPI,
		userKey:  userKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// claudeRequest represents the API request structure
type claudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	System           string          `json:"system"`
	MaxTokens        int             `json:"max_tokens"`
	Messages         []claudeMessage `json:"messages"`
	Temperature      float64         `json:"temperature"`
}

// claudeMessage represents a message in the conversation
type claudeMessage struct {
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
}

// claudeContent represents content within a message
type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// claudeResponse represents the API response structure
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AnalyzeThread analyzes a discussion thread and returns a score and summary
func (c *Client) AnalyzeThread(ctx context.Context, comments []string) (*ThreadAnalysis, error) {
	if len(comments) == 0 {
		return nil, fmt.Errorf("no comments to analyze")
	}

	prompt := buildPrompt(comments)

	req := claudeRequest{
		AnthropicVersion: anthropicVersion,
		System:           strings.TrimSpace(systemPrompt),
		MaxTokens:        256,
		Temperature:      0,
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []claudeContent{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/sonnet/models/%s:streamRawPredict", c.modelAPI, claudeModel)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.userKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 || claudeResp.Content[0].Text == "" {
		return nil, fmt.Errorf("empty response from API")
	}

	return parseAnalysis(claudeResp.Content[0].Text)
}

// buildPrompt constructs the user prompt for thread analysis
func buildPrompt(comments []string) string {
	var prompt bytes.Buffer

	for i, comment := range comments {
		if i == 0 {
			prompt.WriteString("AI SUGGESTION:\n")
		} else {
			prompt.WriteString(fmt.Sprintf("\nRESPONSE %d:\n", i))
		}
		prompt.WriteString(comment)
		prompt.WriteString("\n")
	}

	return prompt.String()
}

// parseAnalysis parses the LLM response into a ThreadAnalysis struct
func parseAnalysis(text string) (*ThreadAnalysis, error) {
	var analysis ThreadAnalysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w (response: %s)", err, text)
	}

	// Clamp score to valid range
	if analysis.DeveloperFeedbackScore < 0 {
		analysis.DeveloperFeedbackScore = 0
	}
	if analysis.DeveloperFeedbackScore > 10 {
		analysis.DeveloperFeedbackScore = 10
	}

	return &analysis, nil
}
