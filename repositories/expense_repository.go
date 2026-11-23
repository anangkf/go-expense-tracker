package repositories

import (
	"database/sql"
	"fmt"
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// EXPENSE SORT ALLOWLIST (query name -> SQL column)
var expenseSortColumns = map[string]string{
	"id":             "expenses.id",
	"name":           "expenses.name",
	"amount":         "expenses.amount",
	"category_id":    "expenses.category_id",
	"bucket_type_id": "expenses.bucket_type_id",
	"created_at":     "expenses.created_at",
	"updated_at":     "expenses.updated_at",
}

// EXPENSE FILTER ALLOWLIST (query name -> SQL expression, cast to text for ILIKE)
var expenseFilterColumns = map[string]string{
	"id":             "expenses.id::text",
	"name":           "expenses.name",
	"amount":         "expenses.amount::text",
	"category_id":    "expenses.category_id::text",
	"bucket_type_id": "expenses.bucket_type_id::text",
	"created_at":     "expenses.created_at::text",
	"updated_at":     "expenses.updated_at::text",
}

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) GetByUserID(userID uint, queryParams utils.QueryParams) (*[]models.Expense, int64, int64, error) {
	var expenses []models.Expense
	var total int64

	query := r.db.Model(&models.Expense{}).Where("expenses.user_id = ?", userID)
	query = query.Joins("Category")

	// APPLY FILTERS
	if value := queryParams.Filters["category_name"]; value != "" {
		query = query.Where(`"Category"."name" ILIKE ?`, "%"+value+"%")
	}
	if value := queryParams.Filters["category_type"]; value != "" {
		query = query.Where(`"Category"."type" = ?`, value)
	}
	for column, value := range utils.SafeFilters(queryParams.Filters, expenseFilterColumns) {
		if value != "" {
			query = query.Where(column+" ILIKE ?", "%"+value+"%")
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
	order := "desc"
	if strings.ToLower(queryParams.Order) == "asc" {
		order = "asc"
	}
	sortColumn := utils.SafeSort(queryParams.SortBy, expenseSortColumns, "expenses.created_at")
	query = query.Order(sortColumn + " " + order)

	// APPLY PAGINATION
	offset := (queryParams.Page - 1) * queryParams.Limit
	query = query.Offset(offset).Limit(queryParams.Limit)

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

func (r *ExpenseRepository) GetTotalIncomeAndExpenses(userID uint, queryParams utils.QueryParams) (float64, float64, error) {
	var totalIncome, totalExpenses float64
	expenseStartDate := queryParams.ExpenseStartDate
	expenseEndDate := queryParams.ExpenseEndDate

	if expenseStartDate == "" || expenseEndDate == "" {
		// DEFAULT EXPENSE DATE RANGE
		now := time.Now()
		firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

		expenseStartDate = firstDayOfMonth.Format("2006-01-02")
		expenseEndDate = lastDayOfMonth.Format("2006-01-02")
	}

	// PARSE DATE RANGE
	loc := time.Now().Location()
	startDate, err := time.ParseInLocation("2006-01-02", expenseStartDate, loc)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid expense_start_date: %w", err)
	}
	endDate, err := time.ParseInLocation("2006-01-02", expenseEndDate, loc)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid expense_end_date: %w", err)
	}
	from := utils.StartOfDay(startDate)
	to := utils.EndOfDayExclusive(endDate)

	// CALCULATE TOTAL INCOME
	if err := r.sumByCategoryType(userID, "income", from, to).Scan(&totalIncome); err != nil {
		return 0, 0, err
	}

	// CALCULATE TOTAL EXPENSES
	if err := r.sumByCategoryType(userID, "expense", from, to).Scan(&totalExpenses); err != nil {
		return 0, 0, err
	}

	return totalIncome, totalExpenses, nil
}

// SUM AMOUNT BY CATEGORY TYPE WITHIN [from, to), EXCLUDING SOFT-DELETED ROWS
func (r *ExpenseRepository) sumByCategoryType(userID uint, categoryType string, from, to time.Time) *sql.Row {
	return r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id AND categories.deleted_at IS NULL").
		Where("expenses.user_id = ? AND expenses.deleted_at IS NULL AND categories.type = ? AND expenses.created_at >= ? AND expenses.created_at < ?",
			userID, categoryType, from, to).
		Select("COALESCE(SUM(expenses.amount), 0)").
		Row()
}
