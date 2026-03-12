package repositories

import (
	"go-expense-tracker-api/models"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetOverview(userID uint, startDate time.Time, endDate time.Time) (*models.Overview, error) {
	var overview models.Overview

	// TOTAL SPENT
	err := r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Where("expenses.user_id = ? AND expenses.created_at BETWEEN ? AND ? AND categories.type = ?", userID, startDate, endDate, "expense").
		Select("COALESCE(SUM(expenses.amount), 0)").
		Row().Scan(&overview.TotalSpent)
	if err != nil {
		return nil, err
	}

	// THIS WEEK SPENT
	startOfWeek := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	err = r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Where("expenses.user_id = ? AND expenses.created_at BETWEEN ? AND ? AND categories.type = ?", userID, startOfWeek, endOfWeek, "expense").
		Select("COALESCE(SUM(expenses.amount), 0)").
		Row().Scan(&overview.ThisWeekSpent)
	if err != nil {
		return nil, err
	}

	// GET INCOME CATEGORY IDs
	var categoryIDs []int

	err = r.db.Model(models.Category{}).
		Where("user_id = ? AND type = ?", userID, "income").
		Pluck("id", &categoryIDs).Error
	if err != nil {
		return nil, err
	}

	// CALCULATE LEFT IN BUDGET
	var totalIncome float64
	err = r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Where("expenses.user_id = ? AND expenses.created_at BETWEEN ? AND ? AND categories.type = ? AND categories.id IN ?", userID, startDate, endDate, "income", categoryIDs).
		Select("COALESCE(SUM(expenses.amount), 0)").
		Row().Scan(&totalIncome)
	if err != nil {
		return nil, err
	}
	overview.LeftInBudget = totalIncome - overview.TotalSpent

	return &overview, nil
}

func (r *DashboardRepository) GetDailyExpenses(userID uint, startDate time.Time, endDate time.Time) ([]models.DailyExpense, error) {
	results := []models.DailyExpense{}

	err := r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Where("expenses.user_id = ? AND expenses.created_at BETWEEN ? AND ? AND categories.type = ?", userID, startDate, endDate, "expense").
		Select("DATE(expenses.created_at) AS date, COALESCE(SUM(expenses.amount), 0) AS total_amount").
		Group("date").
		Order("date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
