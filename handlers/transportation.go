package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TransportationFormID is the fixed form template ID for Student School Transportation.
const TransportationFormID = "transportation_registration"

// TransportFormTemplate returns the fixed form template for Opp City Schools Transportation (student bus registration).
func TransportFormTemplate() *models.FormTemplate {
	now := time.Now().Format(time.RFC3339)
	return &models.FormTemplate{
		ID:          TransportationFormID,
		Name:        "Student School Transportation Form",
		Description: "Parents/guardians request bus transportation for their child, including pick-up/drop-off details and contact information.",
		UserType:    "student",
		Fields:      TransportFormFields(),
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   "system",
	}
}

// TransportFormFields returns the 14 fields for the transportation form (Opp City Schools Transportation Form).
func TransportFormFields() []models.FormField {
	return []models.FormField{
		{Name: "school_year", Label: "School Year", Type: "text", Required: true, Placeholder: "e.g. 2024-2025"},
		{Name: "school", Label: "School", Type: "select", Required: true, Placeholder: "Select school", Options: []string{}},
		{Name: "student_name", Label: "Student's Name", Type: "text", Required: true, Placeholder: ""},
		{Name: "grade", Label: "Grade", Type: "text", Required: true, Placeholder: "e.g. 3"},
		{Name: "address", Label: "Address", Type: "text", Required: true, Placeholder: "Street address"},
		{Name: "city", Label: "City", Type: "text", Required: true, Placeholder: ""},
		{Name: "state", Label: "State", Type: "text", Required: true, Placeholder: "e.g. AL"},
		{Name: "zip", Label: "Zip", Type: "text", Required: true, Placeholder: "ZIP code"},
		{Name: "transportation_needed", Label: "Transportation Needed", Type: "select", Required: true, Placeholder: "Select", Options: []string{"Pick-up only", "Drop-off only", "Both pick-up and drop-off"}},
		{Name: "parent_guardian_name", Label: "Parent/Guardian Name", Type: "text", Required: true, Placeholder: ""},
		{Name: "telephone", Label: "Telephone Number (with area code, must be an active number)", Type: "tel", Required: true, Placeholder: "e.g. 555-123-4567"},
		{Name: "bus_color", Label: "Bus Color", Type: "text", Required: false, Placeholder: "Optional"},
		{Name: "online_form_completed", Label: "Online Form Completed", Type: "text", Required: false, Placeholder: "Optional"},
		{Name: "bus_start_date", Label: "Date the student will start riding the school bus", Type: "date", Required: true, Placeholder: "YYYY-MM-DD"},
	}
}

// EnsureTransportationFormTemplate returns the transportation form template, creating and storing it if it does not exist.
func (h *Handlers) EnsureTransportationFormTemplate() (*models.FormTemplate, error) {
	t, err := h.db.GetFormTemplate(TransportationFormID)
	if err == nil && t != nil {
		return t, nil
	}
	t = TransportFormTemplate()
	if err := h.db.StoreFormTemplate(t); err != nil {
		return nil, fmt.Errorf("failed to store transportation form template: %w", err)
	}
	return t, nil
}

// Trigger phrases for transportation registration (first ~80 chars). "for a student" and "i want to" are ignorable.
var transportRegistrationPhrases = []string{
	"register a transportation for a student",
	"register transportation for a student",
	"register transportation for student",
	"register transportation",
	"request a transportation for a student",
	"request transportation for a student",
	"request transportation for student",
	"request transportation",
	"order a transportation for a student",
	"order transportation for a student",
	"order transportation for student",
	"order transportation",
	"i want to register a transportation",
	"i want to request transportation",
	"i want to order transportation",
	"i wanna register transportation",
	"i wanna request transportation",
	"transportation registration",
	"bus registration",
	"school bus registration",
	"sign up for bus",
	"sign up for school bus",
}

func isTransportationRegistrationRequest(message string) bool {
	s := strings.TrimSpace(message)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	start := lower
	if len(start) > 80 {
		start = start[:80]
	}
	for _, phrase := range transportRegistrationPhrases {
		if strings.HasPrefix(lower, phrase) || strings.Contains(start, phrase) {
			return true
		}
	}
	return false
}

func (h *Handlers) handleTransportationFlow(c *gin.Context, userID, userMessage string) (*models.ChatResponse, error) {
	ctx := context.Background()
	state, _ := h.db.GetTransportationStateByUserID(userID)

	// Ensure the transportation form template exists (for confirmation card and save).
	form, err := h.EnsureTransportationFormTemplate()
	if err != nil {
		log.Printf("[TRANSPORT] Ensure form error: %v", err)
		return nil, fmt.Errorf("failed to load transportation form: %w", err)
	}

	// If we are pending confirmation: user must confirm or request changes
	if state != nil && state.Step == "pending_confirmation" && state.FormID != "" {
		if isConfirmationMessage(userMessage) {
			submitterID := c.GetHeader("X-User-ID")
			if submitterID == "" {
				submitterID = "admin"
			}
			userIDForAnswer := ""
			for _, k := range []string{"student_name", "user_id", "name"} {
				if v, ok := state.GatheredAnswers[k]; ok {
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
				Answers:     state.GatheredAnswers,
				SubmittedAt: time.Now().Format(time.RFC3339),
				SubmittedBy: submitterID,
			}
			if err := h.db.StoreFormAnswer(fa); err != nil {
				log.Printf("[TRANSPORT] Store form answer error: %v", err)
				return nil, fmt.Errorf("failed to save transportation registration: %w", err)
			}
			h.db.DeleteTransportationState(userID)
			return &models.ChatResponse{
				Response: fmt.Sprintf("Transportation registration complete. Your **%s** has been submitted. You can view it under Form Answers.", state.FormName),
			}, nil
		}
		// User wants to change something: re-run gathering with current answers
		reply, err := h.aiService.TransportationFieldGatheringWithCurrent(ctx, form.Fields, state.GatheredAnswers, userMessage)
		if err != nil {
			log.Printf("[TRANSPORT] AI field update error: %v", err)
			return nil, fmt.Errorf("transportation AI error: %w", err)
		}
		complete, answers, ask := parseGatheringResponse(reply)
		if complete && len(answers) > 0 {
			state.Step = "pending_confirmation"
			state.GatheredAnswers = answers
			_ = h.db.StoreTransportationState(userID, state)
			return &models.ChatResponse{
				Response:         "I've updated the details. Please review the card below and reply **Confirm** to submit, or tell me what you'd like to change.",
				ConfirmationCard: h.buildConfirmationCard(state.FormName, state.UserType, answers, form.Fields),
			}, nil
		}
		if ask != "" {
			return &models.ChatResponse{Response: ask}, nil
		}
		return &models.ChatResponse{Response: "What would you like to change? Tell me the field and the new value."}, nil
	}

	// If we have an active session (gathering_fields), continue it
	if state != nil && state.Step == "gathering_fields" && state.FormID != "" {
		reply, err := h.aiService.TransportationFieldGathering(ctx, state.ConversationHistory, form.Fields, userMessage)
		if err != nil {
			log.Printf("[TRANSPORT] AI field gathering error: %v", err)
			return nil, fmt.Errorf("transportation AI error: %w", err)
		}

		complete, answers, ask := parseGatheringResponse(reply)
		if complete && len(answers) > 0 {
			state.Step = "pending_confirmation"
			state.GatheredAnswers = answers
			_ = h.db.StoreTransportationState(userID, state)
			return &models.ChatResponse{
				Response:         "Please review the details below. Reply **Confirm** to submit, or tell me what you'd like to change.",
				ConfirmationCard: h.buildConfirmationCard(state.FormName, state.UserType, answers, form.Fields),
			}, nil
		}

		if ask != "" {
			state.ConversationHistory = append(state.ConversationHistory, models.RegConvTurn{Role: "user", Content: userMessage}, models.RegConvTurn{Role: "assistant", Content: ask})
			state.LastAIResponse = ask
			state.ExchangeCount++
			if state.ExchangeCount >= 20 {
				h.db.DeleteTransportationState(userID)
				return &models.ChatResponse{Response: "We've hit the limit for this session. Please start again by saying you want to register or request transportation for a student."}, nil
			}
			_ = h.db.StoreTransportationState(userID, state)
			return &models.ChatResponse{Response: ask}, nil
		}

		fallback := "Please provide the missing required information so we can complete the transportation form."
		state.ConversationHistory = append(state.ConversationHistory, models.RegConvTurn{Role: "user", Content: userMessage}, models.RegConvTurn{Role: "assistant", Content: fallback})
		state.LastAIResponse = fallback
		state.ExchangeCount++
		_ = h.db.StoreTransportationState(userID, state)
		return &models.ChatResponse{Response: fallback}, nil
	}

	// New transportation registration intent
	if !isTransportationRegistrationRequest(userMessage) {
		if state != nil {
			h.db.DeleteTransportationState(userID)
		}
		return nil, nil
	}

	sid := uuid.New().String()
	state = &models.TransportationRegistrationState{
		ConversationID:      sid,
		Step:                "gathering_fields",
		FormID:              form.ID,
		FormName:            form.Name,
		UserType:            form.UserType,
		GatheredAnswers:     make(map[string]interface{}),
		ConversationHistory: nil,
		ExchangeCount:       0,
		CreatedAt:           time.Now().Format(time.RFC3339),
	}
	if err := h.db.StoreTransportationState(userID, state); err != nil {
		return nil, fmt.Errorf("failed to store transportation state: %w", err)
	}

	reply, err := h.aiService.TransportationFieldGathering(ctx, nil, form.Fields, userMessage)
	if err != nil {
		log.Printf("[TRANSPORT] First gathering AI error: %v", err)
		return nil, fmt.Errorf("transportation AI error: %w", err)
	}

	complete, answers, ask := parseGatheringResponse(reply)
	if complete && len(answers) > 0 {
		state.Step = "pending_confirmation"
		state.GatheredAnswers = answers
		_ = h.db.StoreTransportationState(userID, state)
		return &models.ChatResponse{
			Response:         "Please review the details below. Reply **Confirm** to submit, or tell me what you'd like to change.",
			ConfirmationCard: h.buildConfirmationCard(form.Name, form.UserType, answers, form.Fields),
		}, nil
	}

	if ask != "" {
		state.ConversationHistory = []models.RegConvTurn{
			{Role: "user", Content: userMessage},
			{Role: "assistant", Content: ask},
		}
		state.LastAIResponse = ask
		state.ExchangeCount = 1
		_ = h.db.StoreTransportationState(userID, state)
		return &models.ChatResponse{Response: ask}, nil
	}

	fallback := "I'll help you register for school bus transportation. Please tell me the school year, school name, student's name, grade, and address to get started."
	state.ConversationHistory = []models.RegConvTurn{
		{Role: "user", Content: userMessage},
		{Role: "assistant", Content: fallback},
	}
	state.LastAIResponse = fallback
	state.ExchangeCount = 1
	_ = h.db.StoreTransportationState(userID, state)
	return &models.ChatResponse{Response: fallback}, nil
}
