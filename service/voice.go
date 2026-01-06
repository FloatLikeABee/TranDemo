package service

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"idongivaflyinfa/models"
)

type VoiceService struct {
	voiceSamplesDir string
}

func NewVoiceService(voiceSamplesDir string) *VoiceService {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(voiceSamplesDir, 0755); err != nil {
		log.Printf("Warning: Failed to create voice samples directory: %v", err)
	}
	
	return &VoiceService{
		voiceSamplesDir: voiceSamplesDir,
	}
}

// RegisterVoice registers a voice sample for a user
func (v *VoiceService) RegisterVoice(userID, name, audioData, audioFormat string) (*models.VoiceProfile, error) {
	// Decode base64 audio data
	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio data: %w", err)
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.%s", userID, name, timestamp, audioFormat)
	filePath := filepath.Join(v.voiceSamplesDir, filename)
	
	// Save audio file
	if err := os.WriteFile(filePath, audioBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save audio file: %w", err)
	}
	
	log.Printf("[VOICE] Saved voice sample to: %s", filePath)
	
	// Create or update voice profile
	profile := &models.VoiceProfile{
		UserID:      userID,
		Name:        name,
		VoiceSamples: []string{filename}, // Store filename reference
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	
	return profile, nil
}

// AddVoiceSample adds an additional voice sample to an existing profile
func (v *VoiceService) AddVoiceSample(profile *models.VoiceProfile, audioData, audioFormat string) error {
	// Decode base64 audio data
	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return fmt.Errorf("failed to decode audio data: %w", err)
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.%s", profile.UserID, profile.Name, timestamp, audioFormat)
	filePath := filepath.Join(v.voiceSamplesDir, filename)
	
	// Save audio file
	if err := os.WriteFile(filePath, audioBytes, 0644); err != nil {
		return fmt.Errorf("failed to save audio file: %w", err)
	}
	
	// Add to profile
	profile.VoiceSamples = append(profile.VoiceSamples, filename)
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	
	log.Printf("[VOICE] Added voice sample to profile: %s", filename)
	return nil
}

// RecognizeVoice attempts to recognize a speaker from audio input
// This is a simplified implementation - in production, you'd use a proper speaker verification service
func (v *VoiceService) RecognizeVoice(audioData string, profiles []models.VoiceProfile) (*models.VoiceRecognitionResponse, error) {
	// Decode audio
	audioBytes, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio data: %w", err)
	}
	
	// Generate hash of the audio for simple matching
	// In production, use proper speaker verification (e.g., voice biometrics)
	audioHash := md5.Sum(audioBytes)
	audioHashStr := hex.EncodeToString(audioHash[:])
	
	log.Printf("[VOICE] Recognizing voice, audio hash: %s", audioHashStr)
	
	// Simple matching: compare audio hash with stored samples
	// NOTE: This is a simplified approach. Real speaker verification requires:
	// - Feature extraction (MFCC, spectrograms)
	// - Machine learning models (neural networks)
	// - Voice biometric services (e.g., Azure Speaker Recognition, AWS Voice ID)
	
	// For now, we'll do a basic comparison
	// In production, replace this with actual speaker verification
	var matchedProfile *models.VoiceProfile
	for i := range profiles {
		// Load and compare voice samples
		for _, sampleFile := range profiles[i].VoiceSamples {
			samplePath := filepath.Join(v.voiceSamplesDir, sampleFile)
			sampleBytes, err := os.ReadFile(samplePath)
			if err != nil {
				log.Printf("[VOICE] Warning: Failed to read sample %s: %v", sampleFile, err)
				continue
			}
			
			sampleHash := md5.Sum(sampleBytes)
			sampleHashStr := hex.EncodeToString(sampleHash[:])
			
			// Simple exact match (in production, use similarity threshold)
			if audioHashStr == sampleHashStr {
				matchedProfile = &profiles[i]
				break
			}
		}
		if matchedProfile != nil {
			break
		}
	}
	
	if matchedProfile == nil {
		return &models.VoiceRecognitionResponse{
			Recognized: false,
			Message:    "Sorry, you're not in our school.",
		}, nil
	}
	
	// Extract transcript and intent from audio
	// In production, use speech-to-text service (e.g., Google Speech-to-Text, Azure Speech)
	transcript, intent := v.extractIntent(audioBytes)
	
	response := &models.VoiceRecognitionResponse{
		Recognized: true,
		UserID:     matchedProfile.UserID,
		Name:       matchedProfile.Name,
		Transcript: transcript,
		Intent:     intent,
	}
	
	// Generate appropriate response message
	if intent == "attendance" || intent == "punch_in" {
		response.Message = "Punched in"
	} else if intent == "here" {
		response.Message = "Gotcha!"
	} else {
		response.Message = fmt.Sprintf("Hello %s!", matchedProfile.Name)
	}
	
	return response, nil
}

// extractIntent extracts intent from audio (simplified - in production use speech-to-text)
func (v *VoiceService) extractIntent(audioBytes []byte) (string, string) {
	// This is a placeholder - in production, you would:
	// 1. Convert audio to text using speech-to-text service
	// 2. Analyze text for keywords
	// 3. Return transcript and intent
	
	// For now, return placeholder values
	// The actual implementation would call a speech-to-text API
	transcript := "[Speech-to-text would be implemented here]"
	intent := "attendance" // Default intent
	
	// In production, implement actual speech-to-text:
	// - Google Cloud Speech-to-Text
	// - Azure Speech Services
	// - AWS Transcribe
	// - Or use a library like vosk (offline)
	
	return transcript, intent
}

// DetectAttendanceIntent detects if the transcript contains attendance-related phrases
func (v *VoiceService) DetectAttendanceIntent(transcript string) string {
	lowerTranscript := strings.ToLower(transcript)
	
	attendancePhrases := map[string]string{
		"i'm here":        "here",
		"im here":         "here",
		"here":            "here",
		"punch in":        "punch_in",
		"punchin":         "punch_in",
		"attendance":      "attendance",
		"register":        "attendance",
		"check in":        "punch_in",
		"checkin":         "punch_in",
		"present":         "attendance",
		"mark attendance": "attendance",
	}
	
	for phrase, intent := range attendancePhrases {
		if strings.Contains(lowerTranscript, phrase) {
			return intent
		}
	}
	
	return "unknown"
}

