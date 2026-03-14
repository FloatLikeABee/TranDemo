package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

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

// ServeResultPDFHandler serves a report PDF file for download
// @Summary      Download report PDF
// @Description  Serve a report PDF file by filename (e.g. report_123.pdf)
// @Tags         Results
// @Param        filename  path      string  true  "PDF filename"
// @Success      200      file     binary  "PDF file"
// @Failure      400      {object}  map[string]string  "Filename required or not a PDF"
// @Failure      404      {object}  map[string]string  "File not found"
// @Failure      503      {object}  map[string]string  "SQL Server not configured"
// @Router       /api/results/pdf/{filename} [get]
func (h *Handlers) ServeResultPDFHandler(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" || filepath.Ext(filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid PDF filename is required"})
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

	filePath := filepath.Join(resultsStorage.GetResultsDir(), filename)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Content-Type", "application/pdf")
	c.File(filePath)
}

