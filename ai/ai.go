package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"idongivaflyinfa/cache"
	"idongivaflyinfa/config"
	"idongivaflyinfa/models"
)

type AIService struct {
	apiKey    string
	modelName string
	cache     *cache.Cache
	httpClient *http.Client
	apiURL    string
}

type DashScopeRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []DashScopeMessage `json:"messages"`
	} `json:"input"`
}

type DashScopeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DashScopeResponse struct {
	Output struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	RequestID string `json:"request_id,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

func New(apiKey string, modelName string, cache *cache.Cache) (*AIService, error) {
	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	return &AIService{
		apiKey:     apiKey,
		modelName:  modelName,
		cache:      cache,
		httpClient: httpClient,
		apiURL:     "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
	}, nil
}

func (a *AIService) Close() error {
	// HTTP client doesn't require explicit closing
	return nil
}

func (a *AIService) callDashScopeAPI(ctx context.Context, messages []DashScopeMessage) (string, error) {
	reqBody := DashScopeRequest{
		Model: a.modelName,
	}
	reqBody.Input.Messages = messages

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Set("Content-Type", "application/json")
	
	// Debug: Print request details (remove in production)
	fmt.Printf("Request URL: %s\n", a.apiURL)
	fmt.Printf("Request Model: %s\n", a.modelName)
	fmt.Printf("Request Body: %s\n", string(jsonData))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Debug: Print response details
	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResp struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", fmt.Errorf("API error (status %d): %s - %s (request_id: %s)", 
				resp.StatusCode, errorResp.Code, errorResp.Message, errorResp.RequestID)
		}
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var dashScopeResp DashScopeResponse
	if err := json.Unmarshal(body, &dashScopeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if dashScopeResp.Code != "" && dashScopeResp.Code != "Success" {
		return "", fmt.Errorf("API error: %s - %s", dashScopeResp.Code, dashScopeResp.Message)
	}

	if len(dashScopeResp.Output.Choices) == 0 {
		return "", fmt.Errorf("no response from AI model")
	}

	return dashScopeResp.Output.Choices[0].Message.Content, nil
}

func (a *AIService) GenerateSQL(userPrompt string, sqlFiles []models.SQLFile) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("prompt:%s", userPrompt)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Build prompt using helper
	prompt := BuildSQLPrompt(userPrompt, sqlFiles)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	fmt.Println("prompt:", prompt)

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		fmt.Println("error:", err)
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	sql := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```SQL")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	// Cache the result
	a.cache.SetDefault(cacheKey, sql)

	return sql, nil
}

func (a *AIService) GenerateForm(userPrompt string) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("form_prompt:%s", userPrompt)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Sample JSON form structure - loaded from config
	sampleJSON := config.FormSampleJSON

	// Build prompt using helper
	prompt := BuildFormPrompt(userPrompt, sampleJSON)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate form: %w", err)
	}

	// Clean up the response - remove markdown code blocks if present
	formJSON := strings.TrimSpace(response)
	formJSON = strings.TrimPrefix(formJSON, "```json")
	formJSON = strings.TrimPrefix(formJSON, "```JSON")
	formJSON = strings.TrimPrefix(formJSON, "```")
	formJSON = strings.TrimSuffix(formJSON, "```")
	formJSON = strings.TrimSpace(formJSON)

	// Validate JSON
	var testJSON interface{}
	if err := json.Unmarshal([]byte(formJSON), &testJSON); err != nil {
		return "", fmt.Errorf("generated JSON is invalid: %w", err)
	}

	// Cache the result
	a.cache.SetDefault(cacheKey, formJSON)

	return formJSON, nil
}

func (a *AIService) GenerateHTMLPage(resultFile *models.ResultFile, title string) (string, error) {
	ctx := context.Background()

	// Build prompt using helper
	prompt := BuildHTMLPagePrompt(resultFile, title)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}

	html := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	html = strings.TrimPrefix(html, "```html")
	html = strings.TrimPrefix(html, "```HTML")
	html = strings.TrimPrefix(html, "```")
	html = strings.TrimSuffix(html, "```")
	html = strings.TrimSpace(html)

	return html, nil
}

func (a *AIService) GenerateFormHTMLPage(formJSON string) (string, error) {
	ctx := context.Background()

	// Parse form JSON to extract form name and description
	var formData map[string]interface{}
	if err := json.Unmarshal([]byte(formJSON), &formData); err != nil {
		return "", fmt.Errorf("failed to parse form JSON: %w", err)
	}

	formName := ""
	formDescription := ""
	if name, ok := formData["Name"].(string); ok {
		formName = name
	}
	if desc, ok := formData["Description"].(string); ok {
		formDescription = desc
	}

	// Build prompt using helper
	prompt := BuildFormHTMLPrompt(formJSON, formName, formDescription)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate form HTML: %w", err)
	}

	html := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	html = strings.TrimPrefix(html, "```html")
	html = strings.TrimPrefix(html, "```HTML")
	html = strings.TrimPrefix(html, "```")
	html = strings.TrimSuffix(html, "```")
	html = strings.TrimSpace(html)

	return html, nil
}

// GenerateChatResponse generates a plain chat response for general prompts
func (a *AIService) GenerateChatResponse(userPrompt string) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("chat_prompt:%s", userPrompt)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Build a simple chat prompt
	prompt := fmt.Sprintf("You are a helpful assistant. Please respond to the following user message in a helpful and informative way:\n\n%s", userPrompt)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate chat response: %w", err)
	}

	// Clean up the response - remove markdown code blocks if present
	chatResponse := strings.TrimSpace(response)
	chatResponse = strings.TrimPrefix(chatResponse, "```")
	chatResponse = strings.TrimSuffix(chatResponse, "```")
	chatResponse = strings.TrimSpace(chatResponse)

	// Cache the result
	a.cache.SetDefault(cacheKey, chatResponse)

	return chatResponse, nil
}
