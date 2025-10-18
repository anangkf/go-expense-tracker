package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page    int
	Limit   int
	Filters map[string]string
	SortBy  string
	Order   string
}

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
		for key, val := range c.Request.URL.Query() {
			if key != "page" && key != "limit" && key != "sortBy" && key != "order" {
				filters[key] = val[0]
			}
		}

		queryParams := QueryParams{
			Page:    page,
			Limit:   limit,
			Filters: filters,
			SortBy:  sortBy,
			Order:   order,
		}

		c.Set("queryParams", queryParams)

		c.Next()
	}
}
