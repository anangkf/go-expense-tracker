package utils

import "testing"

var testAllowed = map[string]string{
	"id":         "expenses.id",
	"amount":     "expenses.amount",
	"created_at": "expenses.created_at",
}

func TestSafeSort(t *testing.T) {
	const fallback = "expenses.created_at"
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"allowed key", "amount", "expenses.amount"},
		{"empty falls back", "", fallback},
		{"unknown falls back", "password", fallback},
		{"injection semicolon", "id;DROP TABLE users--", fallback},
		{"injection subquery", "(SELECT 1)", fallback},
		{"raw sql column not accepted", "expenses.amount", fallback},
		{"case sensitive", "ID", fallback},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeSort(tt.input, testAllowed, fallback); got != tt.expect {
				t.Errorf("SafeSort(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestSafeSortNilAllowlist(t *testing.T) {
	if got := SafeSort("id", nil, "fb"); got != "fb" {
		t.Errorf("got %q, want fb", got)
	}
}

func TestSafeFilters(t *testing.T) {
	input := map[string]string{
		"id":                       "5",
		"amount":                   "100",
		"unknown":                  "x",
		"id; DROP TABLE users--":   "1",
		"amount ILIKE '' OR 1=1--": "z",
	}
	got := SafeFilters(input, testAllowed)

	if len(got) != 2 {
		t.Fatalf("expected 2 filters, got %d: %v", len(got), got)
	}
	if got["expenses.id"] != "5" || got["expenses.amount"] != "100" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestSafeFiltersEmpty(t *testing.T) {
	if got := SafeFilters(nil, testAllowed); got == nil || len(got) != 0 {
		t.Errorf("nil input: got %v, want empty non-nil map", got)
	}
	if got := SafeFilters(map[string]string{"id": "1"}, nil); len(got) != 0 {
		t.Errorf("nil allowlist: got %v, want empty", got)
	}
}
