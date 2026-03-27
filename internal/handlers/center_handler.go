package handlers

import (
	"net/http"
	"strconv"

	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type CenterHandler struct {
	repo *repository.CenterRepository
}

func NewCenterHandler(repo *repository.CenterRepository) *CenterHandler {
	return &CenterHandler{repo: repo}
}

func (h *CenterHandler) List(c *gin.Context) {
	centers, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, centers)
}

func (h *CenterHandler) Create(c *gin.Context) {
	var center models.Center
	if err := c.ShouldBindJSON(&center); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Create(&center); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, center)
}

func (h *CenterHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	center, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "center not found"})
		return
	}
	if err := c.ShouldBindJSON(center); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	center.ID = uint(id)
	if err := h.repo.Update(center); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, center)
}

func (h *CenterHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---

type DepartmentHandler struct {
	repo *repository.DepartmentRepository
}

func NewDepartmentHandler(repo *repository.DepartmentRepository) *DepartmentHandler {
	return &DepartmentHandler{repo: repo}
}

func (h *DepartmentHandler) List(c *gin.Context) {
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

	depts, err := h.repo.FindAll(centerFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type DeptWithCount struct {
		models.Department
		SewadarCount int64 `json:"sewadar_count"`
	}
	result := make([]DeptWithCount, 0, len(depts))
	for _, d := range depts {
		count, _ := h.repo.CountSewadars(d.ID)
		result = append(result, DeptWithCount{Department: d, SewadarCount: count})
	}
	c.JSON(http.StatusOK, result)
}

func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dept, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}
	c.JSON(http.StatusOK, dept)
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var dept models.Department
	if err := c.ShouldBindJSON(&dept); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := models.UserRole(c.GetString("role"))
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		if centerID != nil {
			dept.CenterID = *centerID
		}
	}

	if dept.CenterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "center_id is required"})
		return
	}
	if err := h.repo.Create(&dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dept)
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dept, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}
	if err := c.ShouldBindJSON(dept); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dept.ID = uint(id)
	if err := h.repo.Update(dept); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dept)
}

func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
