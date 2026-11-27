package handlers

import "testing"

func TestFeURL(t *testing.T) {
	tests := []struct {
		base, want string
	}{
		{"", "/auth/resend-verification"},
		{"https://app.example.com", "https://app.example.com/auth/resend-verification"},
	}
	for _, tt := range tests {
		h := &AuthHandler{baseURLFE: tt.base}
		if got := h.feURL("/auth/resend-verification"); got != tt.want {
			t.Fatalf("base %q: got %q, want %q", tt.base, got, tt.want)
		}
	}
}
