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

// List godoc
// @Summary List all centers
// @Description Get a list of all centers
// @Tags centers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Center
// @Failure 401 {object} map[string]string
// @Router /api/centers [get]
func (h *CenterHandler) List(c *gin.Context) {
	centers, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, centers)
}

// Create godoc
// @Summary Create a center
// @Description Create a new center (SuperAdmin only)
// @Tags centers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param center body models.Center true "Center details"
// @Success 201 {object} models.Center
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/centers [post]
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

// Update godoc
// @Summary Update a center
// @Description Update an existing center (SuperAdmin only)
// @Tags centers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Center ID"
// @Param center body models.Center true "Updated center details"
// @Success 200 {object} models.Center
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/centers/{id} [put]
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

// Delete godoc
// @Summary Delete a center
// @Description Delete a center and its associated data (SuperAdmin only)
// @Tags centers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Center ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/centers/{id} [delete]
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

// List godoc
// @Summary List departments
// @Description Get a list of departments with optional center filter
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param center_id query int false "Center ID filter"
// @Success 200 {array} DeptWithCount
// @Failure 401 {object} map[string]string
// @Router /api/departments [get]
type DeptWithCount struct {
	models.Department
	SewadarCount int64 `json:"sewadar_count"`
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
	result := make([]DeptWithCount, 0, len(depts))
	for _, d := range depts {
		count, _ := h.repo.CountSewadars(d.ID)
		result = append(result, DeptWithCount{Department: d, SewadarCount: count})
	}
	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// @Summary Get department by ID
// @Description Get details of a specific department
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} models.Department
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/departments/{id} [get]
func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dept, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
		return
	}
	c.JSON(http.StatusOK, dept)
}

// Create godoc
// @Summary Create a department
// @Description Create a new department (Admin/SuperAdmin only)
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param department body models.Department true "Department details"
// @Success 201 {object} models.Department
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/departments [post]
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

// Update godoc
// @Summary Update a department
// @Description Update an existing department (Admin/SuperAdmin only)
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Param department body models.Department true "Updated department details"
// @Success 200 {object} models.Department
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/departments/{id} [put]
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

// Delete godoc
// @Summary Delete a department
// @Description Delete a department (Admin/SuperAdmin only)
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
