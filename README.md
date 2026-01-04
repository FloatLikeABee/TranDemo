# Transfinder Form/Report Assistant

A full-stack application that uses AI (Gemini) to generate SQL queries based on user prompts and reference SQL files.

## Features

- 🤖 AI-powered SQL generation using Google Gemini
- 📁 SQL file storage and management with BadgerDB
- ⚡ In-memory caching for improved performance
- 💬 Chat-based interface for SQL queries
- 📱 Multi-device compatible React frontend

## Architecture

### Backend (Go)
- **Database**: BadgerDB (embedded key-value store)
- **Cache**: go-cache for in-memory caching
- **AI**: Google Gemini API
- **SQL Server**: Microsoft SQL Server connection service
- **API**: Gin web framework

### Package Structure
- `config/` - Configuration management
- `db/` - BadgerDB operations
- `cache/` - In-memory caching
- `ai/` - Gemini AI integration
- `service/` - SQL Server connection and execution service
- `handlers/` - HTTP request handlers
- `models/` - Shared data models

### Frontend (React)
- Modern chat interface
- Responsive design for all devices
- Real-time chat experience

## Setup

### Prerequisites
- Go 1.21 or higher
- Node.js 18+ and npm

### Backend Setup

1. Install Go dependencies:
```bash
go mod download
```

2. Create SQL files directory (optional):
```bash
mkdir sql_files
# Add your reference SQL files here
```

3. Run the server:
```bash
# macOS/Linux (use this instead of go run to avoid LC_UUID error)
./start.sh

# Or using Make
make run

# Or build and run manually (with external linking for macOS)
go build -ldflags="-linkmode=external" -o tran_demo main.go
./tran_demo

# Windows
go run main.go
```

**Note for macOS users:** Do NOT use `go run main.go` directly - it will fail with a "missing LC_UUID" error due to CGO dependencies. Always use `./start.sh` or `make run` instead.

The server will start on port 8080 by default.

### Frontend Setup

1. Navigate to frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm start
```

4. Build for production:
```bash
npm run build
```

## Configuration

**Application Settings:**
- `PORT`: Server port (default: 8080)
- `GEMINI_API_KEY`: Your Gemini API key (default: already set in code)
- `GEMINI_MODEL`: Model name (default: gemini-1.5-flash-latest)
- `DB_PATH`: BadgerDB storage path (default: ./data/badger)
- `SQL_FILES_DIR`: SQL files directory (default: ./sql_files)

**SQL Server Settings (Optional):**
- `SQL_SERVER`: SQL Server hostname (default: localhost)
- `SQL_PORT`: SQL Server port (default: 1433)
- `SQL_DATABASE`: Database name
- `SQL_USER`: SQL Server username
- `SQL_PASSWORD`: SQL Server password
- `SQL_ENCRYPT`: Enable encryption (default: true)

**Results Storage:**
- `RESULTS_DIR`: Directory for storing SQL query results (default: ./results)
- `SITES_DIR`: Directory for storing generated HTML pages (default: ./sites)

If SQL Server credentials are not provided, the service will start without SQL Server functionality.

## API Documentation

### Swagger UI

Interactive API documentation is available via Swagger UI:

```
http://localhost:8080/swagger/index.html
```

The Swagger UI provides:
- Complete API documentation
- Interactive endpoint testing
- Request/response examples
- Schema definitions

### Generating Swagger Docs

To regenerate Swagger documentation after code changes:

```bash
# Install Swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init
```

## API Endpoints

- `POST /api/chat` - Send a chat message and get SQL response
- `POST /api/sql/upload` - Upload a SQL file as reference
- `GET /api/sql/files` - List all stored SQL files
- `POST /api/sql/execute` - Execute SQL query against SQL Server (requires SQL Server configuration)
- `GET /health` - Health check (includes service status)

### Execute SQL Endpoint

The `/api/sql/execute` endpoint allows you to execute SQL queries against your SQL Server:

```bash
curl -X POST http://localhost:8080/api/sql/execute \
  -H "Content-Type: application/json" \
  -d '{"sql": "SELECT * FROM users LIMIT 10", "save": true, "format": "json"}'
```

Request parameters:
- `sql` (required): SQL query to execute
- `save` (optional): Whether to save the result to a file (default: false)
- `format` (optional): File format - "json" or "csv" (default: "json")

Response format:
```json
{
  "columns": ["id", "name", "email"],
  "rows": [
    ["1", "John Doe", "john@example.com"],
    ["2", "Jane Smith", "jane@example.com"]
  ],
  "filename": "result_20231214_150405_1234567890.json"
}
```

### Result File Management Endpoints

- `GET /api/results/files` - List all result files
- `GET /api/results/file/:filename` - Get a specific result file
- `POST /api/results/generate-html` - Generate an HTML page from a result file
- `GET /api/results/html/:filename` - View the generated HTML page

### Generate HTML Page

Generate a professional HTML page from a result file:

```bash
curl -X POST http://localhost:8080/api/results/generate-html \
  -H "Content-Type: application/json" \
  -d '{"filename": "result_20231214_150405_1234567890.json", "title": "User Report"}'
```

Response:
```json
{
  "message": "HTML page generated successfully",
  "filename": "result_20231214_150405_1234567890.html",
  "html_path": "/api/results/html/result_20231214_150405_1234567890.html"
}
```

Then view it at: `http://localhost:8080/api/results/html/result_20231214_150405_1234567890.html`

### Access Swagger UI

Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

You can test all endpoints directly from the Swagger UI interface.

## Usage

1. Upload SQL reference files via the API or place them in the `sql_files` directory
2. Send chat messages with your SQL requirements
3. The AI will generate SQL queries based on your prompts and the reference files

## Project Structure

```
.
├── main.go              # Main application entry point
├── go.mod               # Go dependencies
├── config/              # Configuration package
│   └── config.go
├── models/              # Shared data models
│   └── models.go
├── db/                  # BadgerDB operations
│   └── db.go
├── cache/               # Caching package
│   └── cache.go
├── ai/                  # Gemini AI integration
│   └── ai.go
├── service/             # SQL Server service
│   └── sqlserver.go
├── handlers/            # HTTP handlers
│   └── handlers.go
├── sql_files/           # SQL reference files directory
├── data/                # BadgerDB data directory
└── frontend/            # React frontend application
    ├── src/
    ├── public/
    └── package.json
```

