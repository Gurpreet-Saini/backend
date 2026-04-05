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
	CenterID     *uint
	DepartmentID *uint
	SewadarID    *uint
	DateFrom     *time.Time
	DateTo       *time.Time
}

func (r *AttendanceRepository) FindAll(filter AttendanceFilter) ([]models.Attendance, error) {
	var records []models.Attendance
	q := r.db.Select("attendances.id", "attendances.sewadar_id", "attendances.department_id", "attendances.date", "attendances.check_in", "attendances.check_out", "attendances.marked_by", "attendances.created_at", "attendances.updated_at").
		Preload("Sewadar", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "employee_id", "uuid", "center_id")
		}).
		Preload("Sewadar.Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("MarkedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		})
	if filter.CenterID != nil {
		q = q.Joins("JOIN sewadars ON sewadars.id = attendances.sewadar_id").Where("sewadars.center_id = ?", *filter.CenterID)
	}
	if filter.DepartmentID != nil {
		q = q.Where("attendances.department_id = ?", *filter.DepartmentID)
	}
	if filter.SewadarID != nil {
		q = q.Where("attendances.sewadar_id = ?", *filter.SewadarID)
	}
	if filter.DateFrom != nil {
		q = q.Where("attendances.date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		q = q.Where("attendances.date <= ?", *filter.DateTo)
	}
	err := q.Order("attendances.date DESC, attendances.check_in DESC").Find(&records).Error
	return records, err
}

func (r *AttendanceRepository) FindByID(id uint) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.Select("id", "sewadar_id", "department_id", "date", "check_in", "check_out", "marked_by", "created_at", "updated_at").
		Preload("Sewadar", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "employee_id", "center_id")
		}).
		Preload("Sewadar.Center", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Department", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		First(&a, id).Error
	return &a, err
}

func (r *AttendanceRepository) FindBySewadarAndDate(sewadarID uint, date time.Time) (*models.Attendance, error) {
	var a models.Attendance
	// Use UTC date to match how dates are stored (UTC midnight)
	err := r.db.Select("id", "sewadar_id", "department_id", "date", "check_in", "check_out", "marked_by", "created_at", "updated_at").
		Where("sewadar_id = ? AND date = ?", sewadarID, date.UTC().Format("2006-01-02")).First(&a).Error
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

func (r *AttendanceRepository) CountToday(centerID *uint) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	q := r.db.Model(&models.Attendance{}).Where("attendances.date = ?", today)
	if centerID != nil {
		q = q.Joins("JOIN sewadars ON sewadars.id = attendances.sewadar_id").Where("sewadars.center_id = ?", *centerID)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) TodayByDept(centerID *uint) ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	var results []map[string]interface{}
	
	q := r.db.Table("departments d").
		Select("d.id, d.name, COUNT(a.id) as count").
		Joins("LEFT JOIN attendances a ON a.department_id = d.id AND a.date = ?", today).
		Where("d.deleted_at IS NULL")
		
	if centerID != nil {
		q = q.Where("d.center_id = ?", *centerID)
	}
	
	err := q.Group("d.id, d.name").Order("d.name").Scan(&results).Error
	return results, err
}

func (r *AttendanceRepository) TodayByCenter() ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	var results []map[string]interface{}
	err := r.db.Table("centers c").
		Select("c.id, c.name, COUNT(a.id) as count").
		Joins("LEFT JOIN sewadars s ON s.center_id = c.id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN attendances a ON a.sewadar_id = s.id AND a.date = ?", today).
		Where("c.deleted_at IS NULL").
		Group("c.id, c.name").Order("c.name").
		Scan(&results).Error
	return results, err
}

