package handlers

import (
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"
	"go-expense-tracker-api/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userRepo         *repositories.UserRepository
	categoryRepo     *repositories.CategoryRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	jwtService       *services.JWTService
	validator        *validator.Validate
}

func NewAuthHandler(userRepo *repositories.UserRepository, categoryRepo *repositories.CategoryRepository, refreshTokenRepo *repositories.RefreshTokenRepository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		categoryRepo:     categoryRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		validator:        validator.New(),
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

	if err := h.userRepo.Create(user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	jti := uuid.New().String()

	// GENERATE TOKEN
	token, err := h.jwtService.GenerateToken(user.ID, user.Email, jti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// GENERATE REFRESH TOKEN
	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email, jti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// SAVE REFRESH TOKEN TO DATABASE
	rt := &models.RefreshToken{
		UserID:    user.ID,
		JTI:       jti,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(h.jwtService.Config.JWT.RefreshExpireHours) * time.Hour),
	}

	if err := h.refreshTokenRepo.Create(rt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	response := gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		"token":         token,
		"refresh_token": refreshToken,
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

	// GENERATE JWT ID
	jti := uuid.New().String()

	// GENERATE ACCESS TOKEN
	token, err := h.jwtService.GenerateToken(user.ID, user.Email, jti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// GENERATE REFRESH TOKEN
	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email, jti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}
	log.Println(h.jwtService.Config.JWT.RefreshExpireHours)
	// SAVE REFRESH TOKEN TO DATABASE
	rt := &models.RefreshToken{
		UserID:    user.ID,
		JTI:       jti,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(h.jwtService.Config.JWT.RefreshExpireHours) * time.Hour),
	}

	if err := h.refreshTokenRepo.Create(rt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	response := gin.H{
		"token":         token,
		"refresh_token": refreshToken,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// REFRESH TOKEN
// RefreshToken godoc
// @Summary Refresh access token using refresh token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body models.RefreshTokenRequest true "Refresh token request data"
// @Success 200 {object} utils.Response[models.RefreshTokenResponse]
// @Failure 400 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// VALIDATE REFRESH TOKEN
	claims, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// CHECK IF REFRESH TOKEN EXISTS IN DATABASE AND IS NOT REVOKED
	storedToken, err := h.refreshTokenRepo.GetByJTI(claims.ID)
	if err != nil || storedToken.Token != req.RefreshToken {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Refresh token not found, has been revoked, or already used")
		return
	}

	// REVOKE OLD REFRESH TOKEN
	if _, err := h.refreshTokenRepo.RevokeByJTI(storedToken.JTI); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to revoke old refresh token")
		return
	}

	// GET USER
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not found")
		return
	}

	// GENERATE NEW JTI FOR ROTATION
	newJti := uuid.New().String()

	// GENERATE NEW ACCESS TOKEN
	newToken, err := h.jwtService.GenerateToken(user.ID, user.Email, newJti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate new access token")
		return
	}

	// GENERATE NEW REFRESH TOKEN
	newRefreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email, newJti)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate new refresh token")
		return
	}

	// SAVE NEW REFRESH TOKEN TO DATABASE
	rt := &models.RefreshToken{
		UserID:    user.ID,
		JTI:       newJti,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Duration(h.jwtService.Config.JWT.RefreshExpireHours) * time.Hour),
	}

	if err := h.refreshTokenRepo.Create(rt); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save new refresh token")
		return
	}

	response := gin.H{
		"token":         newToken,
		"refresh_token": newRefreshToken,
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", response)
}

// LOGOUT
// Logout godoc
// @Summary Logout user
// @Description Logout user by revoking refresh token
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// GET JTI FROM CONTEXT (SET BY AUTH MIDDLEWARE)
	jti, exists := c.Get("jti")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Token ID not found in token")
		return
	}

	// REVOKE REFRESH TOKEN BY JTI
	revokedCount, err := h.refreshTokenRepo.RevokeByJTI(jti.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to logout")
		return
	}

	if revokedCount == 0 {
		utils.SuccessResponse(c, http.StatusOK, "No active sessions found or already logged out", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
