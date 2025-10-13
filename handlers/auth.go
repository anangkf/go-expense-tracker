package handlers

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"
	"go-expense-tracker-api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	userRepo   *repositories.UserRepository
	jwtService *services.JWTService
	validator  *validator.Validate
}

func NewAuthHandler(userRepo *repositories.UserRepository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
		validator:  validator.New(),
	}
}

// REGISTER NEW USER
// Register godoc
// @Summary Register users
// @Description Create a new user account
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body models.RegisterRequest true "User registration data"
// @Success 201 {object} utils.Response[models.RegisterResponse]
// @Failure 400 {object} utils.Response[any]
// @Failure 409 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// CHECK IF EMAIL ALREADY EXISTS
	existingUser, _ := h.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email already exists")
		return
	}

	// HASH PASSWORD
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// CREATE NEW USER
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := h.userRepo.Create(*user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// GENERATE TOKEN
	token, err := h.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		"token": token,
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", response)
}

// LOGIN
// Login godoc
// @Summary Login users
// @Description Authenticate user
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body models.LoginRequest true "User authentication data"
// @Success 200 {object} utils.Response[models.LoginResponse]
// @Failure 400 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/login [post]// Login godoc
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// GET USER BY EMAIL
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// PASSWORD VERIFICATION
	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// GENERATE TOKEN
	token, err := h.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := gin.H{
		"token": token,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}
