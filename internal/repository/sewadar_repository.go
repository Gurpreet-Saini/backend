package repository

import (
	"fmt"
	"strings"

	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
)

type SewadarRepository struct {
	db *gorm.DB
}

func NewSewadarRepository(db *gorm.DB) *SewadarRepository {
	return &SewadarRepository{db: db}
}

func (r *SewadarRepository) FindAll(deptID *uint) ([]models.Sewadar, error) {
	var sewadars []models.Sewadar
	q := r.db.Preload("Department").Preload("Department.Center")
	if deptID != nil {
		q = q.Where("department_id = ?", *deptID)
	}
	err := q.Find(&sewadars).Error
	return sewadars, err
}

func (r *SewadarRepository) FindByID(id uint) (*models.Sewadar, error) {
	var s models.Sewadar
	err := r.db.Preload("Department").First(&s, id).Error
	return &s, err
}

func (r *SewadarRepository) FindBySewadarID(empID string) (*models.Sewadar, error) {
	var s models.Sewadar
	err := r.db.Where("sewadar_id = ?", empID).First(&s).Error
	return &s, err
}

func (r *SewadarRepository) Search(query string, deptID *uint) ([]models.Sewadar, error) {
	var sewadars []models.Sewadar
	q := r.db.Preload("Department")
	if query != "" {
		pattern := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		q = q.Where("LOWER(name) LIKE ? OR LOWER(employee_id) LIKE ?", pattern, pattern)
	}
	if deptID != nil {
		q = q.Where("department_id = ?", *deptID)
	}
	err := q.Limit(20).Find(&sewadars).Error
	return sewadars, err
}

func (r *SewadarRepository) Create(s *models.Sewadar) error {
	return r.db.Create(s).Error
}

func (r *SewadarRepository) Update(s *models.Sewadar) error {
	return r.db.Save(s).Error
}

func (r *SewadarRepository) Delete(id uint) error {
	return r.db.Delete(&models.Sewadar{}, id).Error
}

func (r *SewadarRepository) BulkCreate(sewadars []models.Sewadar) error {
	return r.db.CreateInBatches(sewadars, 100).Error
}

func (r *SewadarRepository) Transfer(id uint, newDeptID *uint) error {
	return r.db.Model(&models.Sewadar{}).Where("id = ?", id).Update("department_id", newDeptID).Error
}
