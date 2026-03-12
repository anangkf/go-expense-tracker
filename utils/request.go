package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page             int
	Limit            int
	Filters          map[string]string
	SortBy           string
	Order            string
	StartDate        string
	EndDate          string
	ExpenseStartDate string
	ExpenseEndDate   string
}

func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("user ID is of invalid type")
	}

	return id, nil
}
