package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	ClaudeAPIURL = "https://api.anthropic.com/v1/messages"
	ModelName    = "claude-3-5-sonnet-20241022"
	MaxTokens    = 4000
)

type Client struct {
	apiKey string
	client *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type Response struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	ID           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Type         string `json:"type"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type ReviewRequest struct {
	Title       string
	Description string
	DiffContent string
	Language    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *Client) ReviewCode(req ReviewRequest) (string, error) {
	prompt := c.buildReviewPrompt(req)
	
	claudeReq := Request{
		Model:     ModelName,
		MaxTokens: MaxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	
	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}
	
	httpReq, err := http.NewRequest("POST", ClaudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude API error %d: %s", resp.StatusCode, string(body))
	}
	
	var claudeResp Response
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	
	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}
	
	return claudeResp.Content[0].Text, nil
}

func (c *Client) buildReviewPrompt(req ReviewRequest) string {
	return fmt.Sprintf(`You are an expert code reviewer. Please review this merge request and provide constructive feedback.

**Merge Request Details:**
- Title: %s
- Description: %s

**Code Changes:**
%s

**Instructions:**
1. Focus on code quality, security vulnerabilities, performance issues, and maintainability
2. Provide specific suggestions for improvement with line references where possible
3. Highlight any potential bugs or edge cases
4. Check for proper error handling and input validation
5. Consider if the code follows best practices for the language/framework
6. Be constructive and helpful in your feedback
7. If the code looks good overall, mention the positive aspects

**Format your response as a GitLab comment using Markdown:**
- Use clear headings for different types of issues
- Use code blocks for suggested changes
- Be concise but thorough
- End with a summary recommendation

Please provide your code review now:`, req.Title, req.Description, req.DiffContent)
}