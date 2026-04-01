package handlers

import (
	"net/http"
	"strconv"

	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct {
	repo *repository.FeedbackRepository
}

func NewFeedbackHandler(repo *repository.FeedbackRepository) *FeedbackHandler {
	return &FeedbackHandler{repo: repo}
}

type FeedbackRequest struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Submit godoc
// @Summary Submit feedback
// @Description Submit a new feedback message from an authenticated user
// @Tags feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param feedback body FeedbackRequest true "Feedback details"
// @Success 201 {object} models.Feedback
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/feedback [post]
func (h *FeedbackHandler) Submit(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	feedback := models.Feedback{
		UserID:  &userID,
		Subject: req.Subject,
		Message: req.Message,
	}

	if err := h.repo.Create(&feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}

// List godoc
// @Summary List all feedback
// @Description Get a list of all feedback messages (SuperAdmin only)
// @Tags feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Feedback
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/feedback [get]
func (h *FeedbackHandler) List(c *gin.Context) {
	feedbacks, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feedbacks)
}

// Delete godoc
// @Summary Delete feedback
// @Description Delete a specific feedback message (SuperAdmin only)
// @Tags feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feedback ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/feedback/{id} [delete]
func (h *FeedbackHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "feedback deleted"})
}

// MarkAsRead godoc
// @Summary Mark feedback as read
// @Description Update the status of a feedback message to read (SuperAdmin only)
// @Tags feedback
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feedback ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/feedback/{id}/read [put]
func (h *FeedbackHandler) MarkAsRead(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.MarkAsRead(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "feedback marked as read"})
}
