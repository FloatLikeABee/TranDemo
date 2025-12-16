# Setup Guide

## Prerequisites

1. **Go** (version 1.21 or higher)
   - Download from: https://golang.org/dl/
   - Verify installation: `go version`

2. **Node.js** (version 18 or higher) and npm
   - Download from: https://nodejs.org/
   - Verify installation: `node --version` and `npm --version`

## Backend Setup

1. **Install Go dependencies:**
   ```bash
   go mod download
   ```

2. **Create necessary directories:**
   - The `data/` directory will be created automatically for BadgerDB
   - The `sql_files/` directory already exists with example files

3. **Run the backend server:**
   ```bash
   # Windows
   go run main.go
   # or
   .\start.bat
   
   # Linux/Mac
   go run main.go
   ```

   The server will start on `http://localhost:8080`

## Frontend Setup

1. **Navigate to frontend directory:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start development server:**
   ```bash
   npm start
   ```
   
   The frontend will start on `http://localhost:3000` (default React port)

4. **Build for production:**
   ```bash
   npm run build
   ```
   
   After building, the backend will serve the frontend from the `build` directory.

## Configuration

### Environment Variables (Optional)

You can set these environment variables to customize the application:

- `PORT`: Server port (default: 8080)
- `GEMINI_API_KEY`: Your Gemini API key (default: already set in code)
- `GEMINI_MODEL`: Model name (default: gemini-1.5-flash-latest)
- `DB_PATH`: BadgerDB storage path (default: ./data/badger)
- `SQL_FILES_DIR`: SQL files directory (default: ./sql_files)

### Adding SQL Reference Files

You can add SQL reference files in two ways:

1. **Place files in `sql_files/` directory:**
   - Add `.sql` files to the `sql_files/` directory
   - The server will automatically load them on startup

2. **Upload via API:**
   ```bash
   curl -X POST http://localhost:8080/api/sql/upload \
     -F "file=@your_file.sql"
   ```

## Running the Application

### Development Mode

1. **Terminal 1 - Backend:**
   ```bash
   go run main.go
   ```

2. **Terminal 2 - Frontend:**
   ```bash
   cd frontend
   npm start
   ```

   Access the app at: `http://localhost:3000`

### Production Mode

1. **Build the frontend:**
   ```bash
   cd frontend
   npm run build
   ```

2. **Run the backend (serves both API and frontend):**
   ```bash
   go run main.go
   ```

   Access the app at: `http://localhost:8080`

## Testing the API

### Health Check
```bash
curl http://localhost:8080/health
```

### Chat Endpoint
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a query to get all users"}'
```

### List SQL Files
```bash
curl http://localhost:8080/api/sql/files
```

### Upload SQL File
```bash
curl -X POST http://localhost:8080/api/sql/upload \
  -F "file=@sql_files/example1.sql"
```

## Troubleshooting

### Go Module Issues
If you encounter module errors:
```bash
go mod tidy
go mod download
```

### Port Already in Use
Change the port:
```bash
# Windows PowerShell
$env:PORT="8081"; go run main.go

# Linux/Mac
PORT=8081 go run main.go
```

### Frontend Build Issues
Clear cache and reinstall:
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Gemini API Errors
- Verify your API key is correct
- Check your internet connection
- Ensure you have API quota available

## Project Structure

```
.
├── main.go              # Main Go application
├── go.mod               # Go dependencies
├── go.sum               # Go dependency checksums
├── sql_files/           # SQL reference files
│   ├── example1.sql
│   └── example2.sql
├── data/                # BadgerDB data (auto-created)
├── frontend/            # React application
│   ├── src/
│   │   ├── App.js       # Main React component
│   │   ├── App.css      # Styles
│   │   └── index.js     # Entry point
│   ├── public/
│   └── package.json
├── start.bat            # Windows startup script
├── start.ps1            # PowerShell startup script
└── README.md            # Project documentation
```

