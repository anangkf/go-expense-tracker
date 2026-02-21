package middleware

import (
	"go-expense-tracker-api/utils"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func PaginationAndFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// DEFAULT PAGINATION
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		// DEFAULT SORTING
		sortBy := c.DefaultQuery("sortBy", "id")
		order := strings.ToLower(c.DefaultQuery("order", "asc"))
		if order != "asc" && order != "desc" {
			order = "asc"
		}

		// FILTERING
		filters := make(map[string]string)
		filterKeysExclusion := []string{"page", "limit", "sortBy", "order", "withTotal", "start_date", "end_date", "expense_start_date", "expense_end_date"}
		for key, val := range c.Request.URL.Query() {
			if !slices.Contains(filterKeysExclusion, key) {
				if len(val) > 0 {
					filters[key] = val[0]
				}
			}
		}
		// filters["start_date"] = startDateStr
		// filters["end_date"] = endDateStr

		queryParams := utils.QueryParams{
			Page:             page,
			Limit:            limit,
			Filters:          filters,
			SortBy:           sortBy,
			Order:            order,
			StartDate:        c.Query("start_date"),
			EndDate:          c.Query("end_date"),
			ExpenseStartDate: c.Query("expense_start_date"),
			ExpenseEndDate:   c.Query("expense_end_date"),
		}

		c.Set("queryParams", queryParams)

		c.Next()
	}
}
