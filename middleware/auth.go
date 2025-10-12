package middleware

import (
	"net/http"
	"strings"

	"go-expense-tracker-api/services"
	"go-expense-tracker-api/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization Header required")
			c.Abort()
			return
		}

		// Bearer <token>
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		claims, err := jwtService.ValidateToken(tokenString)

		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token"+err.Error())
			c.Abort()
			return
		}

		// SET USER INFO TO CONTEXT
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
