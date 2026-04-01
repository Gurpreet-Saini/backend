package handlers

import (
	"net/http"
	"strconv"

	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// List godoc
// @Summary List all users
// @Description Get a list of all users, optionally filtered by center (Admin/SuperAdmin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]string
// @Router /api/users [get]
func (h *UserHandler) List(c *gin.Context) {
	role := middleware.GetUserRole(c)
	var centerID *uint
	if role != models.RoleSuperAdmin {
		centerID = middleware.GetUserCenterID(c)
	}

	users, err := h.repo.FindAll(centerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Create godoc
// @Summary Create a user
// @Description Create a new user record (Admin/SuperAdmin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body map[string]interface{} true "User details (username, password, role, center_id, department_id)"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var input struct {
		Username     string `json:"username" binding:"required"`
		Password     string `json:"password" binding:"required"`
		Role         string `json:"role" binding:"required"`
		CenterID     uint   `json:"center_id"`
		DepartmentID *uint  `json:"department_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := middleware.GetUserRole(c)
	user := models.User{
		Username:     input.Username,
		Role:         models.UserRole(input.Role),
		CenterID:     &input.CenterID,
		DepartmentID: input.DepartmentID,
	}
	
	// CenterAdmin cannot create SuperAdmin or CenterAdmin, only Operator. CenterAdmin cannot assign user to different center.
	switch role {
	case models.RoleCenterAdmin:
		user.Role = models.RoleOperator
		user.CenterID = middleware.GetUserCenterID(c)
	case models.RoleSuperAdmin:
		if input.CenterID == 0 {
			user.CenterID = nil
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}
	user.PasswordHash = string(hash)

	if err := h.repo.Create(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	user.PasswordHash = "" // hide password hash in response
	c.JSON(http.StatusCreated, user)
}

// Update godoc
// @Summary Update a user
// @Description Update an existing user record (SuperAdmin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body map[string]interface{} true "Updated user details (password, role, center_id, department_id)"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	id := uint(id64)

	var updateData struct {
		Password string `json:"password"`
		Role     string `json:"role"`
		CenterID uint   `json:"center_id"`
		DeptID   uint   `json:"department_id"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update fields if provided
	if updateData.Role != "" {
		user.Role = models.UserRole(updateData.Role)
	}
	if updateData.CenterID != 0 {
		user.CenterID = &updateData.CenterID
	}
	if updateData.DeptID != 0 {
		user.DepartmentID = &updateData.DeptID
	}

	if updateData.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}
		user.PasswordHash = string(hash)
	}

	if err := h.repo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Delete a user
// @Description Delete a user record (SuperAdmin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	id := uint(id64)

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
