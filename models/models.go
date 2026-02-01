package models

type ChatRequest struct {
	Message    string `json:"message,omitempty"`
	AudioData  string `json:"audio_data,omitempty"`  // Base64 encoded audio for voice input
	AudioFormat string `json:"audio_format,omitempty"` // "wav", "mp3", "webm", etc.
}

type ChatResponse struct {
	Response         string                       `json:"response"`
	SQL              string                       `json:"sql,omitempty"`
	FormJSON         string                       `json:"form_json,omitempty"`
	ConfirmationCard *RegistrationConfirmationCard `json:"confirmation_card,omitempty"`
	ProposedForm     *ProposedFormCard             `json:"proposed_form,omitempty"`
	ResearchContent  string                       `json:"research_content,omitempty"`
}

// ProposedFormCard is sent when a form is generated from document upload; user must confirm before saving.
type ProposedFormCard struct {
	FormTemplate FormTemplate `json:"form_template"`
}

// RegistrationConfirmationCard is sent so the chat UI can show a review card before submitting.
type RegistrationConfirmationCard struct {
	FormName  string                   `json:"form_name"`
	UserType  string                   `json:"user_type"`
	Answers   map[string]interface{}   `json:"answers"`
	Fields    []FormField              `json:"fields"` // name + label for display
}

type SQLFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ChatHistory struct {
	Message   string `json:"message"`
	Response  string `json:"response"`
	Timestamp string `json:"timestamp"`
}

type SQLResult struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	Error    string          `json:"error,omitempty"`
	Filename string          `json:"filename,omitempty"`
}

type ResultFile struct {
	Filename  string        `json:"filename"`
	Query     string        `json:"query,omitempty"`
	Timestamp string        `json:"timestamp"`
	Columns   []string      `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	RowCount  int           `json:"row_count"`
	Error     string        `json:"error,omitempty"`
}

type ResultFileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	Format   string `json:"format"`
}

type GenerateHTMLRequest struct {
	Filename string `json:"filename"`
	Title    string `json:"title,omitempty"`
}

// Complaint flow models
type ComplaintState struct {
	ConversationID string                 `json:"conversation_id"`
	Step           string                 `json:"step"` // "start", "dialogue", "waiting_complaint", "executing", "complete"
	ComplaintText  string                 `json:"complaint_text,omitempty"`
	DialogueResult map[string]interface{} `json:"dialogue_result,omitempty"`
	InitialData    map[string]interface{} `json:"initial_data,omitempty"`
	ExchangeCount  int                    `json:"exchange_count"` // Track number of exchanges
	LastResponse   string                 `json:"last_response,omitempty"` // Store last AI response
}

// Voice recognition models
type VoiceProfile struct {
	UserID      string   `json:"user_id"`
	Name        string   `json:"name"`
	VoiceSamples []string `json:"voice_samples"` // Base64 encoded audio samples or file paths
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type VoiceRegistrationRequest struct {
	Name        string `json:"name" binding:"required"`
	AudioData   string `json:"audio_data" binding:"required"` // Base64 encoded audio
	AudioFormat string `json:"audio_format"` // "wav", "mp3", "webm", etc.
}

type VoiceRecognitionRequest struct {
	AudioData   string `json:"audio_data" binding:"required"` // Base64 encoded audio
	AudioFormat string `json:"audio_format"` // "wav", "mp3", "webm", etc.
}

type VoiceRecognitionResponse struct {
	Recognized bool   `json:"recognized"`
	UserID     string `json:"user_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Transcript string `json:"transcript,omitempty"`
	Intent     string `json:"intent,omitempty"` // "attendance", "punch_in", etc.
	Message    string `json:"message"`
}

// Form system models
type FormField struct {
	Name        string `json:"name"`         // Field identifier (e.g., "name", "age")
	Label       string `json:"label"`        // Display label (e.g., "Full Name")
	Type        string `json:"type"`         // Field type: "text", "email", "number", "tel", "date", "select", etc.
	Required    bool   `json:"required"`     // Whether field is required
	Placeholder string `json:"placeholder"`  // Placeholder text
	Options     []string `json:"options,omitempty"` // Options for select/radio fields
}

type FormTemplate struct {
	ID          string     `json:"id"`           // Unique identifier
	Name        string     `json:"name"`         // Form name (e.g., "Student Registration Form")
	Description string     `json:"description"`  // Form description
	UserType    string     `json:"user_type"`    // "student" or "staff"
	Fields      []FormField `json:"fields"`      // Form fields
	CreatedAt   string     `json:"created_at"`   // Creation timestamp
	UpdatedAt   string     `json:"updated_at"`   // Last update timestamp
	CreatedBy   string     `json:"created_by"`   // User who created the form
}

type FormAnswer struct {
	ID          string                 `json:"id"`           // Unique identifier
	FormID      string                 `json:"form_id"`      // Reference to FormTemplate
	FormName    string                 `json:"form_name"`    // Form name (denormalized for easy access)
	UserID      string                 `json:"user_id"`      // Student or staff ID
	UserType    string                 `json:"user_type"`    // "student" or "staff"
	Answers     map[string]interface{} `json:"answers"`      // Field name -> answer value
	SubmittedAt string                 `json:"submitted_at"` // Submission timestamp
	SubmittedBy string                 `json:"submitted_by"` // User who submitted
}

// RegistrationFlowState holds state for the "register a student" (or similar) chat flow
type RegConvTurn struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // message content
}

type RegistrationState struct {
	ConversationID    string                 `json:"conversation_id"`    // unique session id
	Step              string                 `json:"step"`                 // "selecting_form" | "gathering_fields" | "complete"
	FormID            string                 `json:"form_id,omitempty"`    // chosen form template id (internal, not shown to AI)
	FormName          string                 `json:"form_name,omitempty"`  // form name for context
	UserType          string                 `json:"user_type,omitempty"`  // student | staff from form
	GatheredAnswers   map[string]interface{} `json:"gathered_answers"`    // field name -> value so far
	ConversationHistory []RegConvTurn        `json:"conversation_history"` // full chat history for this session
	LastAIResponse    string                 `json:"last_ai_response,omitempty"`
	ExchangeCount     int                    `json:"exchange_count"`
	CreatedAt         string                 `json:"created_at,omitempty"`
}

