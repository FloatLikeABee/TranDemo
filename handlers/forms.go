package handlers

import (
	"fmt"
	"net/http"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Form Template Handlers

// CreateFormTemplateHandler creates a new form template
// @Summary      Create form template
// @Description  Create a new form template for students or staff
// @Tags         Forms
// @Accept       json
// @Produce      json
// @Param        template  body      models.FormTemplate  true  "Form template"
// @Success      200       {object}  models.FormTemplate
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/forms/templates [post]
func (h *Handlers) CreateFormTemplateHandler(c *gin.Context) {
	var template models.FormTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate required fields
	if template.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form name is required"})
		return
	}
	if template.UserType != "student" && template.UserType != "staff" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type must be 'student' or 'staff'"})
		return
	}

	// Generate ID if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now().Format(time.RFC3339)
	template.CreatedAt = now
	template.UpdatedAt = now

	// Get user ID from header or use default
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}
	template.CreatedBy = userID

	// Store in database
	if err := h.db.StoreFormTemplate(&template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create form template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetFormTemplateHandler retrieves a form template by ID
// @Summary      Get form template
// @Description  Get a form template by its ID
// @Tags         Forms
// @Produce      json
// @Param        id   path      string  true  "Form template ID"
// @Success      200  {object}  models.FormTemplate
// @Failure      404  {object}  map[string]string
// @Router       /api/forms/templates/{id} [get]
func (h *Handlers) GetFormTemplateHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form template ID is required"})
		return
	}

	template, err := h.db.GetFormTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Form template not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListFormTemplatesHandler lists all form templates
// @Summary      List form templates
// @Description  Get all form templates, optionally filtered by user type
// @Tags         Forms
// @Produce      json
// @Param        user_type  query     string  false  "Filter by user type (student or staff)"
// @Success      200        {array}   models.FormTemplate
// @Failure      500        {object}  map[string]string
// @Router       /api/forms/templates [get]
func (h *Handlers) ListFormTemplatesHandler(c *gin.Context) {
	templates, err := h.db.GetAllFormTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve form templates: %v", err)})
		return
	}

	// Filter by user type if provided
	userType := c.Query("user_type")
	if userType != "" {
		var filtered []models.FormTemplate
		for _, template := range templates {
			if template.UserType == userType {
				filtered = append(filtered, template)
			}
		}
		templates = filtered
	}

	c.JSON(http.StatusOK, templates)
}

// UpdateFormTemplateHandler updates an existing form template
// @Summary      Update form template
// @Description  Update an existing form template
// @Tags         Forms
// @Accept       json
// @Produce      json
// @Param        id         path      string  true  "Form template ID"
// @Param        template   body      models.FormTemplate  true  "Updated form template"
// @Success      200        {object}  models.FormTemplate
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /api/forms/templates/{id} [put]
func (h *Handlers) UpdateFormTemplateHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form template ID is required"})
		return
	}

	// Check if template exists
	existing, err := h.db.GetFormTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form template not found"})
		return
	}

	var template models.FormTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate user type
	if template.UserType != "" && template.UserType != "student" && template.UserType != "staff" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type must be 'student' or 'staff'"})
		return
	}

	// Preserve ID and creation info
	template.ID = id
	template.CreatedAt = existing.CreatedAt
	template.CreatedBy = existing.CreatedBy
	template.UpdatedAt = time.Now().Format(time.RFC3339)

	// Store updated template
	if err := h.db.StoreFormTemplate(&template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update form template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteFormTemplateHandler deletes a form template
// @Summary      Delete form template
// @Description  Delete a form template by its ID
// @Tags         Forms
// @Produce      json
// @Param        id   path      string  true  "Form template ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/forms/templates/{id} [delete]
func (h *Handlers) DeleteFormTemplateHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form template ID is required"})
		return
	}

	if err := h.db.DeleteFormTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete form template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form template deleted successfully"})
}

// Form Answer Handlers

// CreateFormAnswerHandler creates a new form answer
// @Summary      Create form answer
// @Description  Submit a new form answer
// @Tags         Form Answers
// @Accept       json
// @Produce      json
// @Param        answer  body      models.FormAnswer  true  "Form answer"
// @Success      200     {object}  models.FormAnswer
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/forms/answers [post]
func (h *Handlers) CreateFormAnswerHandler(c *gin.Context) {
	var answer models.FormAnswer
	if err := c.ShouldBindJSON(&answer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate required fields
	if answer.FormID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID is required"})
		return
	}
	if answer.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	if answer.UserType != "student" && answer.UserType != "staff" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type must be 'student' or 'staff'"})
		return
	}

	// Verify form template exists
	formTemplate, err := h.db.GetFormTemplate(answer.FormID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form template not found"})
		return
	}

	// Generate ID if not provided
	if answer.ID == "" {
		answer.ID = uuid.New().String()
	}

	// Set form name from template
	answer.FormName = formTemplate.Name

	// Set timestamp
	answer.SubmittedAt = time.Now().Format(time.RFC3339)

	// Get user ID from header or use provided
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = answer.UserID
	}
	answer.SubmittedBy = userID

	// Store in database
	if err := h.db.StoreFormAnswer(&answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create form answer: %v", err)})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// GetFormAnswerHandler retrieves a form answer by ID
// @Summary      Get form answer
// @Description  Get a form answer by its ID
// @Tags         Form Answers
// @Produce      json
// @Param        id   path      string  true  "Form answer ID"
// @Success      200  {object}  models.FormAnswer
// @Failure      404  {object}  map[string]string
// @Router       /api/forms/answers/{id} [get]
func (h *Handlers) GetFormAnswerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form answer ID is required"})
		return
	}

	answer, err := h.db.GetFormAnswer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Form answer not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// ListFormAnswersHandler lists all form answers
// @Summary      List form answers
// @Description  Get all form answers, optionally filtered by form ID or user ID
// @Tags         Form Answers
// @Produce      json
// @Param        form_id  query     string  false  "Filter by form ID"
// @Param        user_id  query     string  false  "Filter by user ID"
// @Success      200      {array}   models.FormAnswer
// @Failure      500      {object}  map[string]string
// @Router       /api/forms/answers [get]
func (h *Handlers) ListFormAnswersHandler(c *gin.Context) {
	formID := c.Query("form_id")
	userID := c.Query("user_id")

	var answers []models.FormAnswer
	var err error

	if formID != "" {
		answers, err = h.db.GetFormAnswersByFormID(formID)
	} else if userID != "" {
		answers, err = h.db.GetFormAnswersByUserID(userID)
	} else {
		answers, err = h.db.GetAllFormAnswers()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve form answers: %v", err)})
		return
	}

	c.JSON(http.StatusOK, answers)
}

// UpdateFormAnswerHandler updates an existing form answer
// @Summary      Update form answer
// @Description  Update an existing form answer
// @Tags         Form Answers
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Form answer ID"
// @Param        answer  body      models.FormAnswer  true  "Updated form answer"
// @Success      200     {object}  models.FormAnswer
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/forms/answers/{id} [put]
func (h *Handlers) UpdateFormAnswerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form answer ID is required"})
		return
	}

	// Check if answer exists
	existing, err := h.db.GetFormAnswer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form answer not found"})
		return
	}

	var answer models.FormAnswer
	if err := c.ShouldBindJSON(&answer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate user type if provided
	if answer.UserType != "" && answer.UserType != "student" && answer.UserType != "staff" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User type must be 'student' or 'staff'"})
		return
	}

	// Preserve ID and submission info
	answer.ID = id
	answer.SubmittedAt = existing.SubmittedAt
	answer.SubmittedBy = existing.SubmittedBy

	// Update form name if form ID changed
	if answer.FormID != "" && answer.FormID != existing.FormID {
		formTemplate, err := h.db.GetFormTemplate(answer.FormID)
		if err == nil {
			answer.FormName = formTemplate.Name
		}
	} else {
		answer.FormID = existing.FormID
		answer.FormName = existing.FormName
	}

	// Store updated answer
	if err := h.db.StoreFormAnswer(&answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update form answer: %v", err)})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// DeleteFormAnswerHandler deletes a form answer
// @Summary      Delete form answer
// @Description  Delete a form answer by its ID
// @Tags         Form Answers
// @Produce      json
// @Param        id   path      string  true  "Form answer ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/forms/answers/{id} [delete]
func (h *Handlers) DeleteFormAnswerHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form answer ID is required"})
		return
	}

	if err := h.db.DeleteFormAnswer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete form answer: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form answer deleted successfully"})
}
