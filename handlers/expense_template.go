package handlers

import (
	"errors"
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ExpenseTemplateHandler struct {
	templateRepo *repositories.ExpenseTemplateRepository
	categoryRepo *repositories.CategoryRepository
	validator    *validator.Validate
}

func NewExpenseTemplateHandler(templateRepo *repositories.ExpenseTemplateRepository, categoryRepo *repositories.CategoryRepository) *ExpenseTemplateHandler {
	return &ExpenseTemplateHandler{templateRepo: templateRepo, categoryRepo: categoryRepo, validator: validator.New()}
}

// GetDefaultTemplates godoc
// @Summary Get default expense templates
// @Description Get default expense templates for the current user
// @Tags expense-templates
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[[]models.ExpenseTemplate]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates/default [get]
func (h *ExpenseTemplateHandler) GetDefaultTemplates(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	_, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	templates, err := h.templateRepo.GetDefaultTemplates()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve default templates")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Default templates retrieved successfully", templates)
}

// GetExpenseTemplates godoc
// @Summary Get expense templates
// @Description Get expense templates for the current user
// @Tags expense-templates
// @Accept  json
// @Produce  json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param sortBy query string false "Sort by field" default(id)
// @Param order query string false "Sort order (asc or desc)" default(asc)
// @Param name query string false "Filter by category name"
// @Success 200 {object} utils.ResponseWithPagination[[]models.ExpenseTemplate]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates [get]
func (h *ExpenseTemplateHandler) GetExpenseTemplates(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// GET QUERY PARAMS
	queryParams, _ := c.Get("queryParams")

	// GET EXPENSE TEMPLATES BY USER ID
	templates, total, totalPages, err := h.templateRepo.GetByUserID(userID.(uint), queryParams.(utils.QueryParams))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expense templates")
		return
	}

	response := gin.H{
		"data":        templates,
		"total":       total,
		"page":        queryParams.(utils.QueryParams).Page,
		"limit":       queryParams.(utils.QueryParams).Limit,
		"total_pages": totalPages,
	}

	utils.SuccessResponse(c, http.StatusOK, "Expense templates retrieved successfully", response)
}

// CreateExpenseTemplate godoc
// @Summary Create expense template
// @Description Create a new expense template for the current user
// @Tags expense-templates
// @Accept  json
// @Produce  json
// @Param expense-template body models.ExpenseTemplateRequest true "Expense Template"
// @Success 201 {object} utils.Response[models.ExpenseTemplate]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates [post]
func (h *ExpenseTemplateHandler) CreateExpenseTemplate(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// BIND REQUEST BODY
	var req models.ExpenseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// VALIDATE REQUEST
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// CHECK IF CATEGORY IS DEFAULT OR BELONGS TO USER
	if !category.IsDefault {
		if category.UserID == nil || *category.UserID != userID.(uint) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Category does not belong to this user")
			return
		}
	}

	// CREATE EXPENSE TEMPLATE
	userIDPtr := userID.(uint)
	template := &models.ExpenseTemplate{
		UserID:       &userIDPtr,
		CategoryID:   &req.CategoryID,
		Name:         req.Name,
		Amount:       req.Amount,
		BucketTypeID: category.BucketTypeID,
	}
	templates := []*models.ExpenseTemplate{template}

	if err := h.templateRepo.CreateMany(templates); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create expense template")
		return
	}

	template.Category = *category

	utils.SuccessResponse(c, http.StatusCreated, "Expense template created successfully", template)
}

// CreateMultipleExpenseTemplates godoc
// @Summary Create multiple expense templates
// @Description Create multiple expense templates for the current user
// @Tags expense-templates
// @Accept  json
// @Produce  json
// @Param expense-templates body []models.ExpenseTemplateRequest true "Expense Templates"
// @Success 201 {object} utils.Response[[]models.ExpenseTemplate]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates/multiple [post]
func (h *ExpenseTemplateHandler) CreateMultipleExpenseTemplates(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// BIND REQUEST BODY
	var req []models.ExpenseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// VALIDATE REQUESTS
	for _, req := range req {
		if err := h.validator.Struct(req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	// CREATE EXPENSE TEMPLATES
	userIDPtr := userID.(uint)
	templates := make([]*models.ExpenseTemplate, 0, len(req))
	for _, req := range req {
		category, err := h.categoryRepo.GetByID(req.CategoryID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
			return
		}

		// CHECK IF CATEGORY IS DEFAULT OR BELONGS TO USER
		if !category.IsDefault {
			if category.UserID == nil || *category.UserID != userID.(uint) {
				utils.ErrorResponse(c, http.StatusBadRequest, "Category does not belong to this user")
				return
			}
		}

		template := &models.ExpenseTemplate{
			UserID:       &userIDPtr,
			CategoryID:   &req.CategoryID,
			Name:         req.Name,
			Amount:       req.Amount,
			BucketTypeID: category.BucketTypeID,
			Category:     *category,
		}
		templates = append(templates, template)
	}

	// SAVE EXPENSE TEMPLATES
	if err := h.templateRepo.CreateMany(templates); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create expense templates")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Expense templates created successfully", templates)
}

// GetExpenseTemplate godoc
// @Summary Get expense template
// @Description Get expense template by ID
// @Tags expense-templates
// @Accept  json
// @Produce  json
// @Param id path int true "Expense Template ID"
// @Success 200 {object} utils.Response[models.ExpenseTemplate]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 403 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates/{id} [get]
func (h *ExpenseTemplateHandler) GetExpenseTemplate(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// GET TEMPLATE ID FROM URL PARAMS
	templateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// GET EXPENSE TEMPLATE BY ID
	template, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Expense template not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expense template")
		return
	}

	// CHECK IF TEMPLATE BELONGS TO USER OR IS DEFAULT TEMPLATE
	if template.UserID != nil && *template.UserID != userID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Expense template does not belong to this user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expense template retrieved successfully", template)
}

// UpdateExpenseTemplate godoc
// @Summary Update expense template
// @Description Update expense template by ID
// @Tags expense-templates
// @Accept json
// @Produce json
// @Param id path int true "Expense Template ID"
// @Param expense-template body models.ExpenseTemplateRequest true "Expense Template"
// @Success 200 {object} utils.Response[models.ExpenseTemplate]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 403 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates/{id} [put]
func (g *ExpenseTemplateHandler) UpdateExpenseTemplate(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// GET TEMPLATE ID FROM URL PARAMS
	templateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// GET EXPENSE TEMPLATE BY ID
	template, err := g.templateRepo.GetByID(uint(templateID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Expense template not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expense template")
		return
	}

	// CHECK IF TEMPLATE BELONGS TO USER
	if template.UserID == nil || *template.UserID != userID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Expense template does not belong to this user")
		return
	}

	// BIND REQUEST BODY
	var req models.ExpenseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// VALIDATE REQUEST
	if err := g.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// CHECK IF CATEGORY IS DEFAULT OR BELONGS TO USER
	category, err := g.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}
	if !category.IsDefault {
		if category.UserID == nil || *category.UserID != userID.(uint) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Category does not belong to this user")
			return
		}
	}

	// UPDATE EXPENSE TEMPLATE
	template.Name = req.Name
	template.Amount = req.Amount
	template.CategoryID = &req.CategoryID
	template.BucketTypeID = category.BucketTypeID

	// SAVE CHANGES
	if err := g.templateRepo.Update(template); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update expense template")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expense template updated successfully", template)
}

// DeleteExpenseTemplate godoc
// @Summary Delete expense template
// @Description Delete expense template by ID
// @Tags expense-templates
// @Accept json
// @Produce json
// @Param id path int true "Expense Template ID"
// @Success 200 {object} utils.Response[models.ExpenseTemplate]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 403 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /expense-templates/{id} [delete]
func (h *ExpenseTemplateHandler) DeleteExpenseTemplate(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// GET TEMPLATE ID FROM URL PARAMS
	templateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// GET EXPENSE TEMPLATE BY ID
	template, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Expense template not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get expense template")
		return
	}

	// CHECK IF TEMPLATE BELONGS TO USER
	if template.UserID == nil || *template.UserID != userID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Expense template does not belong to this user")
		return
	}

	// DELETE EXPENSE TEMPLATE
	if err := h.templateRepo.Delete(template); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete expense template")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Expense template deleted successfully", template)
}
