package cron

import (
	"testing"
	"time"
)

type mockSchedule struct {
	minute       int32
	hour         int32
	dayOfMonth   int32
	dayOfWeek    int32
	month        int32
	everyNMonths int32
}

func (ms mockSchedule) GetMinute() int32       { return ms.minute }
func (ms mockSchedule) GetDayOfWeek() int32    { return ms.dayOfWeek }
func (ms mockSchedule) GetDayOfMonth() int32   { return ms.dayOfMonth }
func (ms mockSchedule) GetMonth() int32        { return ms.month }
func (ms mockSchedule) GetHour() int32         { return ms.hour }
func (ms mockSchedule) GetEveryNMonths() int32 { return ms.everyNMonths }

func TestGetNextExecution(t *testing.T) {
	tests := []struct {
		name     string
		timeNow  func() time.Time
		schedule Schedule
		expected time.Time
	}{
		// minute
		{
			name: "Only minute provided, current minute behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 00, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{minute: 10, hour: -1, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 01, 00, 10, 0, 0, time.UTC),
		},
		{
			name: "Only minute provided, current minute ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 00, 11, 0, 0, time.UTC)
			},
			schedule: mockSchedule{minute: 10, hour: -1, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 01, 01, 10, 0, 0, time.UTC),
		},
		// hour
		{
			name: "Only hour provided, current hour behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 04, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{hour: 05, minute: -1, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 01, 05, 00, 0, 0, time.UTC),
		},
		{
			name: "Only hour provided, current hour ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 06, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{hour: 05, minute: -1, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 02, 05, 00, 0, 0, time.UTC),
		},
		{
			name: "Hour and minute provided, hour behind current minute ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 04, 31, 0, 0, time.UTC)
			},
			schedule: mockSchedule{hour: 05, minute: 30, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 01, 05, 30, 0, 0, time.UTC),
		},
		{
			name: "Hour and minute provided, current minute behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 05, 20, 0, 0, time.UTC)
			},
			schedule: mockSchedule{hour: 05, minute: 30, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 01, 05, 30, 0, 0, time.UTC),
		},
		{
			name: "Hour and minute provided, current minute ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 05, 40, 0, 0, time.UTC)
			},
			schedule: mockSchedule{hour: 05, minute: 30, dayOfMonth: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 02, 05, 30, 0, 0, time.UTC),
		},
		// day of the week
		{
			name: "Day of week provided, current day behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 05, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: -1, dayOfMonth: -1, month: -1},
			expected: time.Date(2026, time.February, 04, 00, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of week provided, current day ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 05, 05, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: -1, month: -1},
			expected: time.Date(2026, time.February, 11, 00, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of week and hour provided, hour behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 04, 04, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: 5},
			expected: time.Date(2026, time.February, 04, 05, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of week and hour provided, hour ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 04, 06, 30, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: 5},
			expected: time.Date(2026, time.February, 11, 05, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of week,hour,minute provided, current minute behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 04, 05, 20, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: 05, minute: 30},
			expected: time.Date(2026, time.February, 04, 05, 30, 0, 0, time.UTC),
		},
		{
			name: "Day of week,hour,minute provided, current minute ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 04, 05, 40, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfWeek: 3, hour: 05, minute: 30},
			expected: time.Date(2026, time.February, 11, 05, 30, 0, 0, time.UTC),
		},
		// day of the month
		{
			name: "Day of month provided, current day behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 01, 05, 40, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 10, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 10, 00, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of month provided, current day ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 11, 05, 40, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 10, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.March, 10, 00, 00, 0, 0, time.UTC),
		},
		{
			name: "Day of month provided, current hour ahead (incongruous month days)",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 31, 05, 40, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 31, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 28, 00, 00, 0, 0, time.UTC),
		},
		// month
		{
			name: "Specific month provided, current month behind",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 21, 15, 25, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.May), hour: -1, minute: -1, dayOfMonth: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Specific month provided, current month ahead",
			timeNow: func() time.Time {
				return time.Date(2026, time.June, 21, 15, 25, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.May), hour: -1, minute: -1, dayOfMonth: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.May, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Specific month provided, same month ahead by days",
			timeNow: func() time.Time {
				return time.Date(2026, time.May, 21, 15, 25, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.May), hour: -1, minute: -1, dayOfMonth: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.May, 1, 0, 0, 0, 0, time.UTC),
		},
		// everyNMonths - quarterly
		{
			name: "Quarterly schedule, current date before anchor",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Quarterly schedule, current date after anchor",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 10, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Quarterly schedule, anchor day 31 to month with 30 days",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.January), dayOfMonth: 31, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		// everyNMonths - semi-annual
		{
			name: "Semi-annual schedule, current date before anchor",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Semi-annual schedule, current date after anchor",
			timeNow: func() time.Time {
				return time.Date(2026, time.February, 10, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Semi-annual schedule, anchor day 31 to February (non-leap)",
			timeNow: func() time.Time {
				return time.Date(2026, time.August, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.August), dayOfMonth: 31, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		// End-of-month edge cases - Monthly
		{
			name: "Monthly day 29, Jan to Feb non-leap (28 days)",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 29, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 29, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly day 30, Jan to Feb non-leap (28 days)",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 30, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 30, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly day 31, Jan to Feb non-leap (28 days)",
			timeNow: func() time.Time {
				return time.Date(2026, time.January, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 31, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly day 31, Mar to Apr (30 days)",
			timeNow: func() time.Time {
				return time.Date(2026, time.March, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{dayOfMonth: 31, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly day 30, Jan to Feb leap year (29 days)",
			timeNow: func() time.Time {
				return time.Date(2028, time.January, 30, 10, 0, 0, 0, time.UTC) // 2028 is leap year
			},
			schedule: mockSchedule{dayOfMonth: 30, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly day 29, Jan to Feb leap year (29 days)",
			timeNow: func() time.Time {
				return time.Date(2028, time.January, 29, 10, 0, 0, 0, time.UTC) // 2028 is leap year
			},
			schedule: mockSchedule{dayOfMonth: 29, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC),
		},
		// End-of-month edge cases - Quarterly
		{
			name: "Quarterly day 29, Nov to Feb non-leap",
			timeNow: func() time.Time {
				return time.Date(2026, time.November, 29, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.November), dayOfMonth: 29, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Quarterly day 30, Nov to Feb non-leap",
			timeNow: func() time.Time {
				return time.Date(2026, time.November, 30, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.November), dayOfMonth: 30, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Quarterly day 31, Oct to Jan (31 days, no clamping)",
			timeNow: func() time.Time {
				return time.Date(2026, time.October, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.October), dayOfMonth: 31, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.January, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Quarterly day 31, May to Aug (31 days, no clamping)",
			timeNow: func() time.Time {
				return time.Date(2026, time.May, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.May), dayOfMonth: 31, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
		},
		// End-of-month edge cases - Semi-annual
		{
			name: "Semi-annual day 29, Aug to Feb non-leap",
			timeNow: func() time.Time {
				return time.Date(2026, time.August, 29, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.August), dayOfMonth: 29, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Semi-annual day 30, Aug to Feb non-leap",
			timeNow: func() time.Time {
				return time.Date(2026, time.August, 30, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.August), dayOfMonth: 30, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Semi-annual day 30, Aug to Feb leap year",
			timeNow: func() time.Time {
				return time.Date(2027, time.August, 30, 10, 0, 0, 0, time.UTC) // 2028 is leap year
			},
			schedule: mockSchedule{month: int32(time.August), dayOfMonth: 30, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Semi-annual day 31, Jul to Jan (31 days, no clamping)",
			timeNow: func() time.Time {
				return time.Date(2026, time.July, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.July), dayOfMonth: 31, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.January, 31, 0, 0, 0, 0, time.UTC),
		},
		// End-of-month edge cases - Annual
		{
			name: "Annual Feb 29 to non-leap year (clamp to Feb 28)",
			timeNow: func() time.Time {
				return time.Date(2028, time.February, 29, 10, 0, 0, 0, time.UTC) // 2028 is leap year
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 29, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2029, time.February, 28, 0, 0, 0, 0, time.UTC), // 2029 is not leap year
		},
		{
			name: "Annual Feb 29 leap year to non-leap year (clamp to Feb 28)",
			timeNow: func() time.Time {
				return time.Date(2024, time.February, 29, 10, 0, 0, 0, time.UTC) // 2024 is leap year
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 29, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC), // 2025 is not leap year, clamp to 28
		},
		{
			name: "Annual Feb 28 non-leap to leap year (stays Feb 28)",
			timeNow: func() time.Time {
				return time.Date(2027, time.February, 28, 10, 0, 0, 0, time.UTC) // 2027 is not leap year
			},
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 28, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2028, time.February, 28, 0, 0, 0, 0, time.UTC), // stays 28, doesn't become 29
		},
		{
			name: "Annual Mar 31 (always 31 days, no clamping)",
			timeNow: func() time.Time {
				return time.Date(2026, time.March, 31, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.March), dayOfMonth: 31, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Annual Apr 30 (always 30 days, no clamping)",
			timeNow: func() time.Time {
				return time.Date(2026, time.April, 30, 10, 0, 0, 0, time.UTC)
			},
			schedule: mockSchedule{month: int32(time.April), dayOfMonth: 30, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := scheduler{
				timeNow: tt.timeNow,
			}
			got := sch.GetNextExecution(tt.schedule)
			if got != tt.expected {
				t.Errorf("GetNextExecution() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestGetNextExecutionFrom(t *testing.T) {
	tests := []struct {
		name     string
		from     time.Time
		schedule Schedule
		expected time.Time
	}{
		{
			name:     "Monthly from specific date",
			from:     time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{dayOfMonth: 20, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Monthly from specific date, day already passed",
			from:     time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{dayOfMonth: 20, hour: -1, minute: -1, dayOfWeek: -1, month: -1},
			expected: time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Quarterly from current recurrence date",
			from:     time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Semi-annual from current recurrence date",
			from:     time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 6, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Annual from current recurrence date",
			from:     time.Date(2026, time.February, 5, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{month: int32(time.February), dayOfMonth: 5, everyNMonths: 0, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2027, time.February, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Weekly from specific date",
			from:     time.Date(2026, time.February, 4, 0, 0, 0, 0, time.UTC),                     // Wednesday
			schedule: mockSchedule{dayOfWeek: 3, hour: -1, minute: -1, dayOfMonth: -1, month: -1}, // Wednesday
			expected: time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Quarterly with end-of-month handling",
			from:     time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			schedule: mockSchedule{month: int32(time.January), dayOfMonth: 31, everyNMonths: 3, hour: -1, minute: -1, dayOfWeek: -1},
			expected: time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	sch := NewScheduler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sch.GetNextExecutionFrom(tt.schedule, tt.from)
			if got != tt.expected {
				t.Errorf("GetNextExecutionFrom() = %s, want %s", got, tt.expected)
			}
		})
	}
}
