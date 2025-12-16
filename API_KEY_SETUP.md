# API Key Setup for DashScope

## Issue: Invalid API Key (401 Error)

If you're getting a `401 InvalidApiKey` error, it means the API key being used is not valid for DashScope API.

## DashScope API Key Format

DashScope API keys from Alibaba Cloud typically:
- Are longer strings (usually 32+ characters)
- Do NOT start with `sk-` (that's OpenRouter format)
- Look like: `sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` (but this is NOT DashScope format)

## How to Get a Valid DashScope API Key

1. Go to [Alibaba Cloud DashScope Console](https://dashscope.console.aliyun.com/)
2. Sign in or create an account
3. Navigate to API Keys section
4. Create a new API key
5. Copy the API key (it should be a long string, not starting with `sk-`)

## Setting the API Key

### Option 1: Environment Variable (Recommended)

**Windows PowerShell:**
```powershell
$env:GEMINI_API_KEY="your-actual-dashscope-api-key-here"
```

**Windows CMD:**
```cmd
set GEMINI_API_KEY=your-actual-dashscope-api-key-here
```

**Linux/Mac:**
```bash
export GEMINI_API_KEY="your-actual-dashscope-api-key-here"
```

Then restart your server.

### Option 2: Update Config File

Edit `config/config.go` and change the default value:
```go
GeminiAPIKey: getEnv("GEMINI_API_KEY", "your-actual-dashscope-api-key-here"),
```

## Verify the API Key

When you run the server, check the debug output. You should see:
```
API Key (masked): xxxx...xxxx (length: XX)
```

If the length is very short (< 20 characters), the key is likely incorrect.

## Model Name

The model name should match what DashScope supports. Common models:
- `qwen-coder` (as shown in curl example)
- `qwen2.5-coder` 
- `qwen3-max` (current default)

Set via environment variable:
```powershell
$env:GEMINI_MODEL="qwen-coder"
```

## Testing

After setting the correct API key, test with:
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Create a SQL query to get all users"}'
```

You should see the debug output in the server console showing the request details.

