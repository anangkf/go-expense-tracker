package repositories

import (
	"go-expense-tracker-api/models"

	"gorm.io/gorm"
)

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) GetByUserID(userID uint) (*[]models.Expense, error) {
	var expenses []models.Expense

	err := r.db.Where("user_id = ?", userID).Find(&expenses).Error
	if err != nil {
		return nil, err
	}

	return &expenses, nil
}

func (r *ExpenseRepository) Create(expense *models.Expense) error {
	return r.db.Create(expense).Error
}

func (r *ExpenseRepository) GetByID(id uint) (*models.Expense, error) {
	var expense models.Expense

	err := r.db.Where("id = ?", id).First(&expense).Error
	if err != nil {
		return nil, err
	}

	return &expense, nil
}

func (r *ExpenseRepository) Update(expense *models.Expense) error {
	return r.db.Save(expense).Error
}
