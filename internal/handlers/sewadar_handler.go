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
	userDept := middleware.GetUserDeptID(c)
	var deptFilter *uint
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		deptFilter = &uid
	} else if role != models.RoleCenterAdmin {
		deptFilter = userDept
	}
	sewadars, err := h.svc.GetAll(deptFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sewadars)
}

func (h *SewadarHandler) Search(c *gin.Context) {
	q := c.Query("q")
	var deptFilter *uint
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		id, _ := strconv.Atoi(deptIDStr)
		uid := uint(id)
		deptFilter = &uid
	}
	sewadars, err := h.svc.Search(q, deptFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sewadars)
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

func (h *SewadarHandler) Create(c *gin.Context) {
	var s models.Sewadar
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := c.ShouldBindJSON(s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.ID = uint(id)
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
	sewadars, err := h.svc.ParseFile(file.Filename, data)
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
	data, err := h.svc.ExportExcel(deptFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=sewadars.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
