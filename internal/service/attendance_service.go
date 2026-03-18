package service

import (
	"bytes"
	"errors"
	"time"

	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type AttendanceService struct {
	repo *repository.AttendanceRepository
}

func NewAttendanceService(repo *repository.AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: repo}
}

func (s *AttendanceService) GetAll(filter repository.AttendanceFilter) ([]models.Attendance, error) {
	return s.repo.FindAll(filter)
}

func (s *AttendanceService) CheckIn(sewadarID uint, deptID *uint, markedBy uint) (*models.Attendance, error) {
	today := time.Now()
	existing, err := s.repo.FindBySewadarAndDate(sewadarID, today)
	if err == nil && existing != nil {
		return nil, errors.New("sewadar already checked in today")
	}

	// Database foreign key constraint will validate the department

	now := time.Now()
	record := &models.Attendance{
		SewadarID:    sewadarID,
		DepartmentID: deptID,
		Date:         time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()),
		CheckIn:      now,
		MarkedBy:     markedBy,
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	return s.repo.FindByID(record.ID)
}

func (s *AttendanceService) CheckOut(attendanceID, markedBy uint) (*models.Attendance, error) {
	record, err := s.repo.FindByID(attendanceID)
	if err != nil {
		return nil, errors.New("attendance record not found")
	}
	if record.CheckOut != nil {
		return nil, errors.New("check-out already recorded for this record")
	}
	now := time.Now()
	record.CheckOut = &now
	record.MarkedBy = markedBy
	if err := s.repo.Update(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *AttendanceService) Update(id uint, a *models.Attendance) (*models.Attendance, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}
	existing.CheckIn = a.CheckIn
	existing.CheckOut = a.CheckOut
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

type DashboardStats struct {
	TotalSewadars   int64                    `json:"total_sewadars"`
	TodayAttendance int64                    `json:"today_attendance"`
	TodayByDept     []map[string]interface{} `json:"today_by_dept"`
}

func (s *AttendanceService) GetDashboardStats(sewadarCount int64) (*DashboardStats, error) {
	todayCount, err := s.repo.CountToday()
	if err != nil {
		return nil, err
	}
	byDept, err := s.repo.TodayByDept()
	if err != nil {
		return nil, err
	}
	return &DashboardStats{
		TotalSewadars:   sewadarCount,
		TodayAttendance: todayCount,
		TodayByDept:     byDept,
	}, nil
}

// ExportExcel generates an Excel file of attendance records
func (s *AttendanceService) ExportExcel(filter repository.AttendanceFilter) ([]byte, error) {
	records, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "Attendance"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Employee ID", "Sewadar Name", "Department", "Date", "Check-In", "Check-Out"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, r := range records {
		row := i + 2
		empID, name, deptName := "", "", ""
		if r.Sewadar != nil {
			empID = r.Sewadar.SewadarCode
			name = r.Sewadar.Name
		}
		if r.Department != nil {
			deptName = r.Department.Name
		}
		checkOut := ""
		if r.CheckOut != nil {
			checkOut = r.CheckOut.Format("15:04:05")
		}
		vals := []interface{}{empID, name, deptName, r.Date.Format("2006-01-02"), r.CheckIn.Format("15:04:05"), checkOut}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
