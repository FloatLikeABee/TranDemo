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
   
   **⚠️ IMPORTANT for macOS users:** Do NOT use `go run main.go` directly - it will fail with a "missing LC_UUID" error. Use one of these methods instead:
   
   ```bash
   # macOS/Linux (Recommended)
   ./start.sh
   
   # Or using Make (if available)
   make run
   
   # Or build and run manually
   go build -o tran_demo main.go
   ./tran_demo
   
   # Windows
   go run main.go
   # or
   .\start.bat
   ```

   The server will start on `http://localhost:9090`

## Frontend Setup

1. **Navigate to frontend directory:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```
   
   **Note:** If you get a "react-scripts: command not found" error, make sure `package.json` has a valid `react-scripts` version (should be `5.0.1` or similar, not `^0.0.0`). Then run `npm install` again.

3. **Verify required files exist:**
   The `public/index.html` file should exist. If you get a "Could not find a required file: index.html" error, make sure the `public` directory exists with `index.html` inside it.

4. **Start development server:**
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

- `PORT`: Server port (default: 9090)
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
   curl -X POST http://localhost:9090/api/sql/upload \
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

   Access the app at: `http://localhost:9090`

## Testing the API

### Health Check
```bash
curl http://localhost:9090/health
```

### Chat Endpoint
```bash
curl -X POST http://localhost:9090/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a query to get all users"}'
```

### List SQL Files
```bash
curl http://localhost:9090/api/sql/files
```

### Upload SQL File
```bash
curl -X POST http://localhost:9090/api/sql/upload \
  -F "file=@sql_files/example1.sql"
```

## Troubleshooting

### Go Module Issues
If you encounter module errors:
```bash
go mod tidy
go mod download
```

### macOS "missing LC_UUID load command" Error
If you encounter the `dyld: missing LC_UUID load command` error when running `go run main.go` on macOS:

**This is a known issue with `go run` on macOS when using CGO dependencies (like BadgerDB).**

**✅ Solution 1 (Recommended):** Use the startup script:
```bash
./start.sh
```

**✅ Solution 2:** Use Make (if installed):
```bash
make run
```

**✅ Solution 3:** Build and run manually (with external linking to fix LC_UUID):
```bash
go build -ldflags="-linkmode=external -w -s" -o tran_demo main.go
./tran_demo
```

**❌ DO NOT USE:** `go run main.go` - This will always fail on macOS with this error.

**Why this happens:** The `go run` command creates temporary binaries in `/tmp/go-build*` that don't have proper macOS load commands when CGO is involved. Using `go build` with external linking (`-linkmode=external`) ensures the binary includes the required `LC_UUID` load command that macOS requires.

### Port Already in Use
Change the port:
```bash
# Windows PowerShell
$env:PORT="8081"; go run main.go

# Linux/Mac
PORT=8081 go run main.go
```

### Frontend Build Issues

**"react-scripts: command not found" Error:**
If you get this error, it usually means:
1. Dependencies aren't installed - run `npm install` in the `frontend` directory
2. `package.json` has an invalid `react-scripts` version - check that it's `5.0.1` or similar (not `^0.0.0`)

Fix by reinstalling:
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**Other build issues:**
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

