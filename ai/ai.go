package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"idongivaflyinfa/cache"
	"idongivaflyinfa/config"
	"idongivaflyinfa/models"
)

type AIService struct {
	apiKey               string
	modelName            string
	cache                *cache.Cache
	httpClient           *http.Client
	httpClientLongTimeout *http.Client // For operations that may take longer (HTML generation)
	apiURL               string
	lastRequestTime      time.Time    // Track last request time for rate limiting
	requestMutex         sync.Mutex   // Mutex to protect lastRequestTime
	minRequestInterval   time.Duration // Minimum time between requests
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
	
	// HTTP client with longer timeout for HTML generation (5 minutes)
	httpClientLongTimeout := &http.Client{
		Timeout: 300 * time.Second,
	}

	return &AIService{
		apiKey:               apiKey,
		modelName:            modelName,
		cache:                cache,
		httpClient:           httpClient,
		httpClientLongTimeout: httpClientLongTimeout,
		apiURL:               "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
		lastRequestTime:      time.Time{},
		minRequestInterval:   500 * time.Millisecond, // Minimum 500ms between requests
	}, nil
}

func (a *AIService) Close() error {
	// HTTP client doesn't require explicit closing
	return nil
}

func (a *AIService) callDashScopeAPI(ctx context.Context, messages []DashScopeMessage) (string, error) {
	return a.callDashScopeAPIWithClient(ctx, messages, a.httpClient)
}

// rateLimit ensures minimum time between requests to prevent burst rate errors
func (a *AIService) rateLimit() {
	a.requestMutex.Lock()
	defer a.requestMutex.Unlock()

	now := time.Now()
	timeSinceLastRequest := now.Sub(a.lastRequestTime)

	if timeSinceLastRequest < a.minRequestInterval {
		// Need to wait to maintain minimum interval
		waitTime := a.minRequestInterval - timeSinceLastRequest
		time.Sleep(waitTime)
	}

	a.lastRequestTime = time.Now()
}

func (a *AIService) callDashScopeAPIWithClient(ctx context.Context, messages []DashScopeMessage, client *http.Client) (string, error) {
	// Apply rate limiting before making request
	a.rateLimit()

	reqBody := DashScopeRequest{
		Model: a.modelName,
	}
	reqBody.Input.Messages = messages

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic with exponential backoff for rate limit errors
	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2s, 4s, 8s
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			fmt.Printf("Rate limit hit, retrying after %v (attempt %d/%d)\n", delay, attempt, maxRetries)
			time.Sleep(delay)
			// Re-apply rate limiting after backoff
			a.rateLimit()
		}

		req, err := http.NewRequestWithContext(ctx, "POST", a.apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
		req.Header.Set("Content-Type", "application/json")
		
		// Debug: Print request details (remove in production)
		if attempt == 0 {
			fmt.Printf("Request URL: %s\n", a.apiURL)
			fmt.Printf("Request Model: %s\n", a.modelName)
			fmt.Printf("Request Body: %s\n", string(jsonData))
		}

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue // Retry on network errors
			}
			return "", fmt.Errorf("failed to send request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries {
				continue // Retry on read errors
			}
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		// Debug: Print response details
		fmt.Printf("Response Status: %d\n", resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Response Body: %s\n", string(body))
		}

		// Handle rate limiting (429) with retry
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				// Try to parse error response for more details
				var errorResp struct {
					Code      string `json:"code"`
					Message   string `json:"message"`
					RequestID string `json:"request_id"`
				}
				if err := json.Unmarshal(body, &errorResp); err == nil {
					fmt.Printf("Rate limit error: %s - %s (request_id: %s)\n", 
						errorResp.Code, errorResp.Message, errorResp.RequestID)
				}
				continue // Retry with backoff
			}
			// Max retries reached, return error
			var errorResp struct {
				Code      string `json:"code"`
				Message   string `json:"message"`
				RequestID string `json:"request_id"`
			}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				return "", fmt.Errorf("API error (status %d): %s - %s (request_id: %s). Max retries exceeded.", 
					resp.StatusCode, errorResp.Code, errorResp.Message, errorResp.RequestID)
			}
			return "", fmt.Errorf("API returned status %d: %s. Max retries exceeded.", resp.StatusCode, string(body))
		}

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

	return "", fmt.Errorf("max retries exceeded")
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

// ClassifyDocumentIntent returns "FORM", "RESEARCH", or "SUMMARY" based on user message and document content.
func (a *AIService) ClassifyDocumentIntent(userMessage, extractedText, aiResult string) (string, error) {
	ctx := context.Background()
	prompt := BuildDocumentIntentPrompt(userMessage, extractedText, aiResult)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	reply, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "SUMMARY", err
	}
	s := strings.TrimSpace(strings.ToUpper(reply))
	if strings.Contains(s, "FORM") {
		return "FORM", nil
	}
	if strings.Contains(s, "RESEARCH") {
		return "RESEARCH", nil
	}
	return "SUMMARY", nil
}

// GenerateFormTemplateFromContent generates a FormTemplate (name, description, user_type, fields) from document content.
func (a *AIService) GenerateFormTemplateFromContent(content string, userContext string) (*models.FormTemplate, error) {
	ctx := context.Background()
	prompt := BuildFormTemplateFromContentPrompt(content, userContext)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	reply, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(reply)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	var t models.FormTemplate
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, fmt.Errorf("invalid form template JSON: %w", err)
	}
	if t.UserType == "" {
		t.UserType = "general"
	}
	if t.UserType != "student" && t.UserType != "staff" {
		t.UserType = "general"
	}
	return &t, nil
}

func (a *AIService) GenerateHTMLPage(resultFile *models.ResultFile, title string) (string, error) {
	// Use context with longer timeout for HTML generation (5 minutes)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Build prompt using helper
	prompt := BuildHTMLPagePrompt(resultFile, title)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Use the long timeout client for HTML generation
	response, err := a.callDashScopeAPIWithClient(ctx, messages, a.httpClientLongTimeout)
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
	// Use context with longer timeout for HTML generation (5 minutes)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

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

	// Use the long timeout client for HTML generation
	response, err := a.callDashScopeAPIWithClient(ctx, messages, a.httpClientLongTimeout)
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

// CorrectSpelling corrects spelling errors in user input using AI
// It preserves the user's intent while fixing typos and misspellings
func (a *AIService) CorrectSpelling(userInput string) (string, error) {
	// Skip correction for very short inputs or if input seems fine
	if len(userInput) < 3 {
		return userInput, nil
	}

	// Check cache first
	cacheKey := fmt.Sprintf("spell_correct:%s", userInput)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Build prompt for spelling correction
	prompt := fmt.Sprintf(`You are a spelling and grammar correction assistant. Your task is to correct spelling errors and typos in the user's message while preserving their exact meaning and intent. 

IMPORTANT RULES:
1. Only correct actual spelling mistakes and typos
2. Preserve the user's original meaning and intent completely
3. Keep the same tone and style
4. Do NOT change words that are intentionally informal (like "wanna", "gonna", "yeah")
5. Do NOT add or remove words unless they are clearly typos
6. Fix spacing issues (e.g., "iwanna" -> "i wanna")
7. Return ONLY the corrected text, nothing else - no explanations, no markdown, just the corrected message

User's message to correct:
"%s"

Corrected message:`, userInput)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		// If AI correction fails, return original input
		return userInput, nil
	}

	// Clean up the response
	corrected := strings.TrimSpace(response)
	// Remove any markdown code blocks if present
	corrected = strings.TrimPrefix(corrected, "```")
	corrected = strings.TrimSuffix(corrected, "```")
	corrected = strings.TrimSpace(corrected)

	// If correction is empty or same as input, return original
	if corrected == "" || corrected == userInput {
		return userInput, nil
	}

	// Cache the result
	a.cache.SetDefault(cacheKey, corrected)

	return corrected, nil
}

// GenerateFromMessages calls the model with the given message list (e.g. system + user + assistant + user).
// Used by registration flow and other custom prompts.
func (a *AIService) GenerateFromMessages(ctx context.Context, messages []DashScopeMessage) (string, error) {
	return a.callDashScopeAPI(ctx, messages)
}

// RegistrationFormSelect asks the model to pick one form name from a list (no IDs). Returns the model reply (form name or NONE).
func (a *AIService) RegistrationFormSelect(ctx context.Context, userMessage, formNamesDescriptions string) (string, error) {
	sys, user := BuildFormSelectionPrompt(userMessage, formNamesDescriptions)
	messages := []DashScopeMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: user},
	}
	return a.callDashScopeAPI(ctx, messages)
}

// RegistrationFieldGathering asks the model whether we have all required fields or what to ask next. Returns raw reply (JSON string).
func (a *AIService) RegistrationFieldGathering(ctx context.Context, conversationHistory []models.RegConvTurn, formFields []models.FormField, latestUserMessage string) (string, error) {
	sys, conv := BuildFieldGatheringPrompt(conversationHistory, formFields, latestUserMessage)
	messages := []DashScopeMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: conv},
	}
	return a.callDashScopeAPI(ctx, messages)
}

// RegistrationFieldGatheringWithCurrent merges the user's change request into current answers (confirmation-edit flow).
func (a *AIService) RegistrationFieldGatheringWithCurrent(ctx context.Context, formFields []models.FormField, currentAnswers map[string]interface{}, userMessage string) (string, error) {
	sys, user := BuildFieldGatheringPromptWithCurrent(formFields, currentAnswers, userMessage)
	messages := []DashScopeMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: user},
	}
	return a.callDashScopeAPI(ctx, messages)
}