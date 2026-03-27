package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

type SewadarService struct {
	repo     *repository.SewadarRepository
	deptRepo *repository.DepartmentRepository
}

func NewSewadarService(repo *repository.SewadarRepository, deptRepo *repository.DepartmentRepository) *SewadarService {
	return &SewadarService{repo: repo, deptRepo: deptRepo}
}

func (s *SewadarService) FindAll(deptID *uint, centerID *uint, page, limit int) ([]models.Sewadar, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List("", deptID, centerID, offset, limit)
}

func (s *SewadarService) Search(query string, deptID *uint, centerID *uint, page, limit int) ([]models.Sewadar, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(query, deptID, centerID, offset, limit)
}

func (s *SewadarService) GetByID(id uint) (*models.Sewadar, error) {
	return s.repo.FindByID(id)
}

func (s *SewadarService) GetByUUID(uuid string) (*models.Sewadar, error) {
	return s.repo.FindByUUID(uuid)
}

func (s *SewadarService) Create(s2 *models.Sewadar) error {
	return s.repo.Create(s2)
}

func (s *SewadarService) Update(s2 *models.Sewadar) error {
	return s.repo.Update(s2)
}

func (s *SewadarService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *SewadarService) Transfer(id, newDeptID uint) error {
	// validate dept exists
	if _, err := s.deptRepo.FindByID(newDeptID); err != nil {
		return errors.New("destination department not found")
	}
	return s.repo.Transfer(id, &newDeptID)
}

// ParseFile reads a .csv or .xlsx file and returns sewadars to create
func (s *SewadarService) ParseFile(filename string, data []byte, centerID *uint) ([]models.Sewadar, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var rows [][]string
	var err error
	if ext == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		rows, err = reader.ReadAll()
	} else {
		f, err2 := excelize.OpenReader(bytes.NewReader(data))
		if err2 != nil {
			return nil, fmt.Errorf("invalid excel file: %w", err2)
		}
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, errors.New("excel file has no sheets")
		}
		rows, err = f.GetRows(sheets[0])
	}
	if err != nil {
		return nil, err
	}
	return s.parseRows(rows, centerID)
}

func (s *SewadarService) parseRows(rows [][]string, centerID *uint) ([]models.Sewadar, error) {
	if len(rows) < 2 {
		return nil, errors.New("file must have a header row + at least one data row")
	}
	var sewadars []models.Sewadar
	for i, row := range rows[1:] {
		// New 5-column mapping based on user requirements:
		// A(0): ID, B(1): Name, C(2): Parent/Spouse Name, D(3): Gender, E(4): Badge Status
		if len(row) < 5 {
			fmt.Printf("Skipping row %d: insufficient columns (need 5, got %d)\n", i+2, len(row))
			continue
		}
		
		// All 5 fields are required
		if row[0] == "" || row[1] == "" || row[2] == "" || row[3] == "" || row[4] == "" {
			fmt.Printf("Skipping row %d: missing required fields\n", i+2)
			continue
		}
		
		sw := models.Sewadar{
			SewadarCode:       row[0],
			Name:              row[1],
			ParentSpouseName:  row[2],
			Gender:            row[3],
			BadgeStatus:       row[4],
		}
		
		// Handle CenterID
		if centerID != nil {
			sw.CenterID = *centerID
		}
		
		if sw.CenterID == 0 {
			return nil, fmt.Errorf("row %d: missing center_id", i+2)
		}
		
		sewadars = append(sewadars, sw)
	}
	return sewadars, nil
}

func (s *SewadarService) BulkCreate(sewadars []models.Sewadar) error {
	return s.repo.BulkCreate(sewadars)
}

// ExportExcel generates an Excel file of all sewadars
func (s *SewadarService) ExportExcel(deptID *uint, centerID *uint) ([]byte, error) {
	// For export, we fetch a large number (e.g., 10000) or all without pagination if we add a dedicated method.
	// For now, let's use a very large limit.
	sewadars, _, err := s.repo.List("", deptID, centerID, 0, 10000)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "Sewadars"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Sewadar ID", "Name", "Center", "Department", "Phone", "Email", "Parent/Spouse Name", "Gender", "Badge Status", "Created At"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, sw := range sewadars {
		row := i + 2
		centerName := ""
		if sw.Center != nil {
			centerName = sw.Center.Name
		}
		deptName := ""
		if sw.Department != nil {
			deptName = sw.Department.Name
		}
		vals := []interface{}{sw.SewadarCode, sw.Name, centerName, deptName, sw.Phone, sw.Email, sw.ParentSpouseName, sw.Gender, sw.BadgeStatus, sw.CreatedAt.Format(time.RFC3339)}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
