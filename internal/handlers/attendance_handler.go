package handlers

import (
	"net/http"
	"strconv"
	"time"

	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"
	"attendancemgmt/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	svc        *service.AttendanceService
	sewadarSvc *service.SewadarService
}

func NewAttendanceHandler(svc *service.AttendanceService, sewadarSvc *service.SewadarService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc, sewadarSvc: sewadarSvc}
}

func (h *AttendanceHandler) List(c *gin.Context) {
	role := middleware.GetUserRole(c)
	userDept := middleware.GetUserDeptID(c)

	filter := repository.AttendanceFilter{}
	if role == models.RoleSuperAdmin {
		if centerIDStr := c.Query("center_id"); centerIDStr != "" {
			id, _ := strconv.Atoi(centerIDStr)
			uid := uint(id)
			filter.CenterID = &uid
		}
	} else {
		filter.CenterID = middleware.GetUserCenterID(c)
	}

	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		filter.DepartmentID = &uid
	} else if role == models.RoleDeptViewer && userDept != nil {
		filter.DepartmentID = userDept
	}
	if sewadarIDStr := c.Query("sewadar_id"); sewadarIDStr != "" {
		id, _ := strconv.Atoi(sewadarIDStr)
		uid := uint(id)
		filter.SewadarID = &uid
	}
	if from := c.Query("date_from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err == nil {
			filter.DateFrom = &t
		}
	}
	if to := c.Query("date_to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err == nil {
			filter.DateTo = &t
		}
	}
	records, err := h.svc.GetAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

type CheckInReq struct {
	SewadarID    uint  `json:"sewadar_id" binding:"required"`
	DepartmentID *uint `json:"department_id"`
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req CheckInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	markedBy := middleware.GetUserID(c)
	record, err := h.svc.CheckIn(req.SewadarID, req.DepartmentID, markedBy)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	markedBy := middleware.GetUserID(c)
	record, err := h.svc.CheckOut(uint(id), markedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *AttendanceHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var a models.Attendance
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.svc.Update(uint(id), &a)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *AttendanceHandler) Export(c *gin.Context) {
	filter := repository.AttendanceFilter{}
	role := middleware.GetUserRole(c)
	if role == models.RoleSuperAdmin {
		if centerIDStr := c.Query("center_id"); centerIDStr != "" {
			id, _ := strconv.Atoi(centerIDStr)
			uid := uint(id)
			filter.CenterID = &uid
		}
	} else {
		filter.CenterID = middleware.GetUserCenterID(c)
	}

	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		filter.DepartmentID = &uid
	}
	if from := c.Query("date_from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err == nil {
			filter.DateFrom = &t
		}
	}
	if to := c.Query("date_to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err == nil {
			filter.DateTo = &t
		}
	}
	data, err := h.svc.ExportExcel(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=attendance.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *AttendanceHandler) Dashboard(c *gin.Context) {
	role := middleware.GetUserRole(c)
	var centerFilter *uint
	if role == models.RoleSuperAdmin {
		if centerIDStr := c.Query("center_id"); centerIDStr != "" {
			id, _ := strconv.Atoi(centerIDStr)
			uid := uint(id)
			centerFilter = &uid
		}
	} else {
		centerFilter = middleware.GetUserCenterID(c)
	}

	_, totalSewadars, _ := h.sewadarSvc.FindAll(nil, centerFilter, 1, 1)
	stats, err := h.svc.GetDashboardStats(totalSewadars, centerFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
