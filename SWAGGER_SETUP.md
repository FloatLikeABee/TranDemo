# Swagger API Documentation Setup

## Overview

The API now includes Swagger/OpenAPI documentation for all endpoints. You can access the interactive Swagger UI to explore and test the API.

## Accessing Swagger UI

Once the server is running, access Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

## Generating Swagger Documentation

To regenerate Swagger documentation after making changes to handlers:

### Install Swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate Documentation

From the project root directory:

```bash
swag init
```

This will scan your handlers and generate/update the `docs/` directory with Swagger documentation.

## Swagger Annotations

All handlers are annotated with Swagger comments:

- `@Summary` - Brief description of the endpoint
- `@Description` - Detailed description
- `@Tags` - Group endpoints by category
- `@Accept` - Content types accepted (json, multipart/form-data, etc.)
- `@Produce` - Content types produced (json, text/html, etc.)
- `@Param` - Request parameters (path, query, body, header)
- `@Success` - Success response format
- `@Failure` - Error response formats
- `@Router` - HTTP method and path

## Example

```go
// ChatHandler handles chat requests to generate SQL queries
// @Summary      Generate SQL query from natural language
// @Description  Send a message describing what SQL query you need
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        request  body      models.ChatRequest  true  "Chat request"
// @Success      200      {object}  models.ChatResponse
// @Router       /api/chat [post]
func (h *Handlers) ChatHandler(c *gin.Context) {
    // Handler implementation
}
```

## API Endpoints Documented

### Chat
- `POST /api/chat` - Generate SQL from natural language

### SQL Files
- `POST /api/sql/upload` - Upload SQL reference file
- `GET /api/sql/files` - List SQL reference files

### SQL Execution
- `POST /api/sql/execute` - Execute SQL query

### Results
- `GET /api/results/files` - List result files
- `GET /api/results/file/{filename}` - Get result file
- `POST /api/results/generate-html` - Generate HTML page
- `GET /api/results/html/{filename}` - Serve HTML page

### Health
- `GET /health` - Health check

## Testing with Swagger UI

1. Start the server: `go run main.go`
2. Open browser: `http://localhost:8080/swagger/index.html`
3. Click on any endpoint to expand it
4. Click "Try it out" to test the endpoint
5. Fill in parameters and click "Execute"
6. View the response

## Notes

- Swagger UI is available in development and production
- All endpoints are documented with request/response examples
- Error responses are documented for each endpoint
- The documentation is automatically generated from code comments

