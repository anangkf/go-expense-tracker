package services

import (
	"testing"
	"time"

	"go-expense-tracker-api/config"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret        = "test-access-secret"
	testRefreshSecret = "test-refresh-secret"
)

func newTestService(secret, refreshSecret string) *JWTService {
	return NewJWTService(&config.Config{
		JWT: config.JWTConfig{
			Secret:             secret,
			ExpireHours:        1,
			RefreshSecret:      refreshSecret,
			RefreshExpireHours: 2,
		},
	})
}

func testClaims() *Claims {
	return &Claims{
		UserID: 42,
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "jti-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func signWith(t *testing.T, method jwt.SigningMethod, key interface{}) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, testClaims()).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return s
}

func TestValidTokensPassValidation(t *testing.T) {
	svc := newTestService(testSecret, testRefreshSecret)

	access, err := svc.GenerateToken(42, "user@example.com", "jti-1")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	claims, err := svc.ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != 42 || claims.Email != "user@example.com" || claims.ID != "jti-1" {
		t.Errorf("unexpected access claims: %+v", claims)
	}

	refresh, err := svc.GenerateRefreshToken(42, "user@example.com", "jti-1")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	claims, err = svc.ValidateRefreshToken(refresh)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if claims.UserID != 42 || claims.Email != "user@example.com" || claims.ID != "jti-1" {
		t.Errorf("unexpected refresh claims: %+v", claims)
	}
}

func TestRejectsDisallowedAlgorithms(t *testing.T) {
	svc := newTestService(testSecret, testRefreshSecret)

	tests := []struct {
		name  string
		token func(t *testing.T, secret string) string
	}{
		{"alg=none", func(t *testing.T, _ string) string {
			return signWith(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType)
		}},
		{"alg=HS512", func(t *testing.T, secret string) string {
			return signWith(t, jwt.SigningMethodHS512, []byte(secret))
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name+"/ValidateToken", func(t *testing.T) {
			if _, err := svc.ValidateToken(tc.token(t, testSecret)); err == nil {
				t.Error("expected error, got nil")
			}
		})
		t.Run(tc.name+"/ValidateRefreshToken", func(t *testing.T) {
			if _, err := svc.ValidateRefreshToken(tc.token(t, testRefreshSecret)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestRejectsWrongSecret(t *testing.T) {
	svc := newTestService(testSecret, testRefreshSecret)

	wrongAccess := signWith(t, jwt.SigningMethodHS256, []byte("wrong-secret"))
	if _, err := svc.ValidateToken(wrongAccess); err == nil {
		t.Error("ValidateToken: expected error for wrong secret")
	}

	wrongRefresh := signWith(t, jwt.SigningMethodHS256, []byte("wrong-secret"))
	if _, err := svc.ValidateRefreshToken(wrongRefresh); err == nil {
		t.Error("ValidateRefreshToken: expected error for wrong secret")
	}

	// Access and refresh secrets must not be interchangeable.
	if _, err := svc.ValidateToken(signWith(t, jwt.SigningMethodHS256, []byte(testRefreshSecret))); err == nil {
		t.Error("ValidateToken: accepted token signed with refresh secret")
	}
	if _, err := svc.ValidateRefreshToken(signWith(t, jwt.SigningMethodHS256, []byte(testSecret))); err == nil {
		t.Error("ValidateRefreshToken: accepted token signed with access secret")
	}
}
