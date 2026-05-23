package services

// AllowedReminderMinutes are valid user/task reminder offsets (minutes before due_date).
var AllowedReminderMinutes = []int{5, 10, 15, 30, 60}

// IsValidReminderMinutes reports whether m is an allowed reminder offset.
func IsValidReminderMinutes(m int) bool {
	switch m {
	case 5, 10, 15, 30, 60:
		return true
	default:
		return false
	}
}
