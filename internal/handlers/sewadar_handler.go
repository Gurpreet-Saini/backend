package handlers

import (
	"net/http"
	"strconv"

	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SewadarHandler struct {
	svc *service.SewadarService
}

func NewSewadarHandler(svc *service.SewadarService) *SewadarHandler {
	return &SewadarHandler{svc: svc}
}

func (h *SewadarHandler) List(c *gin.Context) {
	role := middleware.GetUserRole(c)
	var deptFilter *uint
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		deptFilter = &uid
	}

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

	// Pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if limit < 1 {
		limit = 25
	}

	var sewadars []models.Sewadar
	var total int64
	var err error

	if q := c.Query("q"); q != "" {
		sewadars, total, err = h.svc.Search(q, deptFilter, centerFilter, page, limit)
	} else {
		sewadars, total, err = h.svc.FindAll(deptFilter, centerFilter, page, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sewadars,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *SewadarHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	s, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sewadar not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SewadarHandler) GetByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	s, err := h.svc.GetByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sewadar not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SewadarHandler) Create(c *gin.Context) {
	var s models.Sewadar
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := middleware.GetUserRole(c)
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		if centerID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "user has no assigned center"})
			return
		}
		s.CenterID = *centerID
	} else if s.CenterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "center_id is required for super admin"})
		return
	}

	if err := h.svc.Create(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *SewadarHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	s, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sewadar not found"})
		return
	}

	role := middleware.GetUserRole(c)
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		if centerID == nil || *centerID != s.CenterID {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions for this center"})
			return
		}
	}

	if err := c.ShouldBindJSON(s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Restore ID and CenterID if not super admin
	s.ID = uint(id)
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		s.CenterID = *centerID
	}

	if err := h.svc.Update(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SewadarHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *SewadarHandler) Transfer(c *gin.Context) {
	var req struct {
		SewadarID    uint `json:"sewadar_id" binding:"required"`
		DepartmentID uint `json:"department_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// RBAC Check
	role := middleware.GetUserRole(c)
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		sw, err := h.svc.GetByID(req.SewadarID)
		if err != nil || (centerID != nil && sw.CenterID != *centerID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions for this sewadar"})
			return
		}
	}

	if err := h.svc.Transfer(req.SewadarID, req.DepartmentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sewadar transferred successfully"})
}

func (h *SewadarHandler) BulkUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not open file"})
		return
	}
	defer f.Close()
	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		return
	}
	role := middleware.GetUserRole(c)
	var centerID *uint
	if role == models.RoleSuperAdmin {
		if cIdStr := c.PostForm("center_id"); cIdStr != "" {
			if id, err := strconv.Atoi(cIdStr); err == nil {
				uid := uint(id)
				centerID = &uid
			}
		}
	} else {
		centerID = middleware.GetUserCenterID(c)
	}
	
	sewadars, err := h.svc.ParseFile(file.Filename, data, centerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BulkCreate(sewadars); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "sewadars created", "count": len(sewadars)})
}

func (h *SewadarHandler) Export(c *gin.Context) {
	var deptFilter *uint
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		deptFilter = &uid
	}

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

	data, err := h.svc.ExportExcel(deptFilter, centerFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=sewadars.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
