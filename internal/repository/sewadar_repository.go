package repository

import (
	"fmt"
	"strings"

	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SewadarRepository struct {
	db *gorm.DB
}

func NewSewadarRepository(db *gorm.DB) *SewadarRepository {
	return &SewadarRepository{db: db}
}

func (r *SewadarRepository) List(q string, deptID *uint, centerID *uint, offset, limit int) ([]models.Sewadar, int64, error) {
	var sewadars []models.Sewadar
	var total int64

	db := r.db.Model(&models.Sewadar{}).
		Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		})

	if q != "" {
		searchQuery := "%" + strings.ToLower(q) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(employee_id) LIKE ?", searchQuery, searchQuery)
	}

	if deptID != nil {
		db = db.Where("department_id = ?", *deptID)
	}
	if centerID != nil {
		db = db.Where("center_id = ?", *centerID)
	}

	// Count before paging
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paging with Order
	err := db.Order("id DESC").Offset(offset).Limit(limit).Find(&sewadars).Error
	return sewadars, total, err
}

func (r *SewadarRepository) FindByID(id uint) (*models.Sewadar, error) {
	var s models.Sewadar
	err := r.db.Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		}).
		First(&s, id).Error
	return &s, err
}

func (r *SewadarRepository) FindByUUID(uuid string) (*models.Sewadar, error) {
	var s models.Sewadar
	err := r.db.Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		}).
		Where("uuid = ?", uuid).First(&s).Error
	return &s, err
}

func (r *SewadarRepository) FindBySewadarID(empID string) (*models.Sewadar, error) {
	var s models.Sewadar
	err := r.db.Select("sewadars.id", "sewadars.employee_id", "sewadars.name", "sewadars.parent_spouse_name", "sewadars.gender", "sewadars.badge_status", "sewadars.center_id", "sewadars.department_id", "sewadars.phone", "sewadars.email", "sewadars.created_at", "sewadars.updated_at").
		Where("sewadars.employee_id = ?", empID).First(&s).Error
	return &s, err
}

func (r *SewadarRepository) Search(query string, deptID *uint, centerID *uint) ([]models.Sewadar, error) {
	var sewadars []models.Sewadar
	q := r.db.Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		})
	if query != "" {
		pattern := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		q = q.Where("LOWER(sewadars.name) LIKE ? OR LOWER(sewadars.employee_id) LIKE ? OR LOWER(sewadars.uuid) LIKE ?", pattern, pattern, pattern)
	}
	if deptID != nil {
		q = q.Where("sewadars.department_id = ?", *deptID)
	}
	if centerID != nil {
		q = q.Where("sewadars.center_id = ?", *centerID)
	}
	err := q.Limit(100).Find(&sewadars).Error
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
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "employee_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":               gorm.Expr("EXCLUDED.name"),
			"parent_spouse_name": gorm.Expr("EXCLUDED.parent_spouse_name"),
			"gender":             gorm.Expr("EXCLUDED.gender"),
			"badge_status":       gorm.Expr("EXCLUDED.badge_status"),
			"center_id":          gorm.Expr("EXCLUDED.center_id"),
			"department_id":      gorm.Expr("EXCLUDED.department_id"),
			"phone":              gorm.Expr("EXCLUDED.phone"),
			"email":              gorm.Expr("EXCLUDED.email"),
			"updated_at":         gorm.Expr("NOW()"),
			"deleted_at":         nil, // restore soft-deleted records
		}),
	}).CreateInBatches(sewadars, 100).Error
}


func (r *SewadarRepository) Transfer(id uint, newDeptID *uint) error {
	return r.db.Model(&models.Sewadar{}).Where("id = ?", id).Update("department_id", newDeptID).Error
}
