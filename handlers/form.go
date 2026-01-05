package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

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

