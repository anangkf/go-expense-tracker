package utils

// SAFE SORT
// SafeSort mengembalikan ekspresi kolom SQL dari allowlist (key = nama query,
// value = kolom SQL). Jika sortBy tidak ada di allowlist, fallback dikembalikan.
func SafeSort(sortBy string, allowed map[string]string, fallback string) string {
	if column, ok := allowed[sortBy]; ok {
		return column
	}
	return fallback
}

// SAFE FILTERS
// SafeFilters hanya mengembalikan filter yang key-nya ada di allowlist,
// dengan key diganti menjadi kolom SQL.
func SafeFilters(filters map[string]string, allowed map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range filters {
		if column, ok := allowed[key]; ok {
			result[column] = value
		}
	}
	return result
}
