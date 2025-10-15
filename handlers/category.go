package handlers

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"

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

	response := gin.H{
		"category": categories[0],
	}

	utils.SuccessResponse(c, http.StatusCreated, "Category created successfully", response)
}
