package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// RegisterVoiceHandler registers a voice sample for a user
// @Summary      Register voice profile
// @Description  Register a user's voice sample for speaker recognition
// @Tags         Voice Recognition
// @Accept       json
// @Produce      json
// @Param        request  body      models.VoiceRegistrationRequest  true  "Voice registration request"
// @Success      200      {object}  models.VoiceProfile  "Voice profile created"
// @Failure      400      {object}  map[string]string     "Invalid request"
// @Failure      500      {object}  map[string]string     "Failed to register voice"
// @Router       /api/voice/register [post]
func (h *Handlers) RegisterVoiceHandler(c *gin.Context) {
	var req models.VoiceRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from header or generate one
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		// Generate user ID from name hash
		hash := md5.Sum([]byte(strings.ToLower(req.Name)))
		userID = hex.EncodeToString(hash[:])
	}

	// Check if profile already exists
	existingProfile, err := h.db.GetVoiceProfile(userID)
	if err == nil && existingProfile != nil {
		// Add new voice sample to existing profile
		if err := h.voiceService.AddVoiceSample(existingProfile, req.AudioData, req.AudioFormat); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add voice sample: " + err.Error()})
			return
		}
		
		// Update profile in database
		if err := h.db.StoreVoiceProfile(existingProfile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update voice profile: " + err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, existingProfile)
		return
	}

	// Create new voice profile
	profile, err := h.voiceService.RegisterVoice(userID, req.Name, req.AudioData, req.AudioFormat)
	if err != nil {
		log.Printf("[VOICE] Error registering voice: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register voice: " + err.Error()})
		return
	}

	// Store profile in database
	if err := h.db.StoreVoiceProfile(profile); err != nil {
		log.Printf("[VOICE] Error storing voice profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store voice profile: " + err.Error()})
		return
	}

	log.Printf("[VOICE] Registered voice for user: %s (%s)", profile.Name, profile.UserID)
	c.JSON(http.StatusOK, profile)
}

// RecognizeVoiceHandler recognizes a speaker from voice input
// @Summary      Recognize voice
// @Description  Recognize a speaker and detect attendance intent from voice input
// @Tags         Voice Recognition
// @Accept       json
// @Produce      json
// @Param        request  body      models.VoiceRecognitionRequest  true  "Voice recognition request"
// @Success      200      {object}  models.VoiceRecognitionResponse  "Recognition result"
// @Failure      400      {object}  map[string]string                "Invalid request"
// @Failure      500      {object}  map[string]string                "Failed to recognize voice"
// @Router       /api/voice/recognize [post]
func (h *Handlers) RecognizeVoiceHandler(c *gin.Context) {
	var req models.VoiceRecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get all voice profiles
	profiles, err := h.db.GetAllVoiceProfiles()
	if err != nil {
		log.Printf("[VOICE] Error getting voice profiles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load voice profiles: " + err.Error()})
		return
	}

	if len(profiles) == 0 {
		c.JSON(http.StatusOK, models.VoiceRecognitionResponse{
			Recognized: false,
			Message:    "Sorry, you're not in our school.",
		})
		return
	}

	// Recognize voice
	response, err := h.voiceService.RecognizeVoice(req.AudioData, profiles)
	if err != nil {
		log.Printf("[VOICE] Error recognizing voice: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recognize voice: " + err.Error()})
		return
	}

	// If recognized and intent is attendance-related, log it
	if response.Recognized && (response.Intent == "attendance" || response.Intent == "punch_in" || response.Intent == "here") {
		log.Printf("[VOICE] Attendance logged for: %s (%s)", response.Name, response.UserID)
		// Store attendance in chat history
		attendanceMsg := fmt.Sprintf("%s - %s", response.Name, time.Now().Format("2006-01-02 15:04:05"))
		h.db.StoreChatHistory(response.UserID, "Voice attendance", attendanceMsg)
	}

	c.JSON(http.StatusOK, response)
}

// ListVoiceProfilesHandler lists all registered voice profiles
// @Summary      List voice profiles
// @Description  Get a list of all registered voice profiles
// @Tags         Voice Recognition
// @Produce      json
// @Success      200  {object}  map[string][]models.VoiceProfile  "List of voice profiles"
// @Failure      500  {object}  map[string]string                  "Failed to list profiles"
// @Router       /api/voice/profiles [get]
func (h *Handlers) ListVoiceProfilesHandler(c *gin.Context) {
	profiles, err := h.db.GetAllVoiceProfiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list voice profiles: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

// DeleteVoiceProfileHandler deletes a voice profile
// @Summary      Delete voice profile
// @Description  Delete a registered voice profile
// @Tags         Voice Recognition
// @Param        user_id  path      string  true  "User ID"
// @Success      200      {object}  map[string]string  "Profile deleted"
// @Failure      404      {object}  map[string]string  "Profile not found"
// @Failure      500      {object}  map[string]string  "Failed to delete profile"
// @Router       /api/voice/profile/{user_id} [delete]
func (h *Handlers) DeleteVoiceProfileHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if err := h.db.DeleteVoiceProfile(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete voice profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Voice profile deleted successfully"})
}

// HandleVoiceChat processes voice input through the chat interface
func (h *Handlers) HandleVoiceChat(c *gin.Context, userID string, audioData string) (*models.ChatResponse, error) {
	// Get all voice profiles
	profiles, err := h.db.GetAllVoiceProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load voice profiles: %w", err)
	}

	// Recognize voice
	voiceResponse, err := h.voiceService.RecognizeVoice(audioData, profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to recognize voice: %w", err)
	}

	// Generate chat response based on recognition result
	var chatResponse models.ChatResponse
	
	if !voiceResponse.Recognized {
		chatResponse.Response = voiceResponse.Message // "Sorry, you're not in our school."
		return &chatResponse, nil
	}

	// User recognized - check intent
	if voiceResponse.Intent == "attendance" || voiceResponse.Intent == "punch_in" || voiceResponse.Intent == "here" {
		chatResponse.Response = voiceResponse.Message // "Punched in" or "Gotcha!"
		
		// Log attendance
		attendanceMsg := fmt.Sprintf("%s - %s", voiceResponse.Name, time.Now().Format("2006-01-02 15:04:05"))
		h.db.StoreChatHistory(voiceResponse.UserID, "Voice attendance", attendanceMsg)
	} else {
		chatResponse.Response = voiceResponse.Message
	}

	return &chatResponse, nil
}

