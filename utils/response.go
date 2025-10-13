package utils

import (
	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, stausCode int, message string, data any) {
	c.JSON(stausCode, Response[any]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, stausCode int, message string) {
	c.JSON(stausCode, Response[any]{
		Success: false,
		Message: "Error",
		Error:   message,
	})
}

func ValidateErrorResponse(c *gin.Context, statusCode int, errors []string) {
	c.JSON(statusCode, Response[map[string]any]{
		Success: false,
		Message: "Validation Error",
		Data:    gin.H{"errors": errors},
	})
}
