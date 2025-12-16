package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"idongivaflyinfa/models"
)

type ResultsStorage struct {
	resultsDir string
	sitesDir   string
}

func NewResultsStorage(resultsDir string, sitesDir string) (*ResultsStorage, error) {
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}

	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sites directory: %w", err)
	}

	return &ResultsStorage{
		resultsDir: resultsDir,
		sitesDir:   sitesDir,
	}, nil
}

// GenerateFileName creates a unique filename with timestamp and hash
func (r *ResultsStorage) GenerateFileName(format string) string {
	timestamp := time.Now().Format("20060102_150405")
	nanos := time.Now().UnixNano()
	return fmt.Sprintf("result_%s_%d.%s", timestamp, nanos, format)
}

// SaveResultAsJSON saves SQL result as JSON file
func (r *ResultsStorage) SaveResultAsJSON(result *models.SQLResult, query string) (string, error) {
	filename := r.GenerateFileName("json")
	filePath := filepath.Join(r.resultsDir, filename)

	// Create result metadata
	resultData := models.ResultFile{
		Query:     query,
		Timestamp: time.Now().Format(time.RFC3339),
		Columns:   result.Columns,
		Rows:      result.Rows,
		RowCount:  len(result.Rows),
		Error:     result.Error,
	}

	data, err := json.MarshalIndent(resultData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file: %w", err)
	}

	return filename, nil
}

// SaveResultAsCSV saves SQL result as CSV file
func (r *ResultsStorage) SaveResultAsCSV(result *models.SQLResult, query string) (string, error) {
	filename := r.GenerateFileName("csv")
	filePath := filepath.Join(r.resultsDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(result.Columns); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, val := range row {
			if val == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return filename, nil
}

// GetResultFile reads a result file
func (r *ResultsStorage) GetResultFile(filename string) (*models.ResultFile, error) {
	filePath := filepath.Join(r.resultsDir, filename)

	// Check if it's JSON
	if filepath.Ext(filename) == ".json" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		var result models.ResultFile
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		return &result, nil
	}

	// For CSV, read and convert to ResultFile
	if filepath.Ext(filename) == ".csv" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open CSV file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}

		if len(records) == 0 {
			return &models.ResultFile{
				Filename:  filename,
				Columns:   []string{},
				Rows:      [][]interface{}{},
				RowCount:  0,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		}

		// First row is header
		columns := records[0]
		rows := make([][]interface{}, len(records)-1)

		for i, record := range records[1:] {
			row := make([]interface{}, len(record))
			for j, val := range record {
				row[j] = val
			}
			rows[i] = row
		}

		return &models.ResultFile{
			Filename:  filename,
			Columns:   columns,
			Rows:      rows,
			RowCount:  len(rows),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	return nil, fmt.Errorf("unsupported file format")
}

// ListResultFiles returns all result files
func (r *ResultsStorage) ListResultFiles() ([]models.ResultFileInfo, error) {
	files, err := os.ReadDir(r.resultsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read results directory: %w", err)
	}

	var resultFiles []models.ResultFileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".json" && ext != ".csv" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		resultFiles = append(resultFiles, models.ResultFileInfo{
			Filename:    file.Name(),
			Size:        info.Size(),
			Modified:    info.ModTime().Format(time.RFC3339),
			Format:      ext[1:], // Remove the dot
		})
	}

	return resultFiles, nil
}

// GetResultFilePath returns the full path to a result file
func (r *ResultsStorage) GetResultFilePath(filename string) string {
	return filepath.Join(r.resultsDir, filename)
}

// SaveHTMLFile saves an HTML file to the sites directory
func (r *ResultsStorage) SaveHTMLFile(filename string, content []byte) (string, error) {
	// Ensure filename has .html extension
	if filepath.Ext(filename) != ".html" {
		filename += ".html"
	}
	
	filePath := filepath.Join(r.sitesDir, filename)
	
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML file: %w", err)
	}
	
	return filename, nil
}

// GetHTMLFilePath returns the full path to an HTML file in sites directory
func (r *ResultsStorage) GetHTMLFilePath(filename string) string {
	// Ensure filename has .html extension
	if filepath.Ext(filename) != ".html" {
		filename += ".html"
	}
	return filepath.Join(r.sitesDir, filename)
}

