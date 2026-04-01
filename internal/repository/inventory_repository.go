package repository

import (
	"attendancemgmt/backend/internal/models"

	"gorm.io/gorm"
)

type InventoryFilter struct {
	CenterID     *uint
	DepartmentID *uint
	Category     string
}

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) FindAll(filter InventoryFilter) ([]models.Item, error) {
	var items []models.Item
	q := r.db.Preload("Center").Preload("Department")
	
	if filter.CenterID != nil {
		q = q.Where("center_id = ?", *filter.CenterID)
	}
	if filter.DepartmentID != nil {
		q = q.Where("department_id = ?", *filter.DepartmentID)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}

	err := q.Order("name asc").Find(&items).Error
	return items, err
}

func (r *InventoryRepository) FindByID(id uint) (*models.Item, error) {
	var item models.Item
	err := r.db.Preload("Center").Preload("Department").First(&item, id).Error
	return &item, err
}

func (r *InventoryRepository) Create(item *models.Item) error {
	return r.db.Create(item).Error
}

func (r *InventoryRepository) Update(item *models.Item) error {
	return r.db.Save(item).Error
}

func (r *InventoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Item{}, id).Error
}

func (r *InventoryRepository) AddTransaction(tx *models.InventoryTransaction) error {
	return r.db.Transaction(func(dbTx *gorm.DB) error {
		// Create the transaction record
		if err := dbTx.Create(tx).Error; err != nil {
			return err
		}

		// Update the item quantity
		updateQuery := ""
		switch tx.TransactionType {
		case models.TransactionTypeAdd:
			updateQuery = "quantity + ?"
		case models.TransactionTypeSubtract:
			updateQuery = "quantity - ?"
		case models.TransactionTypeSet:
			updateQuery = "?"
		}

		if err := dbTx.Model(&models.Item{}).Where("id = ?", tx.ItemID).UpdateColumn("quantity", gorm.Expr(updateQuery, tx.QuantityChanged)).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *InventoryRepository) GetTransactions(itemID uint) ([]models.InventoryTransaction, error) {
	var transactions []models.InventoryTransaction
	err := r.db.Preload("MarkedByUser").Where("item_id = ?", itemID).Order("created_at desc").Find(&transactions).Error
	return transactions, err
}
