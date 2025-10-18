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

type PaginationResponse[T any] struct {
	Data       T     `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type ResponseWithPagination[T any] struct {
	Response[PaginationResponse[T]]
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, Response[any]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response[any]{
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
