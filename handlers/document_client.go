package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	defaultSummarizePrompt = "Summarize the following content clearly and concisely."
	// Image/PDF reader: try Qwen first, then Mistral as fallback.
	imageReaderProviderPrimary   = "qwen"
	imageReaderModelPrimary      = "qwen-vl-plus"
	imageReaderProviderFallback  = "mistral"
	imageReaderModelFallback     = "mistral-small-latest"
)

// imageReaderProviderModel pairs provider and model for read-and-process.
var imageReaderProviderModels = []struct{ provider, model string }{
	{imageReaderProviderPrimary, imageReaderModelPrimary},
	{imageReaderProviderFallback, imageReaderModelFallback},
}

// detectImageContentType returns an image/* MIME type for the image-reader. Uses content
// detection first so uploads sent as application/octet-stream are sent with the correct type.
func detectImageContentType(fileContent []byte, filename string) string {
	detected := http.DetectContentType(fileContent)
	if strings.HasPrefix(detected, "image/") {
		return strings.TrimSpace(strings.Split(detected, ";")[0])
	}
	// Magic bytes for common image formats (when DetectContentType returns application/octet-stream)
	if len(fileContent) >= 8 {
		switch {
		case len(fileContent) >= 3 && fileContent[0] == 0xFF && fileContent[1] == 0xD8 && fileContent[2] == 0xFF:
			return "image/jpeg"
		case len(fileContent) >= 8 && fileContent[0] == 0x89 && fileContent[1] == 0x50 && fileContent[2] == 0x4E && fileContent[3] == 0x47:
			return "image/png"
		case len(fileContent) >= 6 && fileContent[0] == 0x47 && fileContent[1] == 0x49 && fileContent[2] == 0x46:
			return "image/gif"
		case len(fileContent) >= 12 && fileContent[0] == 0x52 && fileContent[1] == 0x49 && fileContent[2] == 0x46 && fileContent[3] == 0x46 &&
			fileContent[8] == 0x57 && fileContent[9] == 0x45 && fileContent[10] == 0x42 && fileContent[11] == 0x50:
			return "image/webp"
		}
	}
	// Fall back to extension
	if ext := path.Ext(filename); ext != "" {
		if t := mime.TypeByExtension(ext); t != "" && strings.HasPrefix(t, "image/") {
			return strings.TrimSpace(strings.Split(t, ";")[0])
		}
	}
	return "image/jpeg"
}

// readImageAndProcessWithProvider sends one image to image-reader/read-and-process with the given provider/model.
func (h *Handlers) readImageAndProcessWithProvider(fileContent []byte, filename string, systemPrompt string, provider, model string) (extractedText, aiResult string, err error) {
	base := strings.TrimSuffix(h.externalAPIBase, "/")
	url := base + "/image-reader/read-and-process"

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	_ = w.WriteField("system_prompt", systemPrompt)
	_ = w.WriteField("provider", provider)
	_ = w.WriteField("model", model)
	contentType := detectImageContentType(fileContent, filename)
	part, err := w.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="file"; filename="` + filename + `"`},
		"Content-Type":       {contentType},
	})
	if err != nil {
		return "", "", err
	}
	if _, err := part.Write(fileContent); err != nil {
		return "", "", err
	}
	if err := w.Close(); err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("image-reader returned %d: %s", resp.StatusCode, string(data))
	}

	var out struct {
		Success       bool   `json:"success"`
		ExtractedText string `json:"extracted_text"`
		AIResult      string `json:"ai_result"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", "", err
	}
	if !out.Success {
		return "", "", fmt.Errorf("image-reader success=false")
	}
	return out.ExtractedText, out.AIResult, nil
}

// ReadImageAndProcess sends one image to image-reader/read-and-process. Tries Qwen first, then Mistral on failure.
func (h *Handlers) ReadImageAndProcess(file io.Reader, filename string, systemPrompt string) (extractedText, aiResult string, err error) {
	if systemPrompt == "" {
		systemPrompt = defaultSummarizePrompt
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	var lastErr error
	for _, pm := range imageReaderProviderModels {
		extractedText, aiResult, err := h.readImageAndProcessWithProvider(fileContent, filename, systemPrompt, pm.provider, pm.model)
		if err == nil {
			return extractedText, aiResult, nil
		}
		lastErr = err
	}
	return "", "", lastErr
}

// readPDFAndProcessWithProvider sends a PDF to pdf-reader/read with the given llm_provider and model_name.
func (h *Handlers) readPDFAndProcessWithProvider(fileContent []byte, filename string, systemPrompt string, provider, model string) (extractedText, aiResult string, err error) {
	base := strings.TrimSuffix(h.externalAPIBase, "/")
	url := base + "/pdf-reader/read"

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	_ = w.WriteField("system_prompt", systemPrompt)
	_ = w.WriteField("llm_provider", provider)
	_ = w.WriteField("model_name", model)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", "", err
	}
	if _, err := part.Write(fileContent); err != nil {
		return "", "", err
	}
	if err := w.Close(); err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("pdf-reader returned %d: %s", resp.StatusCode, string(data))
	}

	var out struct {
		Success       bool   `json:"success"`
		ExtractedText string `json:"extracted_text"`
		AIResult      string `json:"ai_result"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", "", err
	}
	if !out.Success {
		return "", "", fmt.Errorf("pdf-reader success=false")
	}
	return out.ExtractedText, out.AIResult, nil
}

// ReadPDFAndProcess sends a PDF to pdf-reader/read. Tries Qwen first, then Mistral on failure.
func (h *Handlers) ReadPDFAndProcess(file io.Reader, filename string, systemPrompt string) (extractedText, aiResult string, err error) {
	if systemPrompt == "" {
		systemPrompt = defaultSummarizePrompt
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	var lastErr error
	for _, pm := range imageReaderProviderModels {
		extractedText, aiResult, err := h.readPDFAndProcessWithProvider(fileContent, filename, systemPrompt, pm.provider, pm.model)
		if err == nil {
			return extractedText, aiResult, nil
		}
		lastErr = err
	}
	return "", "", lastErr
}

// Gather calls the gathering API for web research and returns the markdown content.
func (h *Handlers) Gather(prompt string, maxIterations int) (content string, err error) {
	if maxIterations <= 0 {
		maxIterations = 10
	}
	if maxIterations > 20 {
		maxIterations = 20
	}
	base := strings.TrimSuffix(h.externalAPIBase, "/")
	url := base + "/gathering/gather"

	body := struct {
		Prompt         string  `json:"prompt"`
		MaxIterations  int     `json:"max_iterations"`
		LLMProvider    string  `json:"llm_provider,omitempty"`
		ModelName      string  `json:"model_name,omitempty"`
		MaxTokens      int     `json:"max_tokens,omitempty"`
		Temperature    float64 `json:"temperature,omitempty"`
	}{Prompt: prompt, MaxIterations: maxIterations}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gathering returned %d: %s", resp.StatusCode, string(respData))
	}

	var out struct {
		Success bool   `json:"success"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(respData, &out); err != nil {
		return "", err
	}
	if !out.Success {
		return "", fmt.Errorf("gathering success=false")
	}
	return out.Content, nil
}
