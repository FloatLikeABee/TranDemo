# SQL Results Storage and HTML Generation Features

## Overview

The application now includes comprehensive result file management and AI-powered HTML page generation for SQL query results.

## Features

### 1. Result File Storage

SQL query results can be automatically saved to files in JSON or CSV format with unique, trackable filenames.

**File Naming Convention:**
- Format: `result_YYYYMMDD_HHMMSS_NANOSECONDS.{json|csv}`
- Example: `result_20231214_150405_1699986245123456789.json`
- Includes timestamp and nanoseconds for uniqueness and tracking

**Storage Location:**
- Default: `./results/` directory
- Configurable via `RESULTS_DIR` environment variable

### 2. File Formats

#### JSON Format
Stores complete result metadata including:
- Original SQL query
- Column names
- All data rows
- Row count
- Timestamp
- Error information (if any)

#### CSV Format
Standard CSV format with:
- Header row with column names
- Data rows
- Easy to import into Excel, databases, etc.

### 3. API Endpoints

#### List Result Files
```http
GET /api/results/files
```

Returns list of all result files with metadata:
- Filename
- File size
- Last modified date
- Format (json/csv)

#### Get Result File
```http
GET /api/results/file/:filename
```

Returns the complete result file data including all rows.

#### Execute SQL with Save
```http
POST /api/sql/execute
Content-Type: application/json

{
  "sql": "SELECT * FROM users",
  "save": true,
  "format": "json"
}
```

Executes SQL and optionally saves the result.

### 4. AI-Powered HTML Generation

Generate professional, responsive HTML pages from result files using AI.

#### Generate HTML Page
```http
POST /api/results/generate-html
Content-Type: application/json

{
  "filename": "result_20231214_150405_1234567890.json",
  "title": "User Report"
}
```

**Features of Generated HTML:**
- Professional, modern design
- Responsive layout (mobile-friendly)
- Sticky table headers
- Zebra striping for rows
- Hover effects
- Metadata display (row count, columns, timestamp)
- SQL query display (if available)
- Professional color scheme
- Clean typography
- Self-contained (no external dependencies)

#### View HTML Page
```http
GET /api/results/html/:filename
```

Serves the generated HTML page directly in the browser.

### 5. Workflow Example

1. **Execute SQL and Save Result:**
   ```bash
   curl -X POST http://localhost:8080/api/sql/execute \
     -H "Content-Type: application/json" \
     -d '{
       "sql": "SELECT id, name, email FROM users LIMIT 100",
       "save": true,
       "format": "json"
     }'
   ```

2. **Generate HTML Page:**
   ```bash
   curl -X POST http://localhost:8080/api/results/generate-html \
     -H "Content-Type: application/json" \
     -d '{
       "filename": "result_20231214_150405_1234567890.json",
       "title": "User List Report"
     }'
   ```

3. **View in Browser:**
   Open: `http://localhost:8080/api/results/html/result_20231214_150405_1234567890.html`

### 6. File Tracking

The unique filename format allows for:
- Chronological sorting
- Easy identification of when queries were run
- Tracking query execution history
- Integration with external systems
- Audit trails

### 7. Integration Points

- **SQL Execution**: Results automatically saved when `save: true`
- **AI Generation**: HTML pages generated on-demand
- **File Management**: List, view, and serve result files
- **Frontend Integration**: HTML pages can be embedded or linked

## Configuration

Add to your environment or config:
```bash
RESULTS_DIR=./results  # Directory for storing result files
```

## Best Practices

1. **File Management**: Regularly clean up old result files if storage is a concern
2. **HTML Generation**: Generate HTML pages on-demand rather than pre-generating
3. **Format Selection**: Use JSON for complex data, CSV for simple tabular data
4. **Naming**: The automatic naming ensures uniqueness - don't manually rename files
5. **Security**: Result files may contain sensitive data - secure the results directory appropriately

## Technical Details

- **Storage**: Files stored on local filesystem
- **AI Model**: Uses configured Gemini model for HTML generation
- **Format Support**: JSON and CSV formats
- **HTML**: Self-contained HTML with embedded CSS
- **Responsive**: Mobile-first design with proper table scrolling

