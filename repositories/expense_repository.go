package repositories

import (
	"go-expense-tracker-api/middleware"
	"go-expense-tracker-api/models"
	"strings"

	"gorm.io/gorm"
)

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) GetByUserID(userID uint, queryParams middleware.QueryParams) (*[]models.Expense, int64, int64, error) {
	var expenses []models.Expense
	var total int64

	query := r.db.Model(&models.Expense{}).Where("expenses.user_id = ?", userID)
	query = query.Joins("Category")

	// APPLY FILTERS
	for key, value := range queryParams.Filters {
		if value != "" {
			switch key {
			case "category_name":
				query = query.Where(`"Category"."name" ILIKE ?`, "%"+value+"%")
			case "category_type":
				query = query.Where(`"Category"."type" = ?`, value)
			default:
				query = query.Where("expenses."+key+" ILIKE ?", "%"+value+"%")
			}
		}
	}

	// COUNT TOTAL RECORDS
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// CALCULATE TOTAL PAGES
	totalPages := int64(total) / int64(queryParams.Limit)
	if int64(total)%int64(queryParams.Limit) != 0 {
		totalPages++
	}

	// APPLY SORTING
	if queryParams.SortBy != "" {
		order := "asc"
		if strings.ToLower(queryParams.Order) == "desc" {
			order = "desc"
		}
		query = query.Order(queryParams.SortBy + " " + order)
	}

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	orderBy := "expenses." + queryParams.SortBy + " " + queryParams.Order
	query = query.Offset(offset).Limit(queryParams.Limit).Order(orderBy)

	if err := query.Preload("Category").Find(&expenses).Error; err != nil {
		return nil, 0, 0, err
	}

	return &expenses, total, totalPages, nil
}

func (r *ExpenseRepository) Create(expense *models.Expense) error {
	return r.db.Create(expense).Error
}

func (r *ExpenseRepository) GetByID(id uint) (*models.Expense, error) {
	var expense models.Expense

	err := r.db.Preload("Category").Where("id = ?", id).First(&expense).Error
	if err != nil {
		return nil, err
	}

	return &expense, nil
}

func (r *ExpenseRepository) Update(expense *models.Expense) error {
	return r.db.Save(expense).Error
}

func (r *ExpenseRepository) Delete(expense *models.Expense) error {
	return r.db.Delete(&expense).Error
}
