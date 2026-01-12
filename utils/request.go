package utils

type QueryParams struct {
	Page    int
	Limit   int
	Filters map[string]string
	SortBy  string
	Order   string
}
