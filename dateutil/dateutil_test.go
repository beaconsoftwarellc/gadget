package dateutil

import (
	"testing"
	"time"
)

// ... existing TestIncrementMonth ...

func TestIncrementYear(t *testing.T) {
	tests := []struct {
		name     string
		today    time.Time
		years    int
		expected time.Time
	}{
		{
			name:     "Add one year",
			today:    time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC),
			years:    1,
			expected: time.Date(2024, 5, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add zero years",
			today:    time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC),
			years:    0,
			expected: time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Subtract one year",
			today:    time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC),
			years:    -1,
			expected: time.Date(2022, 5, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add one year crossing a leap year",
			today:    time.Date(2023, 2, 28, 12, 0, 0, 0, time.UTC),
			years:    1,
			expected: time.Date(2024, 2, 28, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Subtract one year crossing a leap year",
			today:    time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			years:    -1,
			expected: time.Date(2023, 2, 28, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add years maintaining end-of-month for leap years",
			today:    time.Date(2020, 2, 29, 12, 0, 0, 0, time.UTC),
			years:    4,
			expected: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add years crossing year boundaries",
			today:    time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC),
			years:    1,
			expected: time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Subtract years crossing year boundaries",
			today:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			years:    -1,
			expected: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Handle non-existent date backward (from leap year February)",
			today:    time.Date(2020, 2, 29, 12, 0, 0, 0, time.UTC),
			years:    -1,
			expected: time.Date(2019, 2, 28, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncrementYear(tt.today, tt.years)
			if !result.Equal(tt.expected) {
				t.Errorf("IncrementYear(%v, %d) = %v, want %v", tt.today, tt.years, result, tt.expected)
			}
		})
	}
}

func TestIncrementMonth(t *testing.T) {
	tests := []struct {
		name     string
		today    time.Time
		months   int
		expected time.Time
	}{
		{
			name:     "Add one month to a mid-month date",
			today:    time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC),
			months:   1,
			expected: time.Date(2023, 2, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add one month to end-of-month date",
			today:    time.Date(2023, 1, 31, 12, 0, 0, 0, time.UTC),
			months:   1,
			expected: time.Date(2023, 2, 28, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add two months crossing a non-leap February from January",
			today:    time.Date(2023, 1, 31, 12, 0, 0, 0, time.UTC),
			months:   2,
			expected: time.Date(2023, 3, 31, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add months that keep dates aligned on leap year",
			today:    time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
			months:   2,
			expected: time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add negative months stepping backward",
			today:    time.Date(2023, 3, 15, 12, 0, 0, 0, time.UTC),
			months:   -1,
			expected: time.Date(2023, 2, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Backward crossing leap year to February",
			today:    time.Date(2024, 3, 31, 12, 0, 0, 0, time.UTC),
			months:   -1,
			expected: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add zero months returns same date",
			today:    time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC),
			months:   0,
			expected: time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Add months over year boundary",
			today:    time.Date(2023, 11, 30, 12, 0, 0, 0, time.UTC),
			months:   2,
			expected: time.Date(2024, 1, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Backward months over year boundary",
			today:    time.Date(2023, 1, 30, 12, 0, 0, 0, time.UTC),
			months:   -2,
			expected: time.Date(2022, 11, 30, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IncrementMonth(tt.today, tt.months)
			if !result.Equal(tt.expected) {
				t.Errorf("IncrementMonth(%v, %d) = %v, want %v", tt.today, tt.months, result, tt.expected)
			}
		})
	}
}
