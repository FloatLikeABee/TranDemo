# Transfinder Form/Report Assistant

A full-stack application that uses AI to generate SQL queries, manage forms, handle document uploads (images/PDFs), and support voice input. It includes a chat-based interface, form templates, student/staff registration flows, and optional integration with SQL Server and external document-processing services.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [Optional External Services](#optional-external-services)
- [API Overview](#api-overview)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)

---

## Features

- **AI-powered SQL generation** – Natural language to SQL using configured AI (e.g. Qwen/DashScope)
- **Chat interface** – Text and voice input, file uploads (images/PDFs)
- **Forms system** – Create form templates, collect answers, student/staff registration via chat
- **Document processing** – Upload images or PDFs; optional image-reader/PDF-reader services for extraction
- **Voice** – Voice registration and recognition (requires HTTPS or localhost)
- **SQL Server** – Optional execution of generated SQL against Microsoft SQL Server
- **Results & reports** – Store query results, generate HTML pages
- **Swagger** – Interactive API docs at `/swagger/index.html`

---

## Architecture

| Layer        | Technology / Role |
|-------------|--------------------|
| **Backend** | Go 1.21+, Gin, BadgerDB (embedded), go-cache |
| **Frontend**| React 18, axios, react-speech-recognition |
| **AI**      | Configurable (e.g. DashScope/Qwen); see `config` and `API_KEY_SETUP.md` |
| **Optional**| SQL Server, external API (image-reader, pdf-reader, gathering) |

**Key packages:** `config/`, `db/`, `cache/`, `ai/`, `handlers/`, `models/`, `service/`, `validation/`

---

## Prerequisites

- **Go 1.21 or higher** – [Download](https://golang.org/dl/)  
  Verify: `go version`
- **Node.js 18+ and npm** – [Download](https://nodejs.org/)  
  Verify: `node --version`, `npm --version`
- **Optional:** Microsoft SQL Server (for `/api/sql/execute` and result storage)
- **Optional:** External API for image/PDF reading and gathering (see [Optional External Services](#optional-external-services))

---

## Installation

### 1. Clone and enter the project

```bash
git clone <repository-url>
cd to the root
```

### 2. Backend setup

```bash
# Install Go dependencies
go mod download

# Optional: create dirs (BadgerDB and sql_files are created/used automatically if missing)
mkdir -p data sql_files results sites voice_samples
```

**macOS note:** Do **not** use `go run main.go` if you see a **missing LC_UUID** error (common with CGO/BadgerDB). Use one of:

```bash
# Build then run
go build -o tran_demo main.go
./tran_demo

# Or use the provided script (Windows: start.bat or start.ps1)
```

### 3. Frontend setup

```bash
cd frontend
npm install
cd ..
```

If you see `react-scripts: command not found`, ensure `package.json` has `"react-scripts": "^5.0.1"` (or similar) and run `npm install` again.

### 4. Environment (optional)

Copy or set environment variables as needed (see [Configuration](#configuration)). Minimum to run:

- Backend and frontend work with defaults (port 9090, embedded DB).
- For AI: set `GEMINI_API_KEY` (or configure in `config/config.go`) and optionally `GEMINI_MODEL`.
- For production frontend URL: set `REACT_APP_API_URL` before `npm run build`.

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9090` | Backend HTTP port |
| `GEMINI_API_KEY` | (in code) | AI API key (DashScope/Qwen or other; see `API_KEY_SETUP.md`) |
| `GEMINI_MODEL` | (in code) | AI model name |
| `DB_PATH` | `./data/badger` | BadgerDB data directory |
| `SQL_FILES_DIR` | `./sql_files` | Directory for reference SQL files |
| `RESULTS_DIR` | `./results` | Directory for query result files |
| `SITES_DIR` | `./sites` | Directory for generated HTML pages |
| `VOICE_SAMPLES_DIR` | `./voice_samples` | Voice registration samples |
| `EXTERNAL_API_BASE` | `http://localhost:8000` | Base URL for image-reader, pdf-reader, gathering |
| `SQL_SERVER` | (in code) | SQL Server host |
| `SQL_PORT` | `1433` | SQL Server port |
| `SQL_DATABASE` | (in code) | Database name |
| `SQL_USER` / `SQL_PASSWORD` | (in code) | SQL Server credentials |
| `SQL_ENCRYPT` | `true` | Use encrypted connection to SQL Server |
| `REACT_APP_API_URL` | `http://localhost:9090` | Backend URL used by React (set before `npm run build`) |

---

## Running the Application

### Development (backend + frontend separately)

**Terminal 1 – Backend:**

```bash
# Windows
go run main.go
# Or: .\start.bat  or  .\start.ps1

# macOS/Linux (if go run fails with LC_UUID, use build + run)
go build -o tran_demo main.go && ./tran_demo
```

**Terminal 2 – Frontend:**

```bash
cd frontend
npm start
```

- Backend: `http://localhost:9090`
- Frontend (dev): `http://localhost:3000` (proxies API to 9090 via `package.json` proxy)

### Production (single server)

```bash
cd frontend
npm run build
cd ..
go run main.go
# Or run the built binary: ./tran_demo
```

- App: `http://localhost:9090` (backend serves React from `frontend/build`)

---

## Optional External Services

The app can call an external API (default base: `http://localhost:8000`) for:

- **Image reader** – `POST .../image-reader/read-and-process` (e.g. for chat file uploads)
- **PDF reader** – `POST .../pdf-reader/read`
- **Gathering** – `POST .../gathering/gather` (e.g. research flow)

If these are not running, chat file upload and related features will return an error message; the rest of the app still works. Set `EXTERNAL_API_BASE` to your service URL.

---

## API Overview

- **Health:** `GET /health`
- **Chat:** `POST /api/chat` (JSON body or `multipart/form-data` with `message` and optional `file`)
- **SQL:** `POST /api/sql/upload`, `GET /api/sql/files`, `POST /api/sql/execute`
- **Results:** `GET /api/results/files`, `GET /api/results/file/:filename`, `POST /api/results/generate-html`, `GET /api/results/html/:filename`
- **Voice:** `POST /api/voice/register`, `POST /api/voice/recognize`, `GET /api/voice/profiles`, `DELETE /api/voice/profile/:user_id`
- **Forms:** `GET/POST/PUT/DELETE /api/forms/templates`, `GET/POST/PUT/DELETE /api/forms/answers`
- **Swagger:** `http://localhost:9090/swagger/index.html`

---

## Troubleshooting

### Backend won’t start

**Symptom:** `Failed to initialize database: ...`  
- **Cause:** BadgerDB path not writable or disk full.  
- **Fix:** Ensure `DB_PATH` (default `./data/badger`) is writable; free disk space; fix permissions.

**Symptom:** `Failed to initialize Gemini: ...` or 401 from AI  
- **Cause:** Invalid or missing API key / wrong model.  
- **Fix:** Set `GEMINI_API_KEY` (and optionally `GEMINI_MODEL`). See `API_KEY_SETUP.md`. Restart backend.

**Symptom:** `Port 9090 already in use` or `bind: address already in use`  
- **Fix:** Stop the process using the port or use another port:
  - **Windows (PowerShell):** `$env:PORT="8081"; go run main.go`
  - **Windows (CMD):** `set PORT=8081 && go run main.go`
  - **Linux/macOS:** `PORT=8081 go run main.go`  
  If you change the port, point the frontend at the new URL (e.g. `REACT_APP_API_URL=http://localhost:8081` for production build).

**Symptom (macOS):** `dyld: missing LC_UUID load command` or similar when running `go run main.go`  
- **Cause:** CGO/BadgerDB and how `go run` builds the binary.  
- **Fix:** Build and run the binary instead:
  ```bash
  go build -ldflags="-linkmode=external" -o tran_demo main.go
  ./tran_demo
  ```
  Or use `./start.sh` / `make run` if available.

---

### Frontend issues

**Symptom:** `react-scripts: command not found`  
- **Fix:** From project root: `cd frontend`, then `npm install`. Ensure `react-scripts` in `package.json` is valid (e.g. `^5.0.1`), not `^0.0.0`.

**Symptom:** `Could not find a required file: index.html`  
- **Fix:** Ensure `frontend/public/index.html` exists.

**Symptom:** Blank page or API calls to wrong host  
- **Cause:** Production build used default `http://localhost:9090` but backend runs elsewhere.  
- **Fix:** Set `REACT_APP_API_URL` to your backend URL **before** building:
  ```bash
  # Windows (PowerShell)
  $env:REACT_APP_API_URL="http://your-server:9090"; npm run build

  # Linux/macOS
  REACT_APP_API_URL=http://your-server:9090 npm run build
  ```
  Then redeploy the `frontend/build` output.

**Symptom:** CORS errors in browser  
- **Cause:** Backend not allowing frontend origin.  
- **Fix:** This app sets CORS to allow all origins. If you use a reverse proxy, ensure it doesn’t strip or override CORS headers and that the backend is reachable.

---

### Chat / file upload

**Symptom:** `Invalid file type: application/octet-stream. Expected image file.` (from image-reader)  
- **Cause:** Image was sent with generic MIME type.  
- **Fix:** Backend now detects image type from content and sends correct `Content-Type`. Update to the latest code. If you still see this, ensure the upload is a supported image (e.g. JPEG, PNG, GIF, WebP).

**Symptom:** `Could not process the uploaded file: ... Make sure the Image Reader / PDF Reader service is running`  
- **Cause:** External image-reader or pdf-reader not running or not reachable.  
- **Fix:** Start the external service and set `EXTERNAL_API_BASE` (e.g. `http://localhost:8000`). Or avoid using file upload until the service is available.

---

### Voice

**Symptom:** Microphone permission denied or voice not working  
- **Cause:** Browsers require HTTPS (or localhost) for microphone access.  
- **Fix:** Use `http://localhost:9090` or `http://127.0.0.1:9090`, or enable HTTPS. See `HTTPS_SETUP.md`.

**Symptom:** Voice works on localhost but not on another machine  
- **Cause:** Non-localhost HTTP is not allowed for microphone.  
- **Fix:** Serve the app over HTTPS (e.g. reverse proxy with SSL). See `HTTPS_SETUP.md`.

---

### SQL Server

**Symptom:** `Warning: Failed to initialize SQL Server service` or SQL execution fails  
- **Cause:** Wrong host/port/database/credentials or network/encryption issues.  
- **Fix:** Set `SQL_SERVER`, `SQL_PORT`, `SQL_DATABASE`, `SQL_USER`, `SQL_PASSWORD`, `SQL_ENCRYPT`. Ensure SQL Server allows remote connections and that firewall allows the port (default 1433). For Azure or TLS, `SQL_ENCRYPT=true` is typical.

---

### Forms / registration

**Symptom:** “No registration forms set up” or form list empty  
- **Cause:** No form templates in DB.  
- **Fix:** Create a form template via the Forms UI (e.g. `/forms` or React Forms) or via API `POST /api/forms/templates`. Then retry registration in chat.

**Symptom:** Confirmation card in chat shows no details (only title/buttons)  
- **Cause:** Mismatch between field names and answer keys.  
- **Fix:** Frontend now does resilient lookup and fallback; update to latest. If it persists, check that the form template’s field names match what the AI returns in the registration flow.

---

### Go modules

**Symptom:** `go: module ... not found` or inconsistent dependencies  
- **Fix:**
  ```bash
  go mod tidy
  go mod download
  go build ./...
  ```

---

### Data / persistence

**Symptom:** Data lost after restart  
- **Cause:** BadgerDB stores under `DB_PATH` (default `./data/badger`). If you run from a different directory or delete `data/`, DB is recreated.  
- **Fix:** Run the backend from the same working directory and do not delete `data/` (or backup/restore it).

---

## Project Structure

```
.
├── main.go                 # Entry point, Gin router, CORS, routes
├── go.mod, go.sum
├── config/                 # Config and env (port, API key, paths, SQL Server, external API)
├── db/                     # BadgerDB (SQL files, form templates/answers, registration state)
├── cache/                  # In-memory cache
├── ai/                     # AI client and prompts (e.g. SQL, form selection, field gathering)
├── handlers/               # HTTP handlers (chat, SQL, forms, voice, file upload, etc.)
├── models/                 # Request/response and domain models
├── service/                # SQL Server, results, voice service
├── validation/
├── docs/                   # Swagger generated docs
├── sql_files/              # Reference SQL files (optional)
├── data/                   # BadgerDB data (auto-created)
├── results/, sites/, voice_samples/
├── frontend/               # React app (chat, forms, voice UI)
│   ├── src/
│   ├── public/
│   └── package.json
├── presentation/          # Static HTML (forms, form-answers, etc.)
├── start.bat, start.ps1    # Windows startup scripts
├── README.md               # This file
├── SETUP.md                # Short setup guide
├── API_KEY_SETUP.md        # AI API key and model
├── HTTPS_SETUP.md          # HTTPS and voice
└── VOICE_RECOGNITION.md    # Voice feature details
```

---

## Quick test

```bash
# Health
curl http://localhost:9090/health

# Chat (no file)
curl -X POST http://localhost:9090/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "List all users"}'
```

For more examples and full API description, open **Swagger UI** at `http://localhost:9090/swagger/index.html` after starting the backend.
