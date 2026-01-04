package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

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

