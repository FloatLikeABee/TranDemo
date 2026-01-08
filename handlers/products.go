package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// ProductFileInfo represents information about a product HTML file
type ProductFileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	Type     string `json:"type"` // "form" or "result"
}

// ListProductsHandler lists all HTML files in the products folder
// @Summary      List product files
// @Description  Get a list of all HTML files in the products folder
// @Tags         Products
// @Produce      json
// @Success      200  {object}  map[string][]ProductFileInfo  "List of product files"
// @Failure      500  {object}  map[string]string            "Failed to list files"
// @Router       /api/products/files [get]
func (h *Handlers) ListProductsHandler(c *gin.Context) {
	productsDir := "products"
	
	// Ensure directory exists
	if err := os.MkdirAll(productsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create products directory: %v", err)})
		return
	}

	files, err := os.ReadDir(productsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read products directory: %v", err)})
		return
	}

	var productFiles []ProductFileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only include HTML files
		if filepath.Ext(file.Name()) != ".html" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Determine file type based on filename
		fileType := "result"
		if len(file.Name()) >= 5 && file.Name()[:5] == "form_" {
			fileType = "form"
		}

		productFiles = append(productFiles, ProductFileInfo{
			Filename: file.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
			Type:     fileType,
		})
	}

	// Sort by modified time, newest first
	sort.Slice(productFiles, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, productFiles[i].Modified)
		timeJ, _ := time.Parse(time.RFC3339, productFiles[j].Modified)
		return timeI.After(timeJ)
	})

	c.JSON(http.StatusOK, gin.H{"files": productFiles})
}

// ServeProductHandler serves a specific HTML file from the products folder
// @Summary      Serve product file
// @Description  Serve a specific HTML file from the products folder
// @Tags         Products
// @Produce      text/html
// @Param        filename  path      string  true  "Product file name"
// @Success      200       {string}  string  "HTML content"
// @Failure      404       {object}  map[string]string  "File not found"
// @Router       /products/{filename} [get]
func (h *Handlers) ServeProductHandler(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Security: prevent directory traversal
	if filepath.Base(filename) != filename {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filename"})
		return
	}

	// Ensure it's an HTML file
	if filepath.Ext(filename) != ".html" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only HTML files are allowed"})
		return
	}

	filePath := filepath.Join("products", filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

