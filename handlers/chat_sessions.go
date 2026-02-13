package handlers

import (
	"net/http"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListChatSessionsHandler returns all chat sessions for the current user (newest first).
// @Summary      List chat sessions
// @Tags         Chat
// @Produce      json
// @Header       200      {string}  X-User-ID  "User ID"
// @Success      200      {array}   models.ChatSession
// @Router       /api/chat/sessions [get]
func (h *Handlers) ListChatSessionsHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	if err := h.db.EnsureDefaultChatSession(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure default session"})
		return
	}
	sessions, err := h.db.ListChatSessions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// CreateChatSessionHandler creates a new chat session.
// @Summary      Create a new chat session
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        body  body      object  false  "Optional: { \"title\": \"New chat\" }"
// @Success      201   {object}  models.ChatSession
// @Router       /api/chat/sessions [post]
func (h *Handlers) CreateChatSessionHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	var body struct {
		Title string `json:"title"`
	}
	_ = c.ShouldBindJSON(&body)
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = "New chat"
	}
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	sess := &models.ChatSession{
		ID:        id,
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.db.StoreChatSession(sess); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sess)
}

// GetChatSessionHandler returns one session with its messages.
// @Summary      Get a chat session with messages
// @Tags         Chat
// @Produce      json
// @Param        id   path      string  true  "Session ID"
// @Success      200  {object}  object  "{ \"session\": ChatSession, \"messages\": StoredChatMessage[] }"
// @Router       /api/chat/sessions/{id} [get]
func (h *Handlers) GetChatSessionHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id required"})
		return
	}
	if sessionID == models.DefaultChatSessionID {
		_ = h.db.EnsureDefaultChatSession(userID)
	}
	sess, err := h.db.GetChatSession(userID, sessionID)
	if err != nil || sess == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	messages, err := h.db.GetChatSessionMessages(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": sess, "messages": messages})
}

// UpdateChatSessionHandler updates session title.
// @Summary      Update chat session title
// @Tags         Chat
// @Accept       json
// @Param        id    path      string  true   "Session ID"
// @Param        body  body      object  true   "{ \"title\": \"New title\" }"
// @Success      200   {object}  models.ChatSession
// @Router       /api/chat/sessions/{id} [put]
func (h *Handlers) UpdateChatSessionHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id required"})
		return
	}
	var body struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title required"})
		return
	}
	if err := h.db.UpdateChatSessionTitle(userID, sessionID, title); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	sess, _ := h.db.GetChatSession(userID, sessionID)
	c.JSON(http.StatusOK, sess)
}

// DeleteChatSessionHandler deletes a session and all its messages.
// @Summary      Delete a chat session
// @Tags         Chat
// @Param        id   path      string  true  "Session ID"
// @Success      204  "No Content"
// @Router       /api/chat/sessions/{id} [delete]
func (h *Handlers) DeleteChatSessionHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id required"})
		return
	}
	if sessionID == models.DefaultChatSessionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete default session"})
		return
	}
	if err := h.db.DeleteChatSession(userID, sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
