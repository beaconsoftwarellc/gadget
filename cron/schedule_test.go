package cron

import (
	"testing"
	"time"
)

type mockSchedule struct {
	minute     int32
	hour       int32
	dayOfMonth int32
	dayOfWeek  int32
	month      int32
}

func (ms mockSchedule) GetMinute() int32     { return ms.minute }
func (ms mockSchedule) GetDayOfWeek() int32  { return ms.dayOfWeek }
func (ms mockSchedule) GetDayOfMonth() int32 { return ms.dayOfMonth }
func (ms mockSchedule) GetMonth() int32      { return ms.month }
func (ms mockSchedule) GetHour() int32       { return ms.hour }

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
