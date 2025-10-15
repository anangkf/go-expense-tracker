package handlers

import (
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
}

func NewUserHandler(userRepo *repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GET USER PROFILE
// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[models.UserResponse]
// @Failure 401 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /user/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// GET USER BY ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	response := gin.H{
		"user": user,
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile retrieved successfully", response)
}
