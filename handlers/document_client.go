package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

// readImageAndProcessWithProvider sends one image to image-reader/read-and-process with the given provider/model.
func (h *Handlers) readImageAndProcessWithProvider(fileContent []byte, filename string, systemPrompt string, provider, model string) (extractedText, aiResult string, err error) {
	base := strings.TrimSuffix(h.externalAPIBase, "/")
	url := base + "/image-reader/read-and-process"

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	_ = w.WriteField("system_prompt", systemPrompt)
	_ = w.WriteField("provider", provider)
	_ = w.WriteField("model", model)
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
