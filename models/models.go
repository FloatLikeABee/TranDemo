package models

type ChatRequest struct {
	Message    string `json:"message,omitempty"`
	AudioData  string `json:"audio_data,omitempty"`  // Base64 encoded audio for voice input
	AudioFormat string `json:"audio_format,omitempty"` // "wav", "mp3", "webm", etc.
}

type ChatResponse struct {
	Response string `json:"response"`
	SQL      string `json:"sql,omitempty"`
	FormJSON string `json:"form_json,omitempty"`
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

