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

type CategoryHandler struct {
	categoryRepo *repositories.CategoryRepository
	userRepo     *repositories.UserRepository
	validator    *validator.Validate
}

func NewCategoryHandler(categoryRepo *repositories.CategoryRepository, userRepo *repositories.UserRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
		validator:    validator.New(),
	}
}

// GET CATEGORIES BY USER ID
// GetCategoriesByUserID godoc
// @Summary Get categories by user ID
// @Description Get all categories for the authenticated user
// @Tags categories
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[[]models.Category]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /categories [get]
func (h *CategoryHandler) GetCategoriesByUserID(c *gin.Context) {
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

	// GET CATEGORIES BY USER ID
	categories, err := h.categoryRepo.GetByUserID(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", categories)
}

// CREATE CATEGORY
// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category for the authenticated user
// @Tags categories
// @Accept  json
// @Produce  json
// @Param request body models.CategoryRequest true "Category data"
// @Success 201 {object} utils.Response[models.Category]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
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

	var req models.Category

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// CREATE CATEGORY
	category := &models.Category{
		Name:      req.Name,
		UserID:    &user.ID,
		Type:      req.Type,
		IsDefault: false,
	}
	categories := []*models.Category{category}

	if err := h.categoryRepo.CreateMany(categories); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Category created successfully", categories[0])
}

// UPDATE CATEGORY
// UpdateCategory godoc
// @Summary Update a category
// @Description Update a category for the authenticated user
// @Tags categories
// @Accept  json
// @Produce  json
// @Param id path int true "Category ID"
// @Param request body models.CategoryRequest true "Category data"
// @Success 200 {object} utils.Response[models.Category]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
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

	// GET CATEGORY ID FROM PATH
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// GET CATEGORY BY ID
	category, err := h.categoryRepo.GetByID(uint(categoryID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
		return
	}

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// CHECK IF CATEGORY BELONGS TO USER
	if category.UserID == nil || *category.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to update this category")
		return
	}

	// UPDATE CATEGORY
	category.Name = req.Name
	category.Type = req.Type

	if err := h.categoryRepo.Update(category); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category updated successfully", category)
}

// DELETE CATEGORY BY ID
// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category for the authenticated user
// @Tags categories
// @Accept  json
// @Produce  json
// @Param id path int true "Category ID"
// @Success 200 {object} utils.Response[models.DeleteCategoryResponse]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
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

	// GET CATEGORY ID FROM PATH
	categoryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// GET CATEGORY BY ID
	category, err := h.categoryRepo.GetByID(uint(categoryID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
		return
	}

	// CHECK IF CATEGORY BELONGS TO USER
	if category.UserID == nil || *category.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this category")
		return
	}

	// DELETE CATEGORY
	if err := h.categoryRepo.Delete(uint(categoryID)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category deleted successfully", category)
}
