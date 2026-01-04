package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"idongivaflyinfa/models"
	"idongivaflyinfa/service"

	"github.com/gin-gonic/gin"
)

// isComplaintRequest checks if the user message is about filing a complaint
func isComplaintRequest(message string) bool {
	lowerMsg := strings.ToLower(message)
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
		"report a user",
		"report user",
		"complain about",
		"complain against",
	}

	for _, phrase := range complaintPhrases {
		if strings.Contains(lowerMsg, phrase) {
			return true
		}
	}

	return false
}

// handleComplaintFlow handles the multi-step complaint filing process as a continuous session
func (h *Handlers) handleComplaintFlow(c *gin.Context, userID, userMessage string) (*models.ChatResponse, error) {
	// If user message is a complaint initiation phrase, ALWAYS start a NEW session
	// This handles the case where an old conversation exists but hit max turns
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

		// Step 1: Initialize
		if err := h.complaintService.InitializeProcess(); err != nil {
			return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
		}

		// Step 2: Start dialogue
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

		// Create new state with conversation_id
		complaintState = &models.ComplaintState{
			ConversationID: dialogueResp.ConversationID,
			Step:           "dialogue",
			ExchangeCount:  1, // First exchange (user message + AI response)
			LastResponse:   dialogueResp.Response,
		}

		log.Printf("[COMPLAINT FLOW] About to store state - userID: %s, conversationID: %s", userID, complaintState.ConversationID)

		// Store state immediately - CRITICAL: must store before returning
		if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
			log.Printf("[COMPLAINT FLOW] ERROR storing complaint state: %v", err)
			return nil, fmt.Errorf("failed to store complaint state: %w", err)
		}

		log.Printf("[COMPLAINT FLOW] Successfully stored complaint state for user %s with conversationID %s, step: %s",
			userID, dialogueResp.ConversationID, complaintState.Step)

		// Verify it was stored by trying to retrieve it immediately
		verifyState, verifyErr := h.db.GetComplaintStateByUserID(userID)
		if verifyErr != nil {
			log.Printf("[COMPLAINT FLOW] WARNING: Could not verify stored state: %v", verifyErr)
		} else if verifyState != nil {
			log.Printf("[COMPLAINT FLOW] Verified state stored - conversationID: %s", verifyState.ConversationID)
		} else {
			log.Printf("[COMPLAINT FLOW] WARNING: State was stored but could not be retrieved immediately")
		}

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

		// Step 1: Initialize
		if err := h.complaintService.InitializeProcess(); err != nil {
			return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
		}

		// Step 2: Start dialogue
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

		// Create new state with conversation_id
		complaintState = &models.ComplaintState{
			ConversationID: dialogueResp.ConversationID,
			Step:           "dialogue",
			ExchangeCount:  1, // First exchange (user message + AI response)
			LastResponse:   dialogueResp.Response,
		}

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
			if err := h.complaintService.InitializeProcess(); err != nil {
				return nil, fmt.Errorf("failed to initialize complaint process: %w", err)
			}

			dialogueResp, err := h.complaintService.StartDialogue(userMessage)
			if err != nil {
				return nil, fmt.Errorf("failed to start dialogue: %w", err)
			}

			// Create new state
			complaintState = &models.ComplaintState{
				ConversationID: dialogueResp.ConversationID,
				Step:           "dialogue",
				ExchangeCount:  1,
				LastResponse:   dialogueResp.Response,
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

	// Check if the previous AI response was asking for complaint details
	// This helps us identify when the user is providing complaint text
	prevResponse := complaintState.LastResponse
	isAskingForComplaint := strings.Contains(strings.ToLower(prevResponse), "what's your complaint") ||
		strings.Contains(strings.ToLower(prevResponse), "what is your complaint") ||
		strings.Contains(strings.ToLower(prevResponse), "complaint against the user")

	// Check if this looks like a complaint text (contains details about the complaint)
	lowerMsg := strings.ToLower(userMessage)
	isComplaintText := strings.Contains(lowerMsg, "threat") ||
		strings.Contains(lowerMsg, "harass") ||
		strings.Contains(lowerMsg, "abuse") ||
		strings.Contains(lowerMsg, "inappropriate") ||
		strings.Contains(lowerMsg, "complaint") ||
		len(userMessage) > 20 || // Likely a detailed complaint
		isAskingForComplaint // If AI asked for complaint, this is likely it

	// Save complaint text if detected and not already saved
	if (isComplaintText || isAskingForComplaint) && complaintState.ComplaintText == "" {
		// Save complaint text
		complaintState.ComplaintText = userMessage
		log.Printf("[COMPLAINT FLOW] Saved complaint text: %s", userMessage)
		// Save state immediately to ensure complaint text is persisted
		if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
			log.Printf("Error storing complaint state with complaint text: %v", err)
		}
	}

	// Update LastResponse after checking for complaint text
	complaintState.LastResponse = continueResp.Response

	log.Printf("[COMPLAINT FLOW] Dialogue continued - Exchange count: %d, Response: %s",
		complaintState.ExchangeCount, continueResp.Response)

	// Check if the response contains JSON (indicating we got the form result)
	// This means the dialogue might be complete and we should try to execute
	if strings.Contains(continueResp.Response, "{") && strings.Contains(continueResp.Response, "formId") {
		// Dialogue seems complete, try to execute
		// But we'll continue the session if final_outcome is null
		return h.executeComplaintFlow(userID, complaintState, continueResp)
	}

	// Update state (if not already saved above)
	if !isComplaintText || complaintState.ComplaintText != userMessage {
		if err := h.db.StoreComplaintState(userID, complaintState); err != nil {
			log.Printf("Error storing complaint state: %v", err)
		}
	}

	// Store chat history
	h.db.StoreChatHistory(userID, userMessage, continueResp.Response)

	// Continue the dialogue session - keep going until we get final_outcome
	return &models.ChatResponse{
		Response: continueResp.Response,
	}, nil
}

// executeComplaintFlow executes steps 5-8 of the complaint flow
// This is called when we detect the dialogue might be complete (contains formId JSON)
// It will continue the session if final_outcome is null
func (h *Handlers) executeComplaintFlow(userID string, state *models.ComplaintState, dialogueResp *service.ContinueDialogueResponse) (*models.ChatResponse, error) {
	log.Printf("Executing complaint flow for conversation %s (exchange count: %d)", state.ConversationID, state.ExchangeCount)
	// Step 5: Get dialogue info
	dialogues, err := h.complaintService.GetDialogueInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogue info: %w", err)
	}

	// Find the complaint dialogue
	var complaintDialogue *service.DialogueInfo
	for i := range dialogues {
		if dialogues[i].ID == "flow_localrunsimulationstucomplaintform_dialogue" {
			complaintDialogue = &dialogues[i]
			break
		}
	}

	if complaintDialogue == nil {
		return nil, fmt.Errorf("complaint dialogue not found")
	}

	// Build dialogue_phase1_result from the continue response
	// Use the raw response if available, otherwise construct from fields
	// NOTE: system_prompt should NOT be included in dialogue_phase1_result
	var dialogueResult map[string]interface{}
	if dialogueResp.RawResponse != nil {
		// Copy the raw response but exclude system_prompt
		dialogueResult = make(map[string]interface{})
		for k, v := range dialogueResp.RawResponse {
			if k != "system_prompt" {
				dialogueResult[k] = v
			}
		}
		// Ensure conversation_id and dialogue_id are set
		if dialogueResult["conversation_id"] == nil {
			dialogueResult["conversation_id"] = state.ConversationID
		}
		if dialogueResult["dialogue_id"] == nil {
			dialogueResult["dialogue_id"] = "flow_localrunsimulationstucomplaintform_dialogue"
		}
	} else {
		// Fallback: construct from fields (without system_prompt)
		dialogueResult = map[string]interface{}{
			"conversation_id":      state.ConversationID,
			"dialogue_id":          "flow_localrunsimulationstucomplaintform_dialogue",
			"response":             dialogueResp.Response,
			"turn_number":          dialogueResp.TurnNumber,
			"max_turns":            dialogueResp.MaxTurns,
			"needs_more_info":      dialogueResp.NeedsMoreInfo,
			"is_complete":          dialogueResp.IsComplete,
			"needs_user_input":     dialogueResp.NeedsUserInput,
			"conversation_history": dialogueResp.ConversationHistory,
			"llm_provider":         dialogueResp.LLMProvider,
			"model_name":           dialogueResp.ModelName,
		}
	}

	// Build initial_data (this should come from step 1, but we'll use a default structure)
	initialData := map[string]interface{}{
		"tool_id":   "locallite",
		"tool_name": "LocalLite",
		"columns":   []string{"id", "name"},
		"rows": [][]interface{}{
			{"d0b70978-f2d0-472a-9a3b-97a3f1e91d0c", "User Behavior Complaint form "},
			{"4c4d5a3d-e3d4-405d-8a81-a13f8de02dbf", "User attendance report form"},
		},
		"total_rows":       2,
		"cached":           false,
		"cache_expires_at": nil,
		"metadata":         map[string]interface{}{},
	}

	// Step 6: Execute with dialogue result
	executeResp, err := h.complaintService.ExecuteWithDialogueResult(dialogueResult, initialData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute complaint: %w", err)
	}

	// Check if we already have final_outcome from step 6
	if executeResp.FinalOutcome != nil {
		log.Printf("[COMPLAINT FLOW] Final outcome received at step 6: %v", executeResp.FinalOutcome)
		// Print to console
		finalOutcomeJSON, _ := json.MarshalIndent(executeResp.FinalOutcome, "", "  ")
		log.Printf("[COMPLAINT FLOW] Final outcome (console only):\n%s", string(finalOutcomeJSON))

		// Mark as complete
		state.Step = "complete"
		state.DialogueResult = dialogueResult
		state.InitialData = initialData
		if err := h.db.StoreComplaintState(userID, state); err != nil {
			log.Printf("Error storing final complaint state: %v", err)
		}

		// Store success message in chat history
		successMsg := "We have successfully filed the complaint for you. Your complaint has been received and will be reviewed by our team. Thank you for bringing this to our attention."
		h.db.StoreChatHistory(userID, state.ComplaintText, successMsg)

		return &models.ChatResponse{
			Response: successMsg,
		}, nil
	}

	// Check if final_outcome is null - if so, continue to step 7
	if executeResp.FinalOutcome == nil {
		log.Printf("[COMPLAINT FLOW] Final outcome is null after step 6, proceeding to step 7 (exchange count: %d)", state.ExchangeCount)

		// If we haven't exceeded max exchanges, continue to step 7
		if state.ExchangeCount < 12 {
			// Check if we have complaint text saved
			if state.ComplaintText == "" {
				log.Printf("[COMPLAINT FLOW] WARNING: No complaint text saved, cannot proceed to step 7")
				// Update state and continue session normally
				state.DialogueResult = dialogueResult
				state.InitialData = initialData
				state.Step = "dialogue"
				if err := h.db.StoreComplaintState(userID, state); err != nil {
					log.Printf("Error storing complaint state: %v", err)
				}
				return &models.ChatResponse{
					Response: "Please provide more details about your complaint.",
				}, nil
			}

			// Step 7: Continue dialogue with saved complaint text (don't display response to UI)
			log.Printf("[COMPLAINT FLOW] Step 7: Continuing dialogue with complaint text: %s", state.ComplaintText)
			continueResp, err := h.complaintService.ContinueDialogue(state.ConversationID, state.ComplaintText)
			if err != nil {
				return nil, fmt.Errorf("failed to continue dialogue with complaint: %w", err)
			}

			log.Printf("[COMPLAINT FLOW] Step 7 response received (not shown to UI): %s", continueResp.Response)

			// Update exchange count
			state.ExchangeCount++
			state.LastResponse = continueResp.Response

			// Update dialogue result with new response from step 7
			// Build dialogue_phase1_result for step 8 (without system_prompt)
			dialogueResultForStep8 := make(map[string]interface{})

			// Start with step 7 response data
			if continueResp.RawResponse != nil {
				// Copy from raw response but exclude system_prompt
				for k, v := range continueResp.RawResponse {
					if k != "system_prompt" {
						dialogueResultForStep8[k] = v
					}
				}
			} else {
				// Build from individual fields
				dialogueResultForStep8["conversation_id"] = state.ConversationID
				dialogueResultForStep8["dialogue_id"] = "flow_localrunsimulationstucomplaintform_dialogue"
				dialogueResultForStep8["response"] = continueResp.Response
				dialogueResultForStep8["turn_number"] = continueResp.TurnNumber
				dialogueResultForStep8["max_turns"] = continueResp.MaxTurns
				dialogueResultForStep8["needs_more_info"] = continueResp.NeedsMoreInfo
				dialogueResultForStep8["is_complete"] = continueResp.IsComplete
				dialogueResultForStep8["needs_user_input"] = continueResp.NeedsUserInput
				if len(continueResp.ConversationHistory) > 0 {
					dialogueResultForStep8["conversation_history"] = continueResp.ConversationHistory
				}
				if continueResp.LLMProvider != "" {
					dialogueResultForStep8["llm_provider"] = continueResp.LLMProvider
				}
				if continueResp.ModelName != "" {
					dialogueResultForStep8["model_name"] = continueResp.ModelName
				}
			}

			// Ensure required fields are set and override with explicit values for step 8
			if dialogueResultForStep8["conversation_id"] == nil {
				dialogueResultForStep8["conversation_id"] = state.ConversationID
			}
			if dialogueResultForStep8["dialogue_id"] == nil {
				dialogueResultForStep8["dialogue_id"] = "flow_localrunsimulationstucomplaintform_dialogue"
			}

			// Explicitly set required fields for step 8
			dialogueResultForStep8["turn_number"] = 5
			dialogueResultForStep8["needs_more_info"] = true
			dialogueResultForStep8["is_complete"] = true
			dialogueResultForStep8["needs_user_input"] = false
			dialogueResultForStep8["llm_provider"] = "qwen"
			dialogueResultForStep8["model_name"] = "qwen3-max"

			// Ensure response is set (should be the JSON string from step 7)
			if dialogueResultForStep8["response"] == nil {
				dialogueResultForStep8["response"] = continueResp.Response
			}

			// Ensure initial_data has required fields (metadata is NOT needed)
			if initialData["cached"] == nil {
				initialData["cached"] = false
			}
			if initialData["cache_expires_at"] == nil {
				initialData["cache_expires_at"] = nil
			}
			// Remove metadata if it exists (not needed for step 8)
			delete(initialData, "metadata")

			// Step 8: Execute again with updated dialogue result (without system_prompt)
			log.Printf("[COMPLAINT FLOW] Step 8: Executing with updated dialogue result (turn_number: %v)", dialogueResultForStep8["turn_number"])
			executeResp, err = h.complaintService.ExecuteWithDialogueResult(dialogueResultForStep8, initialData)
			if err != nil {
				return nil, fmt.Errorf("failed to execute complaint (step 8): %w", err)
			}

			// Check if we have final_outcome from step 8
			if executeResp.FinalOutcome != nil {
				log.Printf("[COMPLAINT FLOW] Final outcome received at step 8: %v", executeResp.FinalOutcome)
				// Print to console
				finalOutcomeJSON, _ := json.MarshalIndent(executeResp.FinalOutcome, "", "  ")
				log.Printf("[COMPLAINT FLOW] Final outcome (console only):\n%s", string(finalOutcomeJSON))

				// Mark as complete
				state.Step = "complete"
				state.DialogueResult = dialogueResult
				state.InitialData = initialData
				if err := h.db.StoreComplaintState(userID, state); err != nil {
					log.Printf("Error storing final complaint state: %v", err)
				}

				// Store success message in chat history
				successMsg := "We have successfully filed the complaint for you. Your complaint has been received and will be reviewed by our team. Thank you for bringing this to our attention."
				h.db.StoreChatHistory(userID, state.ComplaintText, successMsg)

				return &models.ChatResponse{
					Response: successMsg,
				}, nil
			}

			// If still no final_outcome and we haven't exceeded exchanges, update state and continue session
			if executeResp.FinalOutcome == nil && state.ExchangeCount < 12 {
				log.Printf("[COMPLAINT FLOW] No final_outcome yet after step 8, continuing session (exchange: %d/12)", state.ExchangeCount)
				// Update state and return a message to continue the session
				state.DialogueResult = dialogueResult
				state.InitialData = initialData
				state.Step = "dialogue" // Keep in dialogue mode to continue
				if err := h.db.StoreComplaintState(userID, state); err != nil {
					log.Printf("Error storing complaint state: %v", err)
				}

				// Return a message to continue the dialogue - user will send another message
				return &models.ChatResponse{
					Response: "Please provide any additional details about your complaint.",
				}, nil
			}
		} else {
			// Max exchanges reached, end session
			state.Step = "complete"
			h.db.StoreComplaintState(userID, state)
			return &models.ChatResponse{
				Response: "I apologize, but we've reached the maximum number of exchanges. Please try filing your complaint again with more specific details.",
			}, nil
		}
	}

	// This should only be reached if final_outcome was found (safety fallback)
	// Print final_outcome to console if it exists
	if executeResp.FinalOutcome != nil {
		log.Printf("[COMPLAINT FLOW] Final outcome received (fallback): %v", executeResp.FinalOutcome)
		finalOutcomeJSON, _ := json.MarshalIndent(executeResp.FinalOutcome, "", "  ")
		log.Printf("[COMPLAINT FLOW] Final outcome (console only):\n%s", string(finalOutcomeJSON))
	}

	// Mark as complete
	state.Step = "complete"
	state.DialogueResult = dialogueResult
	state.InitialData = initialData
	if err := h.db.StoreComplaintState(userID, state); err != nil {
		log.Printf("Error storing final complaint state: %v", err)
	}

	// Store success message in chat history
	successMsg := "We have successfully filed the complaint for you. Your complaint has been received and will be reviewed by our team. Thank you for bringing this to our attention."
	h.db.StoreChatHistory(userID, state.ComplaintText, successMsg)

	return &models.ChatResponse{
		Response: successMsg,
	}, nil
}
