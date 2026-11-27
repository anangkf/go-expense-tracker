package handlers

import (
	"errors"
	"go-expense-tracker-api/config"
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"
	"go-expense-tracker-api/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userRepo         *repositories.UserRepository
	categoryRepo     *repositories.CategoryRepository
	refreshTokenRepo *repositories.RefreshTokenRepository
	emailVerifRepo   *repositories.EmailVerificationRepository
	jwtService       *services.JWTService
	emailService     *services.EmailService
	emailCfg         config.EmailConfig
	baseURLFE        string
	otpSecret        string
	validator        *validator.Validate
}

func NewAuthHandler(userRepo *repositories.UserRepository, categoryRepo *repositories.CategoryRepository, refreshTokenRepo *repositories.RefreshTokenRepository, emailVerifRepo *repositories.EmailVerificationRepository, jwtService *services.JWTService, emailService *services.EmailService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		categoryRepo:     categoryRepo,
		refreshTokenRepo: refreshTokenRepo,
		emailVerifRepo:   emailVerifRepo,
		jwtService:       jwtService,
		emailService:     emailService,
		emailCfg:         cfg.Email,
		baseURLFE:        cfg.Server.BaseURLFE,
		otpSecret:        cfg.JWT.Secret,
		validator:        validator.New(),
	}
}

// PREFIX AN ENDPOINT WITH THE FRONTEND BASE URL WHEN BASE_URL_FE IS SET, OTHERWISE RETURN IT AS IS
func (h *AuthHandler) feURL(endpoint string) string {
	return h.baseURLFE + endpoint
}

// CREATE AND STORE A NEW OTP, THEN EMAIL IT IN THE BACKGROUND (A FAILED SEND CAN BE RETRIED VIA RESEND)
func (h *AuthHandler) issueAndSendOTP(user *models.User) error {
	code, err := services.GenerateOTP()
	if err != nil {
		return err
	}
	err = h.emailVerifRepo.Issue(
		user.ID,
		services.HashOTP(h.otpSecret, user.ID, code),
		time.Now(),
		time.Duration(h.emailCfg.OTPExpireMinutes)*time.Minute,
		time.Duration(h.emailCfg.ResendCooldownSeconds)*time.Second,
		h.emailCfg.ResendMaxPerDay,
	)
	if err != nil {
		return err
	}
	email, name := user.Email, user.Name
	go func() {
		if err := h.emailService.SendVerificationOTP(email, name, code); err != nil {
			log.Printf("Failed to send verification email to user %d: %v", user.ID, err)
		}
	}()
	return nil
}

// REGISTER NEW USER
// Register godoc
// @Summary Register users
// @Description Create a new unverified user account and email a 6 digit verification code. Tokens are issued by login after verification.
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

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// CHECK IF EMAIL ALREADY EXISTS (NEVER OVERWRITE AN EXISTING ACCOUNT)
	existingUser, _ := h.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		if existingUser.EmailVerifiedAt == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Email is registered but not verified. Request a new code via "+h.feURL("/auth/resend-verification"))
			return
		}
		utils.ErrorResponse(c, http.StatusConflict, "Email already exists")
		return
	}

	// HASH PASSWORD
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// CREATE NEW USER (UNVERIFIED)
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}
	if err := h.userRepo.Create(user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// SEND VERIFICATION OTP
	if err := h.issueAndSendOTP(user); err != nil {
		log.Printf("Failed to issue verification code for user %d: %v", user.ID, err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Account created but failed to send verification code. Request a new code via "+h.feURL("/auth/resend-verification"))
		return
	}

	response := models.RegisterResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	}
	utils.SuccessResponse(c, http.StatusCreated, "User registered. Check your email for the verification code", response)
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
// @Failure 401 {object} utils.Response[any]
// @Failure 403 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/login [post]
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
	user, err := h.userRepo.GetByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// PASSWORD VERIFICATION
	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// REJECT UNVERIFIED EMAIL (CHECKED AFTER PASSWORD SO IT DOES NOT LEAK TO STRANGERS)
	if user.EmailVerifiedAt == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Email not verified. Verify with the code we sent, or request a new one via "+h.feURL("/auth/resend-verification"))
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

// VERIFY EMAIL
// VerifyEmail godoc
// @Summary Verify email with OTP
// @Description Verify the 6 digit code sent by email. The code expires and is invalidated after too many wrong attempts.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body models.VerifyEmailRequest true "Email and OTP"
// @Success 200 {object} utils.Response[any]
// @Failure 400 {object} utils.Response[any]
// @Failure 429 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// UNKNOWN AND ALREADY VERIFIED EMAILS GET THE SAME ANSWER AS A WRONG CODE
	const invalidMsg = "Invalid or expired verification code"
	user, err := h.userRepo.GetByEmail(strings.TrimSpace(req.Email))
	if err != nil || user.EmailVerifiedAt != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, invalidMsg)
		return
	}

	err = h.emailVerifRepo.Verify(
		user.ID,
		services.HashOTP(h.otpSecret, user.ID, req.OTP),
		time.Now(),
		h.emailCfg.OTPMaxAttempts,
	)
	switch {
	case err == nil:
		utils.SuccessResponse(c, http.StatusOK, "Email verified successfully", nil)
	case errors.Is(err, repositories.ErrOTPTooManyTries):
		utils.ErrorResponse(c, http.StatusTooManyRequests, "Too many wrong attempts. Request a new code via "+h.feURL("/auth/resend-verification"))
	case errors.Is(err, repositories.ErrOTPInvalid), errors.Is(err, repositories.ErrOTPExpired), errors.Is(err, repositories.ErrOTPNotFound):
		utils.ErrorResponse(c, http.StatusBadRequest, invalidMsg)
	default:
		log.Printf("Failed to verify email for user %d: %v", user.ID, err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify email")
	}
}

// RESEND VERIFICATION
// ResendVerification godoc
// @Summary Resend verification OTP
// @Description Send a new verification code to an unverified account. Limited by a cooldown and a daily maximum. Unknown or already verified emails get the same generic success response.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body models.ResendVerificationRequest true "Email"
// @Success 200 {object} utils.Response[any]
// @Failure 400 {object} utils.Response[any]
// @Failure 429 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	const genericMsg = "If the account exists and is not verified, a new code has been sent"
	user, err := h.userRepo.GetByEmail(strings.TrimSpace(req.Email))
	if err != nil || user.EmailVerifiedAt != nil {
		utils.SuccessResponse(c, http.StatusOK, genericMsg, nil)
		return
	}

	err = h.issueAndSendOTP(user)
	switch {
	case err == nil:
		utils.SuccessResponse(c, http.StatusOK, genericMsg, nil)
	case errors.Is(err, repositories.ErrResendCooldown):
		utils.ErrorResponse(c, http.StatusTooManyRequests, "Please wait before requesting another code")
	case errors.Is(err, repositories.ErrResendLimit):
		utils.ErrorResponse(c, http.StatusTooManyRequests, "Daily resend limit reached. Try again tomorrow")
	default:
		log.Printf("Failed to resend verification code for user %d: %v", user.ID, err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send verification code")
	}
}
