package handlers

import (
	"net/http"
	"strconv"

	"attendancemgmt/backend/internal/middleware"
	"attendancemgmt/backend/internal/models"
	"attendancemgmt/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	repo *repository.InventoryRepository
}

func NewInventoryHandler(repo *repository.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

// ListItems godoc
// @Summary List all inventory items
// @Description Get a list of all items, optionally filtered by center, department, or category.
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param center_id query int false "Center ID filter"
// @Param department_id query int false "Department ID filter"
// @Param category query string false "Category filter"
// @Success 200 {array} models.Item
// @Failure 401 {object} map[string]string
// @Router /api/inventory [get]
func (h *InventoryHandler) ListItems(c *gin.Context) {
	role := middleware.GetUserRole(c)
	
	filter := repository.InventoryFilter{}
	if role == models.RoleSuperAdmin {
		if centerIDStr := c.Query("center_id"); centerIDStr != "" {
			if id, err := strconv.Atoi(centerIDStr); err == nil {
				uid := uint(id)
				filter.CenterID = &uid
			}
		}
	} else {
		filter.CenterID = middleware.GetUserCenterID(c)
	}

	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		if id, err := strconv.Atoi(deptIDStr); err == nil {
			uid := uint(id)
			filter.DepartmentID = &uid
		}
	}
	
	if cat := c.Query("category"); cat != "" {
		filter.Category = cat
	}

	items, err := h.repo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetItem godoc
// @Summary Get item by ID
// @Description Get details of a specific item by its internal ID
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} models.Item
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/inventory/{id} [get]
func (h *InventoryHandler) GetItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateItem godoc
// @Summary Create an item
// @Description Create a new inventory item
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item body models.Item true "Item details"
// @Success 201 {object} models.Item
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/inventory [post]
func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := middleware.GetUserRole(c)
	if role != models.RoleSuperAdmin {
		centerID := middleware.GetUserCenterID(c)
		if centerID != nil {
			item.CenterID = *centerID
		}
	}
	
	if item.CenterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "center_id is required"})
		return
	}

	if err := h.repo.Create(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// UpdateItem godoc
// @Summary Update an item
// @Description Update an existing inventory item
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param item body models.Item true "Updated item details"
// @Success 200 {object} models.Item
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/inventory/{id} [put]
func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	// Basic security check to ensure they can modify it
	role := middleware.GetUserRole(c)
	if role != models.RoleSuperAdmin {
		cid := middleware.GetUserCenterID(c)
		if cid == nil || *cid != item.CenterID {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
	}

	if err := c.ShouldBindJSON(item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.ID = uint(id)
	
	if role != models.RoleSuperAdmin {
		cid := middleware.GetUserCenterID(c)
		item.CenterID = *cid // prevent moving to another center
	}

	if err := h.repo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteItem godoc
// @Summary Delete an item
// @Description Delete an inventory item and its transaction history
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/inventory/{id} [delete]
func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}

type StockUpdateReq struct {
	QuantityChanged int    `json:"quantity_changed" binding:"required"`
	TransactionType string `json:"transaction_type" binding:"required"`
	Remarks         string `json:"remarks"`
}

// UpdateStock godoc
// @Summary Modify item stock
// @Description Adjust stock of an item (ADD, SUBTRACT, SET)
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param transaction body StockUpdateReq true "Stock transaction details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/inventory/{id}/stock [post]
func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req StockUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ttype := models.TransactionType(req.TransactionType)
	if ttype != models.TransactionTypeAdd && ttype != models.TransactionTypeSubtract && ttype != models.TransactionTypeSet {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction_type"})
		return
	}
	
	// Prevent subtract if quantity change is negative (we expect positive number for ADD/SUBTRACT)
	if req.QuantityChanged < 0 && (ttype == models.TransactionTypeAdd || ttype == models.TransactionTypeSubtract) {
		req.QuantityChanged = -req.QuantityChanged
	}

	userID := middleware.GetUserID(c)

	tx := &models.InventoryTransaction{
		ItemID:          uint(id),
		QuantityChanged: req.QuantityChanged,
		TransactionType: ttype,
		Remarks:         req.Remarks,
		MarkedBy:        userID,
	}

	if err := h.repo.AddTransaction(tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stock updated"})
}

// GetTransactions godoc
// @Summary List item transactions
// @Description View transaction history for a specific item
// @Tags inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {array} models.InventoryTransaction
// @Failure 401 {object} map[string]string
// @Router /api/inventory/{id}/transactions [get]
func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	txs, err := h.repo.GetTransactions(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, txs)
}
