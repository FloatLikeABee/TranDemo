@echo off
echo Starting SQL AI Assistant...
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Go is not installed or not in PATH
    exit /b 1
)

echo Installing Go dependencies...
go mod download

echo.
echo Starting Go server on port 9090...
echo Frontend will be available after building React app
echo.
go run main.go

