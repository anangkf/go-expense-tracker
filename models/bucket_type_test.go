package models

import "testing"

func TestIsValidBucketTypeID(t *testing.T) {
	tests := []struct {
		id   uint
		want bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{100, false},
	}

	for _, tt := range tests {
		if got := IsValidBucketTypeID(tt.id); got != tt.want {
			t.Errorf("IsValidBucketTypeID(%d) = %v, want %v", tt.id, got, tt.want)
		}
	}
}
