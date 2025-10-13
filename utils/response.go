package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

func SuccessResponse(c *gin.Context, stausCode int, message string, data interface{}) {
	c.JSON(stausCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, stausCode int, message string) {
	c.JSON(stausCode, Response{
		Success: false,
		Message: "Error",
		Error:   message,
	})
}

func ValidateErrorResponse(c *gin.Context, statusCode int, errors []string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: "Validation Error",
		Data:    gin.H{"errors": errors},
	})
}
