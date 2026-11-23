package utils

import (
	"testing"
	"time"
)

func TestStartOfDay(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)
	in := time.Date(2024, 3, 15, 23, 59, 59, 999, loc)
	want := time.Date(2024, 3, 15, 0, 0, 0, 0, loc)
	got := StartOfDay(in)
	if !got.Equal(want) || got.Location() != loc {
		t.Errorf("got %v, want %v", got, want)
	}
	if !StartOfDay(got).Equal(got) {
		t.Errorf("StartOfDay not idempotent")
	}
}

func TestEndOfDayExclusive(t *testing.T) {
	loc := time.UTC
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"midnight", time.Date(2024, 3, 15, 0, 0, 0, 0, loc), time.Date(2024, 3, 16, 0, 0, 0, 0, loc)},
		{"last nanosecond", time.Date(2024, 3, 15, 23, 59, 59, 999999999, loc), time.Date(2024, 3, 16, 0, 0, 0, 0, loc)},
		{"month end", time.Date(2024, 1, 31, 12, 0, 0, 0, loc), time.Date(2024, 2, 1, 0, 0, 0, 0, loc)},
		{"leap day", time.Date(2024, 2, 28, 10, 0, 0, 0, loc), time.Date(2024, 2, 29, 0, 0, 0, 0, loc)},
		{"year end", time.Date(2024, 12, 31, 8, 0, 0, 0, loc), time.Date(2025, 1, 1, 0, 0, 0, 0, loc)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndOfDayExclusive(tt.in); !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
