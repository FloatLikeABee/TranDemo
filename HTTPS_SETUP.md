# HTTPS Setup for Voice Recognition

## Problem

Modern browsers require HTTPS (or localhost) for microphone access due to security policies. Voice recognition features will not work on HTTP connections (except localhost/127.0.0.1).

## Solutions

### Option 1: Use Localhost (Development)

For development, access the application via:
- `http://localhost:9090`
- `http://127.0.0.1:9090`

These work over HTTP because browsers allow microphone access on localhost.

### Option 2: Enable HTTPS (Production)

For production, you need to set up HTTPS. Here are several options:

#### A. Using a Reverse Proxy (Recommended)

**Nginx Example:**
```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Caddy Example (Auto HTTPS):**
```
your-domain.com {
    reverse_proxy localhost:9090
}
```

#### B. Using Go with TLS

Modify `main.go` to serve HTTPS directly:

```go
// In main.go, replace r.Run() with:
certFile := os.Getenv("SSL_CERT_FILE")
keyFile := os.Getenv("SSL_KEY_FILE")

if certFile != "" && keyFile != "" {
    log.Printf("Server starting on HTTPS port %s", cfg.Port)
    if err := r.RunTLS(":"+cfg.Port, certFile, keyFile); err != nil {
        log.Fatalf("Failed to start HTTPS server: %v", err)
    }
} else {
    log.Printf("Server starting on HTTP port %s", cfg.Port)
    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

Then set environment variables:
```bash
SSL_CERT_FILE=/path/to/cert.pem
SSL_KEY_FILE=/path/to/key.pem
```

#### C. Using Let's Encrypt (Free SSL)

1. Install Certbot:
```bash
# Ubuntu/Debian
sudo apt-get install certbot

# macOS
brew install certbot
```

2. Generate certificate:
```bash
sudo certbot certonly --standalone -d your-domain.com
```

3. Certificates will be in `/etc/letsencrypt/live/your-domain.com/`

4. Use with Nginx or Caddy (see above)

#### D. Self-Signed Certificate (Development/Testing)

Generate self-signed certificate:
```bash
openssl req -x509 -newkey rsa:4096 -nodes -keyout key.pem -out cert.pem -days 365 -subj "/CN=localhost"
```

Then use with Option B above.

**Note:** Browsers will show a security warning for self-signed certificates. Click "Advanced" → "Proceed to localhost" to continue.

### Option 3: Development Server with HTTPS

If using React development server, you can enable HTTPS:

1. Install `mkcert` for local certificates:
```bash
# macOS
brew install mkcert

# Windows (with Chocolatey)
choco install mkcert

# Linux
# Follow instructions at https://github.com/FiloSottile/mkcert
```

2. Create local CA and certificate:
```bash
mkcert -install
mkcert localhost 127.0.0.1
```

3. Update `package.json` in frontend:
```json
{
  "scripts": {
    "start": "HTTPS=true SSL_CRT_FILE=./localhost+1.pem SSL_KEY_FILE=./localhost+1-key.pem react-scripts start"
  }
}
```

## Browser Behavior

- **HTTPS**: ✅ Microphone access allowed
- **localhost/127.0.0.1 over HTTP**: ✅ Microphone access allowed
- **Other domains over HTTP**: ❌ Microphone access blocked

## Error Messages

The application now detects non-secure contexts and shows helpful error messages:

- "Voice recording requires HTTPS or localhost"
- Clear instructions for users
- Guidance for developers

## Quick Test

To quickly test if microphone access works:

1. Open browser console
2. Run: `navigator.mediaDevices.getUserMedia({ audio: true })`
3. If it works, you'll get a MediaStream
4. If blocked, you'll see an error about secure context

## Production Checklist

- [ ] Set up HTTPS with valid SSL certificate
- [ ] Update API_BASE_URL to use HTTPS
- [ ] Test voice registration
- [ ] Test voice recognition
- [ ] Verify microphone permissions work
- [ ] Test on mobile devices (if applicable)

## Troubleshooting

### "Microphone permission denied"
- Check browser settings
- Clear site permissions and re-allow
- Try in incognito/private mode

### "No microphone found"
- Check device connections
- Verify microphone in system settings
- Try different browser

### "Secure context required"
- Use HTTPS or localhost
- Check certificate validity
- Verify browser supports secure contexts

