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

func (s *SewadarService) GetAll(deptID *uint) ([]models.Sewadar, error) {
	return s.repo.FindAll(deptID)
}

func (s *SewadarService) Search(query string, deptID *uint) ([]models.Sewadar, error) {
	return s.repo.Search(query, deptID)
}

func (s *SewadarService) GetByID(id uint) (*models.Sewadar, error) {
	return s.repo.FindByID(id)
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
func (s *SewadarService) ParseFile(filename string, data []byte) ([]models.Sewadar, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return s.parseCSV(data)
	}
	// Fallback to expecting excelize support for .xls / .xlsx
	return s.parseExcel(data)
}

func (s *SewadarService) parseCSV(data []byte) ([]models.Sewadar, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid csv file: %w", err)
	}
	return s.parseRows(rows)
}

func (s *SewadarService) parseExcel(data []byte) ([]models.Sewadar, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid excel file: %w", err)
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("excel file has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	return s.parseRows(rows)
}

func (s *SewadarService) parseRows(rows [][]string) ([]models.Sewadar, error) {
	if len(rows) < 2 {
		return nil, errors.New("file must have a header row + at least one data row")
	}
	var sewadars []models.Sewadar
	for i, row := range rows[1:] {
		// New Columns: Sewadar ID, Name, Department ID, Phone, Email, Parent/Spouse Name, Gender, Badge Status
		if len(row) < 3 {
			continue // Skip blank lines or rows without minimum required
		}
		if row[0] == "" || row[1] == "" {
			return nil, fmt.Errorf("row %d: must have sewadar_id and name", i+2)
		}
		var deptID uint
		fmt.Sscan(row[2], &deptID)
		
		sw := models.Sewadar{
			SewadarCode:  row[0],
			Name:         row[1],
		}
		if deptID > 0 {
			sw.DepartmentID = &deptID
		}
		if len(row) > 3 {
			sw.Phone = row[3]
		}
		if len(row) > 4 {
			sw.Email = row[4]
		}
		if len(row) > 5 {
			sw.ParentSpouseName = row[5]
		}
		if len(row) > 6 {
			sw.Gender = row[6]
		}
		if len(row) > 7 {
			sw.BadgeStatus = row[7]
		}
		sewadars = append(sewadars, sw)
	}
	return sewadars, nil
}

func (s *SewadarService) BulkCreate(sewadars []models.Sewadar) error {
	return s.repo.BulkCreate(sewadars)
}

// ExportExcel generates an Excel file of all sewadars
func (s *SewadarService) ExportExcel(deptID *uint) ([]byte, error) {
	sewadars, err := s.repo.FindAll(deptID)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "Sewadars"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Sewadar ID", "Name", "Department", "Phone", "Email", "Parent/Spouse Name", "Gender", "Badge Status", "Created At"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, sw := range sewadars {
		row := i + 2
		deptName := ""
		if sw.Department != nil {
			deptName = sw.Department.Name
		}
		vals := []interface{}{sw.SewadarCode, sw.Name, deptName, sw.Phone, sw.Email, sw.ParentSpouseName, sw.Gender, sw.BadgeStatus, sw.CreatedAt.Format(time.RFC3339)}
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
