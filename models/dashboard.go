package models

type Overview struct {
	TotalSpent    float64 `json:"total_spent"`
	ThisWeekSpent float64 `json:"this_week_spent"`
	LeftInBudget  float64 `json:"left_in_budget"`
}

type CategorySplit struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
}

type DailyExpense struct {
	Date        string  `json:"date"`
	TotalAmount float64 `json:"total_amount"`
}

type DashboardResponse struct {
	Overview      Overview       `json:"overview"`
	DailyExpenses []DailyExpense `json:"daily_expenses"`
}
