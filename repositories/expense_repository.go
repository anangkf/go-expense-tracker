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
