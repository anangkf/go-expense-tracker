package handlers

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ExpenseHandler struct {
	expenseRepo  *repositories.ExpenseRepository
	userRepo     *repositories.UserRepository
	categoryRepo *repositories.CategoryRepository
	validator    *validator.Validate
}

func NewExpenseHandler(expenseRepo *repositories.ExpenseRepository, userRepo *repositories.UserRepository, categoryRepo *repositories.CategoryRepository) *ExpenseHandler {
	return &ExpenseHandler{
		expenseRepo:  expenseRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		validator:    validator.New(),
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

// CREATE EXPENSE
// CreateExpense godoc
// @Summary Create a new expense
// @Description Create a new expense for the authenticated user
// @Tags expenses
// @Accept  json
// @Produce  json
// @Param   expense  body  models.ExpenseRequest  true  "Expense data"
// @Success 201 {object} utils.Response[models.Expense]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expenses [post]
func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
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

	// VALIDATE REQUEST BODY
	var req models.ExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// VALIDATE CATEGORY ID
	category, err := h.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// VALIDATE CATEGORY BELONGING TO USER
	if !category.IsDefault {
		if category.UserID == nil || *category.UserID != user.ID {
			utils.ErrorResponse(c, http.StatusBadRequest, "Category does not belong to this user")
			return
		}
	}

	// CREATE EXPENSE
	expense := models.Expense{
		Name:       req.Name,
		Amount:     req.Amount,
		UserID:     user.ID,
		CategoryID: category.ID,
	}

	// SAVE EXPENSE
	if err := h.expenseRepo.Create(&expense); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create expense")
		return
	}

	// SET CATEGORY TO EXPENSE STRUCT MANUALLY
	expense.Category = *category

	// RETURN CREATED EXPENSE
	utils.SuccessResponse(c, http.StatusCreated, "Expense created successfully", expense)
}

// GET EXPENSE BY ID
// GetExpenseByID godoc
// @Summary Get an expense by ID
// @Description Get an expense by ID for the authenticated user
// @Tags expenses
// @Accept  json
// @Produce  json
// @Param   id  path  int  true  "Expense ID"
// @Success 200 {object} utils.Response[models.Expense]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expenses/{id} [get]
func (h *ExpenseHandler) GetExpenseByID(c *gin.Context) {
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

	// GET EXPENSE ID FROM URL PARAM
	expenseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	// GET EXPENSE BY ID
	expense, err := h.expenseRepo.GetByID(uint(expenseID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expense not found")
		return
	}

	// CHECK IF EXPENSE BELONGS TO USER
	if expense.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Expense does not belong to this user")
		return
	}

	// RETURN EXPENSE
	utils.SuccessResponse(c, http.StatusOK, "Expense retrieved successfully", expense)
}

// UPDATE EXPENSE
// UpdateExpense godoc
// @Summary Update an expense
// @Description Update an expense for the authenticated user
// @Tags expenses
// @Accept  json
// @Produce  json
// @Param   id  path  int  true  "Expense ID"
// @Param   expense  body  models.ExpenseRequest  true  "Expense data"
// @Success 200 {object} utils.Response[models.Expense]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expenses/{id} [put]
func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
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

	// GET EXPENSE ID FROM URL PARAM
	expenseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	// GET EXPENSE BY ID
	expense, err := h.expenseRepo.GetByID(uint(expenseID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expense not found")
		return
	}

	// CHECK IF EXPENSE BELONGS TO USER
	if expense.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Expense does not belong to this user")
		return
	}

	// VALIDATE REQUEST BODY
	var req models.ExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// VALIDATE CATEGORY ID
	category, err := h.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// VALIDATE CATEGORY BELONGING TO USER
	if !category.IsDefault {
		if category.UserID == nil || *category.UserID != user.ID {
			utils.ErrorResponse(c, http.StatusBadRequest, "Category does not belong to this user")
			return
		}
	}

	// UPDATE EXPENSE FIELDS
	expense.Name = req.Name
	expense.Amount = req.Amount
	expense.CategoryID = category.ID

	// SAVE UPDATED EXPENSE
	if err := h.expenseRepo.Update(expense); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update expense")
		return
	}

	// RETURN UPDATED EXPENSE
	utils.SuccessResponse(c, http.StatusOK, "Expense updated successfully", expense)
}

// DELETE EXPENSE
// DeleteExpense godoc
// @Summary Delete an expense
// @Description Delete an expense for the authenticated user
// @Tags expenses
// @Accept  json
// @Produce  json
// @Param   id  path  int  true  "Expense ID"
// @Success 200 {object} utils.Response[models.Expense]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expenses/{id} [delete]
func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
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

	// GET EXPENSE ID FROM URL PARAM
	expenseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	// GET EXPENSE BY ID
	expense, err := h.expenseRepo.GetByID(uint(expenseID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expense not found")
		return
	}

	// CHECK IF EXPENSE BELONGS TO USER
	if expense.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Expense does not belong to this user")
		return
	}

	// DELETE EXPENSE
	if err := h.expenseRepo.Delete(expense); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete expense")
		return
	}

	// RETURN SUCCESS MESSAGE
	utils.SuccessResponse(c, http.StatusOK, "Expense deleted successfully", expense)
}
