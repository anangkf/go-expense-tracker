package services

import (
	"errors"
	"go-expense-tracker-api/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Config *config.Config
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		Config: cfg,
	}
}

// GENERATE JWT TOKEN
func (j *JWTService) GenerateToken(userID uint, email string, jti string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(j.Config.JWT.ExpireHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.Config.JWT.Secret))

	if err != nil {
		return "", nil
	}

	return tokenString, nil
}

// GENERATE REFRESH TOKEN
func (j *JWTService) GenerateRefreshToken(userID uint, email string, jti string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(j.Config.JWT.RefreshExpireHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.Config.JWT.RefreshSecret))

	if err != nil {
		return "", nil
	}

	return tokenString, nil
}

// VALIDATE TOKEN
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.Config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}

func (j *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.Config.JWT.RefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("Invalid refresh token")
	}

	return claims, nil
}
