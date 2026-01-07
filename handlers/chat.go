package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/models"
	"idongivaflyinfa/validation"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat requests to generate SQL queries
// @Summary      Generate SQL query from natural language
// @Description  Send a message describing what SQL query you need, and the AI will generate it based on reference SQL files
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        request  body      models.ChatRequest  true  "Chat request with message"
// @Header       200      {string}  X-User-ID          "Optional user ID for chat history"
// @Success      200      {object}  models.ChatResponse "Generated SQL query"
// @Failure      400      {object}  map[string]string   "Invalid request"
// @Failure      500      {object}  map[string]string   "Internal server error"
// @Router       /api/chat [post]
func (h *Handlers) ChatHandler(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get user ID FIRST - before anything else
	// Use "admin" as default since we don't have a user system yet
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}

	// PRIORITY 0: Check if this is a voice input
	if req.AudioData != "" {
		log.Printf("[CHAT HANDLER] Voice input detected from user: %s", userID)
		response, err := h.HandleVoiceChat(c, userID, req.AudioData)
		if err != nil {
			log.Printf("[CHAT HANDLER] Error handling voice chat: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process voice: %v", err)})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// PRIORITY 0.5: Correct spelling errors in user message
	correctedMessage, err := h.aiService.CorrectSpelling(req.Message)
	if err != nil {
		log.Printf("[CHAT HANDLER] Error correcting spelling: %v, using original message", err)
		correctedMessage = req.Message
	} else if correctedMessage != req.Message {
		log.Printf("[CHAT HANDLER] Spelling corrected: '%s' -> '%s'", req.Message, correctedMessage)
		req.Message = correctedMessage
	}

	log.Printf("[CHAT HANDLER] User: %s, Message: %s", userID, req.Message)

	// PRIORITY 1: Check if user has an active complaint conversation (simplified check)
	// Just check if there's any complaint state with a conversation_id - if yes, continue the session
	complaintState, err := h.db.GetComplaintStateByUserID(userID)
	if err == nil && complaintState != nil {
		// If we have a conversation_id and it's not complete, continue the session
		if complaintState.ConversationID != "" && complaintState.Step != "complete" {
			log.Printf("[CHAT HANDLER] User %s has active complaint conversation (conversationID: %s, step: %s, exchanges: %d)",
				userID, complaintState.ConversationID, complaintState.Step, complaintState.ExchangeCount)
			response, err := h.handleComplaintFlow(c, userID, req.Message)
			if err != nil {
				log.Printf("[CHAT HANDLER] Error continuing complaint flow: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process complaint: %v", err)})
				return
			}
			c.JSON(http.StatusOK, response)
			return
		} else if complaintState.Step == "complete" {
			log.Printf("[CHAT HANDLER] Complaint session is complete for user %s, starting new flow", userID)
		}
	} else {
		log.Printf("[CHAT HANDLER] No complaint state found for user %s (error: %v)", userID, err)
	}

	// PRIORITY 2: Check if this is a NEW complaint request
	if isComplaintRequest(req.Message) {
		log.Printf("[CHAT HANDLER] Detected NEW complaint request from user %s", userID)
		response, err := h.handleComplaintFlow(c, userID, req.Message)
		if err != nil {
			log.Printf("[CHAT HANDLER] Error handling complaint flow: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process complaint: %v", err)})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Load SQL files (only if not in complaint flow)
	sqlFiles, err := h.db.GetSQLFiles()
	if err != nil {
		log.Printf("Error loading SQL files from DB: %v", err)
		// Try loading from directory as fallback
		sqlFiles, err = h.db.LoadSQLFilesFromDir(h.sqlFilesDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load SQL files"})
			return
		}
	}

	// Check if this is a form generation request  TODO: change this to AI decision
	lowerPrompt := strings.ToLower(req.Message)
	isFormRequest := (strings.Contains(lowerPrompt, "create") && strings.Contains(lowerPrompt, "form")) ||
		strings.Contains(lowerPrompt, "i want a new form") ||
		strings.Contains(lowerPrompt, "generate a form") ||
		strings.Contains(lowerPrompt, "make a form") ||
		strings.Contains(lowerPrompt, "build a form") ||
		(strings.Contains(lowerPrompt, "form") && (strings.Contains(lowerPrompt, "new") || strings.Contains(lowerPrompt, "create")))

	var responseText string
	var sql string
	var formJSON string

	if isFormRequest {
		// Generate form JSON
		formJSON, err = h.aiService.GenerateForm(req.Message)
		if err != nil {
			log.Printf("Error generating form: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate form: %v", err)})
			return
		}

		// Generate form HTML page
		html, err := h.aiService.GenerateFormHTMLPage(formJSON)
		if err != nil {
			log.Printf("Error generating form HTML: %v", err)
			// Continue even if HTML generation fails
		} else {
			// Save HTML to products folder
			productsDir := "products"
			if err := os.MkdirAll(productsDir, 0755); err != nil {
				log.Printf("Error creating products directory: %v", err)
			} else {
				// Generate HTML filename with timestamp
				timestamp := time.Now().Format("20060102_150405")
				htmlFilename := fmt.Sprintf("form_%s.html", timestamp)
				htmlPath := filepath.Join(productsDir, htmlFilename)

				if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
					log.Printf("Error saving form HTML file: %v", err)
				} else {
					log.Printf("Form HTML page saved to: %s", htmlPath)
				}
			}
		}

		responseText = fmt.Sprintf("Here's the form JSON based on your request:\n\n%s", formJSON)
	} else {
		// Check if the prompt contains report-related keywords
		hasReportKeywords := strings.Contains(lowerPrompt, "generate") ||
			strings.Contains(lowerPrompt, "create report") ||
			strings.Contains(lowerPrompt, "i want a report") ||
			strings.Contains(lowerPrompt, "i need to make") ||
			strings.Contains(lowerPrompt, "i need a report") ||
			strings.Contains(lowerPrompt, "make a report") ||
			strings.Contains(lowerPrompt, "generate a report") ||
			strings.Contains(lowerPrompt, "create a report")

		if !hasReportKeywords {
			// Check if the prompt makes sense (not gibberish)
			if !validation.IsValidPrompt(req.Message) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "The request appears to be invalid or gibberish. Please provide a meaningful message."})
				return
			}

			// If it's a valid prompt but not a report request, treat it as a general chat
			chatResponse, err := h.aiService.GenerateChatResponse(req.Message)
			if err != nil {
				log.Printf("Error generating chat response: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate response: %v", err)})
				return
			}

			responseText = chatResponse

			// Store chat history
			userID := c.GetHeader("X-User-ID")
			if userID == "" {
				userID = "admin"
			}

			if err := h.db.StoreChatHistory(userID, req.Message, responseText); err != nil {
				log.Printf("Error storing chat history: %v", err)
			}

			response := models.ChatResponse{
				Response: responseText,
				SQL:      "",
			}

			log.Printf("Sending chat response to client")
			c.JSON(http.StatusOK, response)
			return
		}

		// Generate SQL using AI
		sql, err = h.aiService.GenerateSQL(req.Message, sqlFiles)
		if err != nil {
			log.Printf("Error generating SQL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate SQL: %v", err)})
			return
		}

		// Ensure SQL is not empty
		if strings.TrimSpace(sql) == "" {
			log.Printf("Generated SQL is empty")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Generated SQL query is empty"})
			return
		}

		log.Printf("SQL generated successfully, length: %d", len(sql))

		// Check if SQL starts with "with" (case-insensitive)
		sqlTrimmed := strings.TrimSpace(sql)
		finalSQL := sql
		if !strings.HasPrefix(strings.ToLower(sqlTrimmed), "with") {
			// Prepend StudentReportSqlHead
			finalSQL = config.StudentReportSqlHead + "\n" + sql
			log.Printf("Prepended StudentReportSqlHead to SQL")
		}

		responseText = fmt.Sprintf("Here's the SQL query based on your request:\n\n%s", sql)
		log.Printf("Prepared response text, length: %d", len(responseText))

		// Execute SQL and save result in background (don't block response)
		// Check if SQL service is available before starting goroutine
		if h.sqlService == nil {
			log.Printf("SQL service is nil, skipping background SQL execution and HTML generation")
		} else {
			// Capture variables needed for the goroutine
			sqlService := h.sqlService
			aiService := h.aiService
			go func() {
				log.Printf("Background goroutine started for SQL execution")
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in background SQL execution: %v", r)
					}
				}()

				resultsStorage := sqlService.GetResultsStorage()
				if resultsStorage == nil {
					log.Printf("Results storage is nil, skipping background execution")
					return
				}

				log.Printf("Starting SQL execution with query length: %d", len(finalSQL))
				// Execute SQL and save as JSON
				sqlResult, err := sqlService.ExecuteQueryWithSave(finalSQL, "json", true)
				if err != nil {
					log.Printf("Error executing SQL: %v", err)
					return
				}
				if sqlResult.Error != "" {
					log.Printf("SQL execution error: %s", sqlResult.Error)
					return
				}
				if sqlResult.Filename == "" {
					log.Printf("No filename returned from SQL execution")
					return
				}
				log.Printf("SQL executed successfully, result file: %s", sqlResult.Filename)

				// Load the ResultFile
				resultFile, err := resultsStorage.GetResultFile(sqlResult.Filename)
				if err != nil {
					log.Printf("Error loading result file: %v", err)
					return
				}
				log.Printf("Result file loaded, rows: %d", resultFile.RowCount)

				// Generate HTML page
				title := fmt.Sprintf("SQL Query Results - %s", sqlResult.Filename)
				log.Printf("Generating HTML page with title: %s", title)
				html, err := aiService.GenerateHTMLPage(resultFile, title)
				if err != nil {
					log.Printf("Error generating HTML: %v", err)
					return
				}
				log.Printf("HTML generated successfully, length: %d", len(html))

				// Save HTML to products folder
				productsDir := "products"
				if err := os.MkdirAll(productsDir, 0755); err != nil {
					log.Printf("Error creating products directory: %v", err)
					return
				}
				// Generate HTML filename from result filename
				htmlFilename := sqlResult.Filename
				ext := filepath.Ext(htmlFilename)
				if ext != "" {
					htmlFilename = htmlFilename[:len(htmlFilename)-len(ext)]
				}
				htmlFilename += ".html"
				htmlPath := filepath.Join(productsDir, htmlFilename)

				if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
					log.Printf("Error saving HTML file: %v", err)
				} else {
					log.Printf("HTML page saved successfully to: %s", htmlPath)
				}
			}()
		}
	}

	// Store chat history (userID already set at the beginning of the function)

	if err := h.db.StoreChatHistory(userID, req.Message, responseText); err != nil {
		log.Printf("Error storing chat history: %v", err)
	}

	response := models.ChatResponse{
		Response: responseText,
		SQL:      sql,
	}
	if formJSON != "" {
		response.FormJSON = formJSON
	}

	log.Printf("Sending response to client")
	c.JSON(http.StatusOK, response)
	log.Printf("Response sent successfully")
}

