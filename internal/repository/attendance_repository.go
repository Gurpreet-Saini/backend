package repository

import (
	"time"

	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
)

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

type AttendanceFilter struct {
	DepartmentID *uint
	SewadarID    *uint
	DateFrom     *time.Time
	DateTo       *time.Time
}

func (r *AttendanceRepository) FindAll(filter AttendanceFilter) ([]models.Attendance, error) {
	var records []models.Attendance
	q := r.db.Preload("Sewadar").Preload("Department").Preload("MarkedByUser")
	if filter.DepartmentID != nil {
		q = q.Where("department_id = ?", *filter.DepartmentID)
	}
	if filter.SewadarID != nil {
		q = q.Where("sewadar_id = ?", *filter.SewadarID)
	}
	if filter.DateFrom != nil {
		q = q.Where("date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("date <= ?", *filter.DateTo)
	}
	err := q.Order("date DESC, check_in DESC").Find(&records).Error
	return records, err
}

func (r *AttendanceRepository) FindByID(id uint) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Preload("Sewadar").Preload("Department").First(&a, id).Error
	return &a, err
}

func (r *AttendanceRepository) FindBySewadarAndDate(sewadarID uint, date time.Time) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Where("sewadar_id = ? AND date = ?", sewadarID, date.Format("2006-01-02")).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttendanceRepository) Create(a *models.Attendance) error {
	return r.db.Create(a).Error
}

func (r *AttendanceRepository) Update(a *models.Attendance) error {
	return r.db.Save(a).Error
}

func (r *AttendanceRepository) CountToday() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&models.Attendance{}).Where("date = ?", today).Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) TodayByDept() ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT d.id, d.name, COUNT(a.id) as count
		FROM departments d
		LEFT JOIN attendances a ON a.department_id = d.id AND a.date = ?
		WHERE d.deleted_at IS NULL
		GROUP BY d.id, d.name
		ORDER BY d.name
	`, today).Scan(&results).Error
	return results, err
}
