package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Trigger phrases must appear at or near the beginning of the message (first ~80 chars).
var registerStudentPhrases = []string{
	"i want to register a student",
	"i wanna register a student",
	"i want to register student",
	"i wanna register student",
	"register a student",
	"register student",
	"student register",
	"student registration",
	"i want to register",
	"i wanna register",
}

func isRegisterStudentRequest(message string) bool {
	s := strings.TrimSpace(message)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	start := lower
	if len(start) > 80 {
		start = start[:80]
	}
	for _, phrase := range registerStudentPhrases {
		if strings.HasPrefix(lower, phrase) || strings.Contains(start, phrase) {
			return true
		}
	}
	return false
}

// parseGatheringResponse tries to extract {"complete":true,"answers":{...}} or {"complete":false,"ask":"..."} from model output.
func parseGatheringResponse(raw string) (complete bool, answers map[string]interface{}, ask string) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	// Find first { ... } block
	start := strings.Index(s, "{")
	if start < 0 {
		return false, nil, ""
	}
	end := strings.LastIndex(s, "}")
	if end < start {
		return false, nil, ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s[start:end+1]), &m); err != nil {
		return false, nil, ""
	}
	if v, ok := m["complete"].(bool); ok && v {
		if a, ok := m["answers"].(map[string]interface{}); ok {
			return true, a, ""
		}
	}
	if v, ok := m["ask"].(string); ok {
		return false, nil, v
	}
	return false, nil, ""
}

// extractFormName returns the form name the model chose, or "" if NONE / unclear.
func extractFormName(modelReply string, formNames []string) string {
	s := strings.TrimSpace(strings.ToLower(modelReply))
	if strings.HasPrefix(s, "none") || s == "none" {
		return ""
	}
	for _, n := range formNames {
		if strings.Contains(s, strings.ToLower(n)) {
			return n
		}
	}
	// Try first token that matches a form name
	for _, n := range formNames {
		if strings.Contains(modelReply, n) {
			return n
		}
	}
	return ""
}

func (h *Handlers) handleRegistrationFlow(c *gin.Context, userID, userMessage string) (*models.ChatResponse, error) {
	ctx := context.Background()
	state, _ := h.db.GetRegistrationStateByUserID(userID)

	// If we have an active session (gathering_fields), continue it
	if state != nil && state.Step == "gathering_fields" && state.FormID != "" {
		form, err := h.db.GetFormTemplate(state.FormID)
		if err != nil || form == nil {
			log.Printf("[REG] Form %s not found, clearing state", state.FormID)
			h.db.DeleteRegistrationState(userID)
			return &models.ChatResponse{Response: "That form is no longer available. You can start again by saying you want to register a student."}, nil
		}

		// Pass existing history + current user message; we'll append both user and assistant after we get the reply
		reply, err := h.aiService.RegistrationFieldGathering(ctx, state.ConversationHistory, form.Fields, userMessage)
		if err != nil {
			log.Printf("[REG] AI field gathering error: %v", err)
			return nil, fmt.Errorf("registration AI error: %w", err)
		}

		complete, answers, ask := parseGatheringResponse(reply)
		if complete && len(answers) > 0 {
			// Resolve user_id for the form answer
			submitterID := c.GetHeader("X-User-ID")
			if submitterID == "" {
				submitterID = "admin"
			}
			userIDForAnswer := ""
			for _, k := range []string{"user_id", "student_id", "staff_number", "id", "name"} {
				if v, ok := answers[k]; ok {
					if s, ok := v.(string); ok && s != "" {
						userIDForAnswer = s
						break
					}
				}
			}
			if userIDForAnswer == "" {
				userIDForAnswer = submitterID
			}

			fa := &models.FormAnswer{
				ID:          uuid.New().String(),
				FormID:      state.FormID,
				FormName:    state.FormName,
				UserID:      userIDForAnswer,
				UserType:    state.UserType,
				Answers:     answers,
				SubmittedAt: time.Now().Format(time.RFC3339),
				SubmittedBy: submitterID,
			}
			if err := h.db.StoreFormAnswer(fa); err != nil {
				log.Printf("[REG] Store form answer error: %v", err)
				return nil, fmt.Errorf("failed to save registration: %w", err)
			}
			h.db.DeleteRegistrationState(userID)
			return &models.ChatResponse{
				Response: fmt.Sprintf("Registration complete. Your **%s** response has been saved. You can view it under Form Answers.", state.FormName),
			}, nil
		}

		if ask != "" {
			state.ConversationHistory = append(state.ConversationHistory, models.RegConvTurn{Role: "user", Content: userMessage}, models.RegConvTurn{Role: "assistant", Content: ask})
			state.LastAIResponse = ask
			state.ExchangeCount++
			if state.ExchangeCount >= 15 {
				h.db.DeleteRegistrationState(userID)
				return &models.ChatResponse{Response: "We've hit the limit for this session. Please start again by saying you want to register a student."}, nil
			}
			_ = h.db.StoreRegistrationState(userID, state)
			return &models.ChatResponse{Response: ask}, nil
		}

		// Unparseable: treat as "ask" and prompt again
		fallback := "Please provide the missing required fields so we can complete the form."
		state.ConversationHistory = append(state.ConversationHistory, models.RegConvTurn{Role: "user", Content: userMessage}, models.RegConvTurn{Role: "assistant", Content: fallback})
		state.LastAIResponse = fallback
		state.ExchangeCount++
		_ = h.db.StoreRegistrationState(userID, state)
		return &models.ChatResponse{Response: fallback}, nil
	}

	// New registration intent
	if !isRegisterStudentRequest(userMessage) {
		if state != nil {
			h.db.DeleteRegistrationState(userID)
		}
		return nil, nil // caller will continue with normal chat
	}

	templates, err := h.db.GetAllFormTemplates()
	if err != nil {
		log.Printf("[REG] Get templates error: %v", err)
		return nil, fmt.Errorf("failed to load forms: %w", err)
	}
	if len(templates) == 0 {
		return &models.ChatResponse{
			Response: "There are no registration forms set up yet. Use the **Forms** menu to create a Student Registration (or similar) form, then try again.",
		}, nil
	}

	var namesDesc []string
	var formNames []string
	for _, t := range templates {
		formNames = append(formNames, t.Name)
		desc := t.Description
		if desc == "" {
			desc = "no description"
		}
		namesDesc = append(namesDesc, fmt.Sprintf("%s (%s)", t.Name, desc))
	}
	formListForAI := strings.Join(namesDesc, "\n")

	chosen, err := h.aiService.RegistrationFormSelect(ctx, userMessage, formListForAI)
	if err != nil {
		log.Printf("[REG] Form select AI error: %v", err)
		return nil, fmt.Errorf("registration form selection error: %w", err)
	}
	if chosen == "" {
		chosen = " "
	}
	chosenName := extractFormName(chosen, formNames)
	if chosenName == "" {
		return &models.ChatResponse{
			Response: "I couldn't match that to a specific form. Try something like: \"I want to register a student\" and we'll use the Student Registration form if you have one, or add one under **Forms**.",
		}, nil
	}

	var selected *models.FormTemplate
	for i := range templates {
		if templates[i].Name == chosenName {
			selected = &templates[i]
			break
		}
	}
	if selected == nil {
		return &models.ChatResponse{
			Response: "I couldn't find that form. Please use **Forms** to create or check form names, then try again.",
		}, nil
	}

	sid := uuid.New().String()
	state = &models.RegistrationState{
		ConversationID:      sid,
		Step:                "gathering_fields",
		FormID:              selected.ID,
		FormName:            selected.Name,
		UserType:            selected.UserType,
		GatheredAnswers:     make(map[string]interface{}),
		ConversationHistory: nil,
		ExchangeCount:       0,
		CreatedAt:           time.Now().Format(time.RFC3339),
	}
	if err := h.db.StoreRegistrationState(userID, state); err != nil {
		return nil, fmt.Errorf("failed to store registration state: %w", err)
	}

	// First gathering turn: do we already have all required fields from this first message? Pass empty history.
	reply, err := h.aiService.RegistrationFieldGathering(ctx, nil, selected.Fields, userMessage)
	if err != nil {
		log.Printf("[REG] First gathering AI error: %v", err)
		return nil, fmt.Errorf("registration AI error: %w", err)
	}

	complete, answers, ask := parseGatheringResponse(reply)
	if complete && len(answers) > 0 {
		submitterID := c.GetHeader("X-User-ID")
		if submitterID == "" {
			submitterID = "admin"
		}
		userIDForAnswer := ""
		for _, k := range []string{"user_id", "student_id", "staff_number", "id", "name"} {
			if v, ok := answers[k]; ok {
				if s, ok := v.(string); ok && s != "" {
					userIDForAnswer = s
					break
				}
			}
		}
		if userIDForAnswer == "" {
			userIDForAnswer = submitterID
		}
		fa := &models.FormAnswer{
			ID:          uuid.New().String(),
			FormID:      selected.ID,
			FormName:    selected.Name,
			UserID:      userIDForAnswer,
			UserType:    selected.UserType,
			Answers:     answers,
			SubmittedAt: time.Now().Format(time.RFC3339),
			SubmittedBy: submitterID,
		}
		if err := h.db.StoreFormAnswer(fa); err != nil {
			return nil, fmt.Errorf("failed to save registration: %w", err)
		}
		h.db.DeleteRegistrationState(userID)
		return &models.ChatResponse{
			Response: fmt.Sprintf("Registration complete. Your **%s** response has been saved from what you provided. You can view it under Form Answers.", selected.Name),
		}, nil
	}

	if ask != "" {
		state.ConversationHistory = []models.RegConvTurn{
			{Role: "user", Content: userMessage},
			{Role: "assistant", Content: ask},
		}
		state.LastAIResponse = ask
		state.ExchangeCount = 1
		_ = h.db.StoreRegistrationState(userID, state)
		return &models.ChatResponse{Response: ask}, nil
	}

	fallback := "Please tell me the details needed for " + selected.Name + " (e.g. name, age, contact)."
	state.ConversationHistory = []models.RegConvTurn{
		{Role: "user", Content: userMessage},
		{Role: "assistant", Content: fallback},
	}
	state.LastAIResponse = fallback
	state.ExchangeCount = 1
	_ = h.db.StoreRegistrationState(userID, state)
	return &models.ChatResponse{Response: fallback}, nil
}
