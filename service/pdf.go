package service

import (
	"fmt"
	"path/filepath"
	"time"

	"idongivaflyinfa/models"

	"github.com/jung-kurt/gofpdf/v2"
)

const reportMaxCellWidth = 60  // truncate cell text for PDF
const reportTableMaxRows = 10 // already limited by caller (report uses 10 rows)

// WriteResultPDF writes a SQL result to a PDF file with a simple table. Returns the filename (base name) and error.
func WriteResultPDF(resultsDir string, result *models.SQLResult, title string) (filename string, err error) {
	if result == nil || len(result.Columns) == 0 {
		return "", fmt.Errorf("no result data to write")
	}

	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for wide tables
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "", 9)
	colCount := len(result.Columns)
	if colCount == 0 {
		return "", fmt.Errorf("no columns")
	}

	// Column widths: distribute evenly, or use fixed small width for many columns
	pageWidth := 277.0 // A4 landscape width in mm minus margins
	margin := 10.0
	usable := pageWidth - 2*margin
	cellW := usable / float64(colCount)
	if cellW > 40 {
		cellW = 40
	}
	cellH := 6.0

	// Header row
	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Arial", "B", 9)
	for _, col := range result.Columns {
		text := truncateForCell(col, reportMaxCellWidth)
		pdf.CellFormat(cellW, cellH, text, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	// Data rows
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	fill := false
	rowCount := len(result.Rows)
	if rowCount > reportTableMaxRows {
		rowCount = reportTableMaxRows
	}
	for i := 0; i < rowCount; i++ {
		row := result.Rows[i]
		if fill {
			pdf.SetFillColor(245, 245, 245)
		}
		for j, val := range row {
			text := ""
			if val != nil {
				text = fmt.Sprintf("%v", val)
			}
			text = truncateForCell(text, reportMaxCellWidth)
			pdf.CellFormat(cellW, cellH, text, "1", 0, "L", fill, 0, "")
			if j >= colCount-1 {
				break
			}
		}
		if len(row) < colCount {
			for k := len(row); k < colCount; k++ {
				pdf.CellFormat(cellW, cellH, "", "1", 0, "L", fill, 0, "")
			}
		}
		pdf.Ln(-1)
		fill = !fill
	}

	// Footer: row count
	pdf.Ln(4)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total rows: %d", rowCount), "", 1, "L", false, 0, "")

	outFilename := fmt.Sprintf("report_%s.pdf", uniqueReportSuffix())
	outPath := filepath.Join(resultsDir, outFilename)
	if err := pdf.OutputFileAndClose(outPath); err != nil {
		return "", fmt.Errorf("failed to write PDF: %w", err)
	}
	return outFilename, nil
}

func truncateForCell(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func uniqueReportSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
}
