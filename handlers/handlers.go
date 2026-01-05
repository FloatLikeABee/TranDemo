package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/models"
	"idongivaflyinfa/service"
	"idongivaflyinfa/validation"

	"github.com/gin-gonic/gin"
)

// @title           Transfinder Form/Report Assistant API
// @version         1.0
// @description     Transfinder Form/Report Assistant API - Generate SQL queries using AI and execute them against SQL Server
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /

// @schemes   http https

type Handlers struct {
	db           *db.DB
	aiService    *ai.AIService
	sqlService   *service.SQLServerService
	sqlFilesDir  string
}

func New(db *db.DB, aiService *ai.AIService, sqlService *service.SQLServerService, sqlFilesDir string) *Handlers {
	return &Handlers{
		db:          db,
		aiService:   aiService,
		sqlService:  sqlService,
		sqlFilesDir: sqlFilesDir,
	}
}

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

	// Load SQL files
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
		hasReportKeywords := strings.Contains(lowerPrompt, "report") ||
			strings.Contains(lowerPrompt, "generate") ||
			strings.Contains(lowerPrompt, "create") ||
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
				userID = "default"
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

	// Store chat history (using a simple user ID for now)
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default"
	}

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

// UploadSQLFileHandler uploads a SQL file as reference
// @Summary      Upload SQL reference file
// @Description  Upload a SQL file that will be used as reference when generating SQL queries
// @Tags         SQL Files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "SQL file to upload"
// @Success      200   {object}  map[string]string  "File uploaded successfully"
// @Failure      400   {object}  map[string]string  "No file provided"
// @Failure      500   {object}  map[string]string  "Failed to store file"
// @Router       /api/sql/upload [post]
func (h *Handlers) UploadSQLFileHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Read file content
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	content := make([]byte, file.Size)
	_, err = src.Read(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Store in database
	if err := h.db.StoreSQLFile(file.Filename, string(content)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store SQL file"})
		return
	}

	// Also save to filesystem
	filePath := filepath.Join(h.sqlFilesDir, file.Filename)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		log.Printf("Warning: Failed to save file to filesystem: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "SQL file uploaded successfully", "filename": file.Filename})
}

// ListSQLFilesHandler lists all stored SQL reference files
// @Summary      List SQL reference files
// @Description  Get a list of all SQL files stored as references
// @Tags         SQL Files
// @Produce      json
// @Success      200  {object}  map[string][]string  "List of SQL file names"
// @Failure      500  {object}  map[string]string     "Failed to load files"
// @Router       /api/sql/files [get]
func (h *Handlers) ListSQLFilesHandler(c *gin.Context) {
	sqlFiles, err := h.db.GetSQLFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load SQL files"})
		return
	}

	// Return just names
	names := make([]string, len(sqlFiles))
	for i, f := range sqlFiles {
		names[i] = f.Name
	}

	c.JSON(http.StatusOK, gin.H{"files": names})
}

// ExecuteSQLHandler executes a SQL query against SQL Server
// @Summary      Execute SQL query
// @Description  Execute a SQL query against the configured SQL Server and optionally save the results
// @Tags         SQL Execution
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "SQL execution request"  example({"sql": "SELECT * FROM users", "save": true, "format": "json"})
// @Success      200      {object}  models.SQLResult  "Query execution result"
// @Failure      400      {object}  map[string]string  "Invalid request"
// @Failure      503      {object}  map[string]string  "SQL Server not configured"
// @Failure      500      {object}  map[string]string  "Query execution error"
// @Router       /api/sql/execute [post]
func (h *Handlers) ExecuteSQLHandler(c *gin.Context) {
	var req struct {
		SQL    string `json:"sql" example:"SELECT * FROM users"`
		Save   bool   `json:"save" example:"true"`
		Format string `json:"format" example:"json"` // "json" or "csv"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if h.sqlService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL Server service is not configured"})
		return
	}

	format := req.Format
	if format == "" {
		format = "json" // Default to JSON
	}
	if format != "json" && format != "csv" {
		format = "json"
	}

	result, err := h.sqlService.ExecuteQueryWithSave(req.SQL, format, req.Save)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "result": result})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListResultFilesHandler lists all result files
// @Summary      List result files
// @Description  Get a list of all saved SQL query result files (JSON/CSV)
// @Tags         Results
// @Produce      json
// @Success      200  {object}  map[string][]models.ResultFileInfo  "List of result files"
// @Failure      503  {object}  map[string]string                   "SQL Server not configured"
// @Failure      500  {object}  map[string]string                  "Failed to list files"
// @Router       /api/results/files [get]
func (h *Handlers) ListResultFilesHandler(c *gin.Context) {
	if h.sqlService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL Server service is not configured"})
		return
	}

	resultsStorage := h.sqlService.GetResultsStorage()
	if resultsStorage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Results storage is not initialized"})
		return
	}

	files, err := resultsStorage.ListResultFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list files: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// GetResultFileHandler retrieves a specific result file
// @Summary      Get result file
// @Description  Get the complete content of a specific result file by filename
// @Tags         Results
// @Produce      json
// @Param        filename  path      string  true  "Result file name"
// @Success      200       {object}  models.ResultFile  "Result file content"
// @Failure      400       {object}  map[string]string   "Filename required"
// @Failure      404       {object}  map[string]string   "File not found"
// @Failure      503       {object}  map[string]string    "SQL Server not configured"
// @Router       /api/results/file/{filename} [get]
func (h *Handlers) GetResultFileHandler(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	if h.sqlService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL Server service is not configured"})
		return
	}

	resultsStorage := h.sqlService.GetResultsStorage()
	if resultsStorage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Results storage is not initialized"})
		return
	}

	resultFile, err := resultsStorage.GetResultFile(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("File not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, resultFile)
}

// GenerateHTMLHandler generates an HTML page from a result file
// @Summary      Generate HTML page
// @Description  Use AI to generate a professional HTML page displaying the content of a result file
// @Tags         Results
// @Accept       json
// @Produce      json
// @Param        request  body      models.GenerateHTMLRequest  true  "HTML generation request"
// @Success      200      {object}  map[string]string  "HTML page generated successfully"
// @Failure      400      {object}  map[string]string  "Invalid request"
// @Failure      404      {object}  map[string]string  "Result file not found"
// @Failure      503      {object}  map[string]string  "SQL Server not configured"
// @Failure      500      {object}  map[string]string  "Failed to generate HTML"
// @Router       /api/results/generate-html [post]
func (h *Handlers) GenerateHTMLHandler(c *gin.Context) {
	var req models.GenerateHTMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if h.sqlService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL Server service is not configured"})
		return
	}

	resultsStorage := h.sqlService.GetResultsStorage()
	if resultsStorage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Results storage is not initialized"})
		return
	}

	// Load the result file
	resultFile, err := resultsStorage.GetResultFile(req.Filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("File not found: %v", err)})
		return
	}

	// Generate title if not provided
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("SQL Query Results - %s", req.Filename)
	}

	// Generate HTML using AI
	html, err := h.aiService.GenerateHTMLPage(resultFile, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate HTML: %v", err)})
		return
	}

	// Generate HTML filename from result filename
	htmlFilename := req.Filename
	ext := filepath.Ext(htmlFilename)
	if ext != "" {
		htmlFilename = htmlFilename[:len(htmlFilename)-len(ext)]
	}
	htmlFilename += ".html"

	// Save HTML file to sites directory
	savedFilename, err := resultsStorage.SaveHTMLFile(htmlFilename, []byte(html))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save HTML file: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "HTML page generated successfully",
		"filename":   savedFilename,
		"html_path": fmt.Sprintf("/api/results/html/%s", savedFilename),
	})
}

// ServeHTMLHandler serves a generated HTML page
// @Summary      Serve HTML page
// @Description  Serve a previously generated HTML page from the sites directory
// @Tags         Results
// @Produce      text/html
// @Param        filename  path      string  true  "HTML file name"
// @Success      200       {string}  string  "HTML content"
// @Failure      400       {object}  map[string]string  "Filename required"
// @Failure      404       {object}  map[string]string  "HTML file not found"
// @Failure      503       {object}  map[string]string   "SQL Server not configured"
// @Router       /api/results/html/{filename} [get]
func (h *Handlers) ServeHTMLHandler(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	if h.sqlService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL Server service is not configured"})
		return
	}

	resultsStorage := h.sqlService.GetResultsStorage()
	if resultsStorage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Results storage is not initialized"})
		return
	}

	htmlPath := resultsStorage.GetHTMLFilePath(filename)
	
	// Check if file exists
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "HTML file not found"})
		return
	}

	// Serve the HTML file
	c.File(htmlPath)
}

// HealthHandler checks the health status of the service
// @Summary      Health check
// @Description  Check the health status of all services (database, AI service, SQL Server)
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string  "Service health status"
// @Router       /health [get]
func (h *Handlers) HealthHandler(c *gin.Context) {
	status := gin.H{
		"status":      "healthy",
		"db":          "connected",
		"ai_service":  "ready",
		"sql_server":  "not_configured",
	}

	if h.sqlService != nil && h.sqlService.IsConnected() {
		status["sql_server"] = "connected"
	}

	c.JSON(http.StatusOK, status)
}

