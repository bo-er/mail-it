package util

import "time"

// GetFirstDayOfWeek returns the first day of this week
func GetFirstDayOfWeek() time.Time {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return weekStartDate
}

// GetFirstDayOfLastWeek returns the first day of last week
func GetFirstDayOfLastWeek() time.Time {
	firstDayOfWeek := GetFirstDayOfWeek()
	return firstDayOfWeek.Add(time.Duration(-7*24) * time.Hour)
}

// GetSaturdayOfLastWeek returns the last day of last week
func GetSaturdayOfLastWeek() time.Time {
	firstDayOfWeek := GetFirstDayOfWeek()
	return firstDayOfWeek.Add(time.Duration(-2*24) * time.Hour)
}
