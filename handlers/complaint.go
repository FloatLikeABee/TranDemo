package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// isComplaintRequest checks if the user message is about filing a complaint
// It detects both explicit complaint requests and messages containing complaint details
func isComplaintRequest(message string) bool {
	lowerMsg := strings.ToLower(message)
	
	// Explicit complaint phrases
	complaintPhrases := []string{
		"file a complaint",
		"file complaint",
		"filing a complaint",
		"filing complaint",
		"want to file a complaint",
		"want to file complaint",
		"i want to file a complaint",
		"i want to file complaint",
		"i wanna file a complaint",
		"i wanna file complaint",
		"complaint against",
		"file complaint against",
		"file a complaint on",
		"file a complaint against",
		"report a user",
		"report user",
		"complain about",
		"complain against",
		"report on",
		"i need to report a student's behavior",
		"need to report a student's behavior",
		"report a student's behavior",
		"report student's behavior",
		"report student behavior",
		"report behavior",
		"behavior report",
		"misconduct form",
		"fill out a misconduct form",
		"fill out misconduct form",
		"misconduct",
		"please help me fill out a misconduct form",
		"help me fill out a misconduct form",
		"help fill out misconduct form",
	}

	for _, phrase := range complaintPhrases {
		if strings.Contains(lowerMsg, phrase) {
			return true
		}
	}

	// Also detect messages that contain complaint details even without explicit "file complaint"
	// These patterns suggest the user wants to report something
	complaintIndicators := []string{
		"threat",
		"threatening",
		"kill",
		"gun",
		"weapon",
		"harass",
		"harassment",
		"abuse",
		"abusive",
		"inappropriate",
		"violence",
		"violent",
		"attack",
		"assault",
		"bully",
		"bullying",
		"misconduct",
		"behavior",
		"student's behavior",
		"student behavior",
		"on the bus",
		"bus behavior",
		"bus incident",
	}
	
	// Check if message contains complaint indicators AND mentions a person/name
	// This helps distinguish between actual complaints and general conversation
	hasComplaintIndicator := false
	for _, indicator := range complaintIndicators {
		if strings.Contains(lowerMsg, indicator) {
			hasComplaintIndicator = true
			break
		}
	}
	
	// If it has complaint indicators and mentions "on", "against", "report", "form", or "complaint" (suggesting reporting something), treat as complaint
	if hasComplaintIndicator && (strings.Contains(lowerMsg, " on ") || 
		strings.Contains(lowerMsg, " against ") || 
		strings.Contains(lowerMsg, "complaint") ||
		strings.Contains(lowerMsg, "report") ||
		strings.Contains(lowerMsg, "form") ||
		strings.Contains(lowerMsg, "misconduct") ||
		strings.Contains(lowerMsg, "behavior")) {
		return true
	}

	return false
}

// handleComplaintFlow handles the multi-step complaint filing process
func (h *Handlers) handleComplaintFlow(c *gin.Context, userID, userMessage string) (*models.ChatResponse, error) {
	// Correct spelling errors in user message before processing
	correctedMessage, err := h.aiService.CorrectSpelling(userMessage)
	if err != nil {
		log.Printf("[COMPLAINT FLOW] Error correcting spelling: %v, using original message", err)
		correctedMessage = userMessage
	} else if correctedMessage != userMessage {
		log.Printf("[COMPLAINT FLOW] Spelling corrected: '%s' -> '%s'", userMessage, correctedMessage)
		userMessage = correctedMessage
	}

	// If user message is a complaint initiation phrase, ALWAYS start a NEW session
	isNewComplaintRequest := isComplaintRequest(userMessage)

	// Get existing complaint state (if any)
	complaintState, err := h.db.GetComplaintStateByUserID(userID)

	// If user is initiating a new complaint, clear old state and start fresh
	if isNewComplaintRequest {
		if complaintState != nil && complaintState.ConversationID != "" {
			log.Printf("[COMPLAINT FLOW] User initiated new complaint, clearing old state (conversationID: %s)", complaintState.ConversationID)
			// Mark old state as complete to clear it
			complaintState.Step = "complete"
			h.db.StoreComplaintState(userID, complaintState)
		}
		complaintState = nil // Force new session
	}

	// If no state exists or state is complete, start a NEW complaint session
	if err != nil || complaintState == nil || complaintState.Step == "complete" || complaintState.ConversationID == "" {
		log.Printf("[COMPLAINT FLOW] Starting NEW complaint session for user: %s", userID)

		// Step 1: Initialize and capture initial_data
		initResp, err := h.complaintService.InitializeProcess()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
		}

		// Step 2: Start dialogue with the full user message (including complaint details)
		log.Printf("[COMPLAINT FLOW] Starting dialogue with full message: %s", userMessage)
		dialogueResp, err := h.complaintService.StartDialogue(userMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to start dialogue: %w", err)
		}

		log.Printf("[COMPLAINT FLOW] Started dialogue, conversationID: '%s'", dialogueResp.ConversationID)

		// Validate conversationID before storing
		if dialogueResp.ConversationID == "" {
			log.Printf("[COMPLAINT FLOW] ERROR: conversationID is empty! Cannot store state.")
			return nil, fmt.Errorf("conversation_id is empty from dialogue start response")
		}

		// Create new state with conversation_id and initial_data
		complaintState = &models.ComplaintState{
			ConversationID: dialogueResp.ConversationID,
			Step:           "dialogue",
			ExchangeCount:  1, // First exchange (user message + AI response)
			LastResponse:   dialogueResp.Response,
			InitialData:   initResp.InitialData, // Store initial_data from first execute step
		}
		
		log.Printf("[COMPLAINT FLOW] Stored initial_data with %d keys", len(initResp.InitialData))

		log.Printf("[COMPLAINT FLOW] About to store state - userID: %s, conversationID: %s", userID, complaintState.ConversationID)

		// Store state immediately - CRITICAL: must store before returning
		if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
			log.Printf("[COMPLAINT FLOW] ERROR storing complaint state: %v", err)
			return nil, fmt.Errorf("failed to store complaint state: %w", err)
		}

		log.Printf("[COMPLAINT FLOW] Successfully stored complaint state for user %s with conversationID %s, step: %s",
			userID, dialogueResp.ConversationID, complaintState.Step)

		// Store chat history
		h.db.StoreChatHistory(userID, userMessage, dialogueResp.Response)

		return &models.ChatResponse{
			Response: dialogueResp.Response,
		}, nil
	}

	// We have an active session - continue it
	log.Printf("[COMPLAINT FLOW] Continuing existing session for user %s, conversationID: %s, step: %s, exchanges: %d",
		userID, complaintState.ConversationID, complaintState.Step, complaintState.ExchangeCount)

	// Continue existing session - check if we've exceeded max exchanges
	if complaintState.ExchangeCount >= 12 {
		log.Printf("[COMPLAINT FLOW] Session exceeded 12 exchanges, clearing old state and starting new session for user %s", userID)
		// Mark old state as complete and start fresh
		complaintState.Step = "complete"
		h.db.StoreComplaintState(userID, complaintState)
		// Fall through to start a new session (will be handled by the check at the top)
		complaintState = nil
		err = fmt.Errorf("session exceeded max exchanges")
	}

	// If we cleared the state above, start a new session
	if complaintState == nil {
		log.Printf("[COMPLAINT FLOW] Starting NEW complaint session (old session cleared) for user: %s", userID)

		// Step 1: Initialize and capture initial_data
		initResp, err := h.complaintService.InitializeProcess()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
		}

		// Step 2: Start dialogue with the full user message (including complaint details)
		log.Printf("[COMPLAINT FLOW] Starting dialogue with full message: %s", userMessage)
		dialogueResp, err := h.complaintService.StartDialogue(userMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to start dialogue: %w", err)
		}

		log.Printf("[COMPLAINT FLOW] Started dialogue, conversationID: '%s'", dialogueResp.ConversationID)

		// Validate conversationID before storing
		if dialogueResp.ConversationID == "" {
			log.Printf("[COMPLAINT FLOW] ERROR: conversationID is empty! Cannot store state.")
			return nil, fmt.Errorf("conversation_id is empty from dialogue start response")
		}

		// Create new state with conversation_id and initial_data
		complaintState = &models.ComplaintState{
			ConversationID: dialogueResp.ConversationID,
			Step:           "dialogue",
			ExchangeCount:  1, // First exchange (user message + AI response)
			LastResponse:   dialogueResp.Response,
			InitialData:   initResp.InitialData, // Store initial_data from first execute step
		}
		
		log.Printf("[COMPLAINT FLOW] Stored initial_data with %d keys", len(initResp.InitialData))

		// Store state immediately
		if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
			log.Printf("[COMPLAINT FLOW] ERROR storing complaint state: %v", err)
			return nil, fmt.Errorf("failed to store complaint state: %w", err)
		}

		// Store chat history
		h.db.StoreChatHistory(userID, userMessage, dialogueResp.Response)

		return &models.ChatResponse{
			Response: dialogueResp.Response,
		}, nil
	}

	log.Printf("[COMPLAINT FLOW] Continuing session - Exchange %d/12, ConversationID: %s",
		complaintState.ExchangeCount+1, complaintState.ConversationID)

	// Continue dialogue in the session
	continueResp, err := h.complaintService.ContinueDialogue(complaintState.ConversationID, userMessage)
	if err != nil {
		// Check if error is "Maximum number of turns reached"
		errStr := err.Error()
		if strings.Contains(errStr, "Maximum number of turns reached") || strings.Contains(errStr, "maximum number of turns") {
			log.Printf("[COMPLAINT FLOW] Old conversation hit max turns, starting new session for user %s", userID)
			// Clear old state and start fresh
			complaintState.Step = "complete"
			h.db.StoreComplaintState(userID, complaintState)

			// Start new session
			initResp, err := h.complaintService.InitializeProcess()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
			}

			// Start dialogue with the full user message (including complaint details)
			log.Printf("[COMPLAINT FLOW] Starting dialogue with full message: %s", userMessage)
			dialogueResp, err := h.complaintService.StartDialogue(userMessage)
			if err != nil {
				return nil, fmt.Errorf("failed to start dialogue: %w", err)
			}

			// Create new state with initial_data
			complaintState = &models.ComplaintState{
				ConversationID: dialogueResp.ConversationID,
				Step:           "dialogue",
				ExchangeCount:  1,
				LastResponse:   dialogueResp.Response,
				InitialData:   initResp.InitialData, // Store initial_data from first execute step
			}

			if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
				return nil, fmt.Errorf("failed to store complaint state: %w", err)
			}

			h.db.StoreChatHistory(userID, userMessage, dialogueResp.Response)

			return &models.ChatResponse{
				Response: dialogueResp.Response,
			}, nil
		}

		log.Printf("[COMPLAINT FLOW] Error continuing dialogue: %v", err)
		return nil, fmt.Errorf("failed to continue dialogue: %w", err)
	}

	// Increment exchange count
	complaintState.ExchangeCount++
	complaintState.ConversationID = continueResp.ConversationID
	complaintState.LastResponse = continueResp.Response

	log.Printf("[COMPLAINT FLOW] Dialogue continued - Exchange count: %d, is_complete: %v, Response: %s",
		complaintState.ExchangeCount, continueResp.IsComplete, continueResp.Response)

	// NEW FLOW: Check if is_complete is true
	if continueResp.IsComplete {
		log.Printf("[COMPLAINT FLOW] Dialogue is complete, executing with response body")
		
		// Use the entire response body as the request body for execute
		// The RawResponse contains the full response from continue dialogue
		var dialogueResult map[string]interface{}
		if continueResp.RawResponse != nil {
			// Use the raw response directly, but make a copy to avoid modifying the original
			dialogueResult = make(map[string]interface{})
			for k, v := range continueResp.RawResponse {
				dialogueResult[k] = v
			}
		} else {
			// If RawResponse is not available, construct it from the fields
			dialogueResult = make(map[string]interface{})
			dialogueResult["conversation_id"] = continueResp.ConversationID
			dialogueResult["dialogue_id"] = continueResp.DialogueID
			dialogueResult["response"] = continueResp.Response
			dialogueResult["turn_number"] = continueResp.TurnNumber
			dialogueResult["max_turns"] = continueResp.MaxTurns
			dialogueResult["needs_more_info"] = continueResp.NeedsMoreInfo
			dialogueResult["is_complete"] = continueResp.IsComplete
			dialogueResult["needs_user_input"] = continueResp.NeedsUserInput
			if len(continueResp.ConversationHistory) > 0 {
				dialogueResult["conversation_history"] = continueResp.ConversationHistory
			}
			if continueResp.LLMProvider != "" {
				dialogueResult["llm_provider"] = continueResp.LLMProvider
			}
			if continueResp.ModelName != "" {
				dialogueResult["model_name"] = continueResp.ModelName
			}
		}

		// Build the execute request body with the structure expected by the API
		// Structure: resume_from_phase, dialogue_phase1_result (from continue), and initial_data (from first execute)
		executeRequestBody := map[string]interface{}{
			"resume_from_phase":     "dialogue",
			"dialogue_phase1_result": dialogueResult,
		}

		// Add initial_data from the first execute step (required)
		if complaintState.InitialData != nil {
			executeRequestBody["initial_data"] = complaintState.InitialData
			log.Printf("[COMPLAINT FLOW] Added initial_data to execute request with %d keys", len(complaintState.InitialData))
		} else {
			log.Printf("[COMPLAINT FLOW] WARNING: initial_data is nil, execute request may fail")
		}

		// Execute using the request body
		executeResp, err := h.complaintService.ExecuteWithResponseBody(executeRequestBody)
		if err != nil {
			log.Printf("[COMPLAINT FLOW] Error executing with response body: %v", err)
			return nil, fmt.Errorf("failed to execute complaint: %w", err)
		}

		// Check if we have final_outcome
		if executeResp.FinalOutcome != nil {
			log.Printf("[COMPLAINT FLOW] Final outcome received: %v", executeResp.FinalOutcome)
			// Print to console
			finalOutcomeJSON, _ := json.MarshalIndent(executeResp.FinalOutcome, "", "  ")
			log.Printf("[COMPLAINT FLOW] Final outcome (console only):\n%s", string(finalOutcomeJSON))

			// Mark as complete
			complaintState.Step = "complete"
			if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
				log.Printf("Error storing final complaint state: %v", err)
			}

			// Store success message in chat history
			successMsg := "We have successfully filed the complaint for you. Your complaint has been received and will be reviewed by our team. Thank you for bringing this to our attention."
			h.db.StoreChatHistory(userID, userMessage, successMsg)

			return &models.ChatResponse{
				Response: successMsg,
			}, nil
		} else {
			log.Printf("[COMPLAINT FLOW] No final outcome received, but dialogue is complete")
			// Even if no final_outcome, mark as complete since dialogue is done
			complaintState.Step = "complete"
			if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
				log.Printf("Error storing complaint state: %v", err)
			}

			// Return the response message
			return &models.ChatResponse{
				Response: continueResp.Response,
			}, nil
		}
	}

	// Dialogue is not complete yet - just display the response message
	// Update state
	if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
		log.Printf("Error storing complaint state: %v", err)
	}

	// Store chat history
	h.db.StoreChatHistory(userID, userMessage, continueResp.Response)

	// Return the response message
	return &models.ChatResponse{
		Response: continueResp.Response,
	}, nil
}
