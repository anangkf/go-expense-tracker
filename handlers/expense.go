package handlers

import (
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ExpenseHandler struct {
	expenseRepo *repositories.ExpenseRepository
	userRepo    *repositories.UserRepository
	validator   *validator.Validate
}

func NewExpenseHandler(expenseRepo *repositories.ExpenseRepository, userRepo *repositories.UserRepository) *ExpenseHandler {
	return &ExpenseHandler{
		expenseRepo: expenseRepo,
		userRepo:    userRepo,
		validator:   validator.New(),
	}
}

// GET EXPENSES BY USER ID
// GetExpensesByUserID godoc
// @Summary Get expenses by user ID
// @Description Get all expenses for the authenticated user
// @Tags expenses
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[[]models.Expense]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expenses [get]
func (h *ExpenseHandler) GetExpensesByUserID(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET EXPENSES BY USER ID
	expenses, err := h.expenseRepo.GetByUserID(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expenses")
		return
	}

	// RETURN EXPENSES
	utils.SuccessResponse(c, http.StatusOK, "Expenses retrieved successfully", expenses)
}
