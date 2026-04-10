package service

import (
	"errors"
	"fmt"
	"strconv"
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

	now := time.Now().UTC()
	// Store date as UTC date so it matches the UTC-parsed date filter in queries
	todayUTC := time.Now().UTC()
	record := &models.Attendance{
		SewadarID:    sewadarID,
		DepartmentID: deptID,
		Date:         time.Date(todayUTC.Year(), todayUTC.Month(), todayUTC.Day(), 0, 0, 0, 0, time.UTC),
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
	TodayByCenter   []map[string]interface{} `json:"today_by_center"`
	HistoricalTrend []map[string]interface{} `json:"historical_trend"`
}

func (s *AttendanceService) GetDashboardStats(sewadarCount int64, centerID *uint, from, to *time.Time, interval string) (*DashboardStats, error) {
	todayCount, err := s.repo.CountToday(centerID)
	if err != nil {
		return nil, err
	}
	byDept, err := s.repo.TodayByDept(centerID)
	if err != nil {
		return nil, err
	}
	var byCenter []map[string]interface{}
	if centerID == nil {
		byCenter, err = s.repo.TodayByCenter()
		if err != nil {
			return nil, err
		}
	}

	// Default values for trend if not provided
	startDate := time.Now().AddDate(0, 0, -6) // Last 7 days
	endDate := time.Now()
	if from != nil {
		startDate = *from
	}
	if to != nil {
		endDate = *to
	}
	if interval == "" {
		interval = "day"
	}

	trend, err := s.repo.GetAttendanceTrend(centerID, startDate, endDate, interval)
	if err != nil {
		trend = []map[string]interface{}{}
	}

	return &DashboardStats{
		TotalSewadars:   sewadarCount,
		TodayAttendance: todayCount,
		TodayByDept:     byDept,
		TodayByCenter:   byCenter,
		HistoricalTrend: trend,
	}, nil
}

// ExportExcel generates a highly organized, multi-sheet Excel report
func (s *AttendanceService) ExportExcel(filter repository.AttendanceFilter) ([]byte, error) {
	records, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	
	// --- Sheet 1: Sewadar Analytics (Primary View) ---
	analSheet := "Sewadar Analytics"
	f.SetSheetName("Sheet1", analSheet)
	
	type SewadarStats struct {
		ID            string
		Name          string
		CenterName    string
		DaysPresent   int
	}
	
	centerSummary := make(map[string]int)
	sewadarMap := make(map[uint]*SewadarStats)
	
	for _, r := range records {
		cName := "Unknown Center"
		if r.Sewadar != nil && r.Sewadar.Center != nil {
			cName = r.Sewadar.Center.Name
		}
		centerSummary[cName]++
		
		if sStats, ok := sewadarMap[r.SewadarID]; ok {
			sStats.DaysPresent++
		} else {
			id, name := "N/A", "N/A"
			if r.Sewadar != nil { 
				id = r.Sewadar.SewadarCode
				name = r.Sewadar.Name 
			}
			sewadarMap[r.SewadarID] = &SewadarStats{
				ID:          id,
				Name:        name,
				CenterName:  cName,
				DaysPresent: 1,
			}
		}
	}

	// Calculate total applicable days
	totalApplicableDays := 0
	if filter.DateFrom != nil && filter.DateTo != nil {
		start := time.Date(filter.DateFrom.Year(), filter.DateFrom.Month(), filter.DateFrom.Day(), 0, 0, 0, 0, time.Local)
		end := time.Date(filter.DateTo.Year(), filter.DateTo.Month(), filter.DateTo.Day(), 0, 0, 0, 0, time.Local)
		totalApplicableDays = int(end.Sub(start).Hours()/24) + 1
	}

	headersAnal := []string{"Sewadar ID", "Name", "Center", "Total Present", "Total Applicable Days", "Attendance %"}
	for i, h := range headersAnal {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(analSheet, cell, h)
	}
	f.SetCellStyle(analSheet, "A1", "F1", headerStyle)
	
	rowIdx := 2
	for _, stats := range sewadarMap {
		f.SetCellValue(analSheet, "A"+strconv.Itoa(rowIdx), stats.ID)
		f.SetCellValue(analSheet, "B"+strconv.Itoa(rowIdx), stats.Name)
		f.SetCellValue(analSheet, "C"+strconv.Itoa(rowIdx), stats.CenterName)
		f.SetCellValue(analSheet, "D"+strconv.Itoa(rowIdx), stats.DaysPresent)
		f.SetCellValue(analSheet, "E"+strconv.Itoa(rowIdx), totalApplicableDays)
		
		percent := 0.0
		if totalApplicableDays > 0 {
			percent = float64(stats.DaysPresent) / float64(totalApplicableDays) * 100
		}
		f.SetCellValue(analSheet, "F"+strconv.Itoa(rowIdx), fmt.Sprintf("%.2f%%", percent))
		rowIdx++
	}

	// --- Sheet 2: Center Summary ---
	summarySheet := "Center Summary"
	f.NewSheet(summarySheet)
	f.SetCellValue(summarySheet, "A1", "Center Name")
	f.SetCellValue(summarySheet, "B1", "Total Attendance Count")
	f.SetCellStyle(summarySheet, "A1", "B1", headerStyle)
	
	rowIdx = 2
	for name, count := range centerSummary {
		f.SetCellValue(summarySheet, "A"+strconv.Itoa(rowIdx), name)
		f.SetCellValue(summarySheet, "B"+strconv.Itoa(rowIdx), count)
		rowIdx++
	}

	// --- Sheet 3: Raw Attendance Logs ---
	logsSheet := "Attendance Logs"
	f.NewSheet(logsSheet)
	headersLogs := []string{"ID", "Name", "Center", "Dept", "Date", "Check-In", "Check-Out"}
	for i, h := range headersLogs {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(logsSheet, cell, h)
	}
	f.SetCellStyle(logsSheet, "A1", "G1", headerStyle)
	
	for i, r := range records {
		rowIdx = i + 2
		empID, name, cName, deptName := "", "N/A", "N/A", "N/A"
		if r.Sewadar != nil {
			empID = r.Sewadar.SewadarCode
			name = r.Sewadar.Name
			if r.Sewadar.Center != nil {
				cName = r.Sewadar.Center.Name
			}
		}
		if r.Department != nil {
			deptName = r.Department.Name
		}
		checkOut := ""
		if r.CheckOut != nil {
			checkOut = r.CheckOut.Format("15:04:05")
		}
		vals := []interface{}{empID, name, cName, deptName, r.Date.Format("2006-01-02"), r.CheckIn.Format("15:04:05"), checkOut}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowIdx)
			f.SetCellValue(logsSheet, cell, v)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
