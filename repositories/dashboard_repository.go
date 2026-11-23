package repositories

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/utils"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// SUM OF EXPENSES BY CATEGORY TYPE IN [from, toExclusive)
func (r *DashboardRepository) sumByType(userID uint, categoryType string, from time.Time, toExclusive time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id AND categories.deleted_at IS NULL").
		Where("expenses.user_id = ? AND expenses.deleted_at IS NULL AND expenses.created_at >= ? AND expenses.created_at < ? AND categories.type = ?",
			userID, from, toExclusive, categoryType).
		Select("COALESCE(SUM(expenses.amount), 0)").
		Row().Scan(&total)
	return total, err
}

// startDate AND endDate ARE DATES (INCLUSIVE), CONVERTED TO [start, end + 1 day)
func (r *DashboardRepository) GetOverview(userID uint, startDate time.Time, endDate time.Time) (*models.Overview, error) {
	var overview models.Overview

	from := utils.StartOfDay(startDate)
	to := utils.EndOfDayExclusive(endDate)

	// TOTAL SPENT
	totalSpent, err := r.sumByType(userID, "expense", from, to)
	if err != nil {
		return nil, err
	}
	overview.TotalSpent = totalSpent

	// THIS WEEK SPENT (WEEK STARTS ON SUNDAY)
	now := time.Now()
	startOfWeek := utils.StartOfDay(now.AddDate(0, 0, -int(now.Weekday())))
	endOfWeekExclusive := startOfWeek.AddDate(0, 0, 7)
	thisWeek, err := r.sumByType(userID, "expense", startOfWeek, endOfWeekExclusive)
	if err != nil {
		return nil, err
	}
	overview.ThisWeekSpent = thisWeek

	// TOTAL INCOME FROM ALL INCOME CATEGORIES (DEFAULT + USER'S), SCOPED BY expenses.user_id
	totalIncome, err := r.sumByType(userID, "income", from, to)
	if err != nil {
		return nil, err
	}

	// CALCULATE LEFT IN BUDGET
	overview.LeftInBudget = totalIncome - overview.TotalSpent

	return &overview, nil
}

func (r *DashboardRepository) GetDailyExpenses(userID uint, startDate time.Time, endDate time.Time) ([]models.DailyExpense, error) {
	results := []models.DailyExpense{}

	from := utils.StartOfDay(startDate)
	to := utils.EndOfDayExclusive(endDate)

	err := r.db.Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id AND categories.deleted_at IS NULL").
		Where("expenses.user_id = ? AND expenses.deleted_at IS NULL AND expenses.created_at >= ? AND expenses.created_at < ? AND categories.type = ?",
			userID, from, to, "expense").
		Select("DATE(expenses.created_at) AS date, COALESCE(SUM(expenses.amount), 0) AS total_amount").
		Group("date").
		Order("date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
