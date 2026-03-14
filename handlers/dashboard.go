package handlers

import (
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardRepo *repositories.DashboardRepository
	userRepo      *repositories.UserRepository
}

func NewDashboardHandler(dashboardRepo *repositories.DashboardRepository, userRepo *repositories.UserRepository) *DashboardHandler {
	return &DashboardHandler{
		dashboardRepo: dashboardRepo,
		userRepo:      userRepo,
	}
}

// GetDashboard retrieves the dashboard overview for the authenticated user
// @Summary Get dashboard overview
// @Description Get the dashboard overview for the authenticated user, including total expenses, categories, and daily expenses
// @Tags dashboard
// @Accept json
// @Produce json
// @Param start_date query string false "Start date for the dashboard period (YYYY-MM-DD)"
// @Param end_date query string false "End date for the dashboard period (YYYY-MM-DD)"
// @Success 200 {object} utils.Response[models.DashboardResponse]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, -1)
	startDateStr := c.DefaultQuery("start_date", startOfMonth.Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", endOfMonth.Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format. Please use YYYY-MM-DD.")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format. Please use YYYY-MM-DD.")
		return
	}

	overview, err := h.dashboardRepo.GetOverview(userID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard overview")
		return
	}

	dailyExpenses, err := h.dashboardRepo.GetDailyExpenses(userID, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get daily expenses")
		return
	}

	response := gin.H{
		"overview":       overview,
		"daily_expenses": dailyExpenses,
	}

	utils.SuccessResponse(c, http.StatusOK, "Dashboard overview retrieved successfully", response)
}
