# PowerShell script to start the application
Write-Host "Starting SQL AI Assistant..." -ForegroundColor Green

# Check if Go is installed
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if (-not $goInstalled) {
    Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

# Install Go dependencies
Write-Host "Installing Go dependencies..." -ForegroundColor Yellow
go mod download

# Start the server
Write-Host "Starting Go server on port 9090..." -ForegroundColor Yellow
Write-Host "Frontend will be available after building React app" -ForegroundColor Yellow
Write-Host ""
go run main.go

