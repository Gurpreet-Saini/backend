package repository

import (
	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) Create(feedback *models.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *FeedbackRepository) FindAll() ([]models.Feedback, error) {
	var feedbacks []models.Feedback
	err := r.db.Preload("User").Order("created_at desc").Find(&feedbacks).Error
	return feedbacks, err
}

func (r *FeedbackRepository) Delete(id uint) error {
	return r.db.Delete(&models.Feedback{}, id).Error
}

func (r *FeedbackRepository) MarkAsRead(id uint) error {
	return r.db.Model(&models.Feedback{}).Where("id = ?", id).Update("is_read", true).Error
}
