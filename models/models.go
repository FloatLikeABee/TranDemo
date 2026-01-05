package models

type ChatRequest struct {
	Message string `json:"message"`
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

