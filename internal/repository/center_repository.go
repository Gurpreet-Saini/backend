package repository

import (
	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
)

type CenterRepository struct {
	db *gorm.DB
}

func NewCenterRepository(db *gorm.DB) *CenterRepository {
	return &CenterRepository{db: db}
}

func (r *CenterRepository) FindAll() ([]models.Center, error) {
	var centers []models.Center
	err := r.db.Select("id", "name", "location", "created_at", "updated_at").
		Preload("Departments", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		}).Find(&centers).Error
	return centers, err
}

func (r *CenterRepository) FindByID(id uint) (*models.Center, error) {
	var center models.Center
	err := r.db.Select("id", "name", "location", "created_at", "updated_at").
		Preload("Departments", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "center_id", "name", "description")
		}).First(&center, id).Error
	return &center, err
}

func (r *CenterRepository) Create(center *models.Center) error {
	return r.db.Create(center).Error
}

func (r *CenterRepository) Update(center *models.Center) error {
	return r.db.Save(center).Error
}

func (r *CenterRepository) Delete(id uint) error {
	return r.db.Delete(&models.Center{}, id).Error
}

// ---

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) FindAll() ([]models.Department, error) {
	var depts []models.Department
	err := r.db.Select("id", "center_id", "name", "description", "created_at", "updated_at").
		Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "location")
		}).Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) FindByID(id uint) (*models.Department, error) {
	var dept models.Department
	err := r.db.Select("id", "center_id", "name", "description", "created_at", "updated_at").
		Preload("Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "location")
		}).First(&dept, id).Error
	return &dept, err
}

func (r *DepartmentRepository) FindByCenterID(centerID uint) ([]models.Department, error) {
	var depts []models.Department
	err := r.db.Select("id", "center_id", "name", "description", "created_at", "updated_at").
		Where("center_id = ?", centerID).Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) Create(dept *models.Department) error {
	return r.db.Create(dept).Error
}

func (r *DepartmentRepository) Update(dept *models.Department) error {
	return r.db.Save(dept).Error
}

func (r *DepartmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Department{}, id).Error
}

func (r *DepartmentRepository) CountSewadars(deptID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Sewadar{}).Where("department_id = ?", deptID).Count(&count).Error
	return count, err
}
