package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const ComplaintAPIBaseURL = "http://localhost:8000"

type ComplaintService struct {
	httpClient *http.Client
}

func NewComplaintService() *ComplaintService {
	return &ComplaintService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Step 1: Initialize the process
func (s *ComplaintService) InitializeProcess() error {
	url := fmt.Sprintf("%s/special-flows-1/localrunsimulationstucomplaintform/execute", ComplaintAPIBaseURL)
	
	reqBody := map[string]interface{}{
		"initial_input": "",
		"context":       map[string]interface{}{},
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("[COMPLAINT STEP 1] Request URL: %s", url)
	log.Printf("[COMPLAINT STEP 1] Request Body: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[COMPLAINT STEP 1] Response Status: %d", resp.StatusCode)
	log.Printf("[COMPLAINT STEP 1] Response Body: %s", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Step 2: Start dialogue
type StartDialogueRequest struct {
	InitialMessage string `json:"initial_message"`
	NResults       int    `json:"n_results"`
}

type StartDialogueResponse struct {
	ConversationID string `json:"conversation_id"`
	Response       string `json:"response"`
	// Add other fields as needed
}

func (s *ComplaintService) StartDialogue(initialMessage string) (*StartDialogueResponse, error) {
	url := fmt.Sprintf("%s/dialogues/flow_localrunsimulationstucomplaintform_dialogue/start", ComplaintAPIBaseURL)
	
	reqBody := StartDialogueRequest{
		InitialMessage: initialMessage,
		NResults:       20, // Maximum allowed by API (was 50, but API limit is 20)
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("[COMPLAINT STEP 2] Request URL: %s", url)
	log.Printf("[COMPLAINT STEP 2] Request Body: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("[COMPLAINT STEP 2] Response Status: %d", resp.StatusCode)
	log.Printf("[COMPLAINT STEP 2] Response Body: %s", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	// Always parse as raw JSON first to extract all fields
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	result := StartDialogueResponse{}
	
	// Extract conversation_id - try multiple possible field names
	if convID, ok := rawResp["conversation_id"].(string); ok && convID != "" {
		result.ConversationID = convID
	} else if convID, ok := rawResp["conversationId"].(string); ok && convID != "" {
		result.ConversationID = convID
	} else if convID, ok := rawResp["conversationID"].(string); ok && convID != "" {
		result.ConversationID = convID
	}
	
	// Extract response - try multiple possible field names
	if respText, ok := rawResp["response"].(string); ok {
		result.Response = respText
	} else if message, ok := rawResp["message"].(string); ok {
		result.Response = message
	} else if content, ok := rawResp["content"].(string); ok {
		result.Response = content
	}
	
	if result.ConversationID == "" {
		log.Printf("[COMPLAINT STEP 2] WARNING: conversationID is empty! Raw response keys: %v", getKeys(rawResp))
		// Try to find it in nested structures
		if data, ok := rawResp["data"].(map[string]interface{}); ok {
			if convID, ok := data["conversation_id"].(string); ok {
				result.ConversationID = convID
			}
		}
	}
	
	log.Printf("[COMPLAINT STEP 2] Parsed - ConversationID: '%s', Response length: %d", result.ConversationID, len(result.Response))
	if result.ConversationID == "" {
		log.Printf("[COMPLAINT STEP 2] ERROR: conversationID is still empty after parsing!")
		return nil, fmt.Errorf("conversation_id not found in response")
	}
	
	return &result, nil
}

// Helper function to get keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Step 3-4: Continue dialogue
type ContinueDialogueRequest struct {
	ConversationID string `json:"conversation_id"`
	UserMessage    string `json:"user_message"`
}

type ContinueDialogueResponse struct {
	Response        string                 `json:"response"`
	ConversationID  string                 `json:"conversation_id"`
	DialogueID      string                 `json:"dialogue_id"`
	TurnNumber      int                    `json:"turn_number"`
	MaxTurns        int                    `json:"max_turns"`
	NeedsMoreInfo   bool                   `json:"needs_more_info"`
	IsComplete      bool                   `json:"is_complete"`
	NeedsUserInput  bool                   `json:"needs_user_input"`
	ConversationHistory []map[string]interface{} `json:"conversation_history"`
	LLMProvider     string                 `json:"llm_provider"`
	ModelName       string                 `json:"model_name"`
	RawResponse     map[string]interface{} `json:"-"` // Store full response
}

func (s *ComplaintService) ContinueDialogue(conversationID, userMessage string) (*ContinueDialogueResponse, error) {
	url := fmt.Sprintf("%s/dialogues/flow_localrunsimulationstucomplaintform_dialogue/continue", ComplaintAPIBaseURL)
	
	reqBody := ContinueDialogueRequest{
		ConversationID: conversationID,
		UserMessage:    userMessage,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("[COMPLAINT CONTINUE] Request URL: %s", url)
	log.Printf("[COMPLAINT CONTINUE] ConversationID: %s", conversationID)
	log.Printf("[COMPLAINT CONTINUE] UserMessage: %s", userMessage)
	log.Printf("[COMPLAINT CONTINUE] Request Body: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("[COMPLAINT CONTINUE] Response Status: %d", resp.StatusCode)
	log.Printf("[COMPLAINT CONTINUE] Response Body: %s", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	result := ContinueDialogueResponse{
		RawResponse: rawResp,
	}
	
	// Extract fields
	if convID, ok := rawResp["conversation_id"].(string); ok {
		result.ConversationID = convID
	} else {
		result.ConversationID = conversationID
	}
	if respText, ok := rawResp["response"].(string); ok {
		result.Response = respText
	} else if message, ok := rawResp["message"].(string); ok {
		result.Response = message
	}
	if dialogueID, ok := rawResp["dialogue_id"].(string); ok {
		result.DialogueID = dialogueID
	}
	if turnNum, ok := rawResp["turn_number"].(float64); ok {
		result.TurnNumber = int(turnNum)
	}
	if maxTurns, ok := rawResp["max_turns"].(float64); ok {
		result.MaxTurns = int(maxTurns)
	}
	if needsMore, ok := rawResp["needs_more_info"].(bool); ok {
		result.NeedsMoreInfo = needsMore
	}
	if isComplete, ok := rawResp["is_complete"].(bool); ok {
		result.IsComplete = isComplete
	}
	if needsInput, ok := rawResp["needs_user_input"].(bool); ok {
		result.NeedsUserInput = needsInput
	}
	if history, ok := rawResp["conversation_history"].([]interface{}); ok {
		result.ConversationHistory = make([]map[string]interface{}, len(history))
		for i, h := range history {
			if hMap, ok := h.(map[string]interface{}); ok {
				result.ConversationHistory[i] = hMap
			}
		}
	}
	if provider, ok := rawResp["llm_provider"].(string); ok {
		result.LLMProvider = provider
	}
	if model, ok := rawResp["model_name"].(string); ok {
		result.ModelName = model
	}
	
	log.Printf("[COMPLAINT CONTINUE] Parsed - ConversationID: %s, Response: %s", result.ConversationID, result.Response)
	
	return &result, nil
}

// Step 5: Get dialogue info
type DialogueInfo struct {
	ID          string `json:"id"`
	SystemPrompt string `json:"system_prompt"`
	// Add other fields as needed
}

func (s *ComplaintService) GetDialogueInfo() ([]DialogueInfo, error) {
	url := fmt.Sprintf("%s/dialogues", ComplaintAPIBaseURL)
	
	log.Printf("[COMPLAINT STEP 5] Request URL: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json, text/plain, */*")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("[COMPLAINT STEP 5] Response Status: %d", resp.StatusCode)
	log.Printf("[COMPLAINT STEP 5] Response Body: %s", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	var dialogues []DialogueInfo
	if err := json.Unmarshal(body, &dialogues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	log.Printf("[COMPLAINT STEP 5] Found %d dialogues", len(dialogues))
	
	return dialogues, nil
}

// Step 6 & 8: Execute with dialogue result
type ExecuteRequest struct {
	ResumeFromPhase    string                 `json:"resume_from_phase"`
	DialoguePhase1Result map[string]interface{} `json:"dialogue_phase1_result"`
	InitialData        map[string]interface{} `json:"initial_data"`
}

type ExecuteResponse struct {
	FinalOutcome interface{} `json:"final_outcome"`
	// Add other fields as needed
}

func (s *ComplaintService) ExecuteWithDialogueResult(dialogueResult map[string]interface{}, initialData map[string]interface{}) (*ExecuteResponse, error) {
	url := fmt.Sprintf("%s/special-flows-1/localrunsimulationstucomplaintform/execute", ComplaintAPIBaseURL)
	
	reqBody := ExecuteRequest{
		ResumeFromPhase:    "dialogue",
		DialoguePhase1Result: dialogueResult,
		InitialData:        initialData,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("[COMPLAINT EXECUTE] Request URL: %s", url)
	log.Printf("[COMPLAINT EXECUTE] Request Body: %s", string(jsonData))
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("[COMPLAINT EXECUTE] Response Status: %d", resp.StatusCode)
	log.Printf("[COMPLAINT EXECUTE] Response Body: %s", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	var result ExecuteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// Try to extract final_outcome from raw JSON
		var rawResp map[string]interface{}
		if err2 := json.Unmarshal(body, &rawResp); err2 != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		
		if outcome, ok := rawResp["final_outcome"]; ok {
			result.FinalOutcome = outcome
		}
	}
	
	if result.FinalOutcome != nil {
		log.Printf("[COMPLAINT EXECUTE] Final outcome received: %v", result.FinalOutcome)
	} else {
		log.Printf("[COMPLAINT EXECUTE] Final outcome is NULL")
	}
	
	return &result, nil
}

