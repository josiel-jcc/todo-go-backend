package services

import (
	"time"
	"todo-go-backend/internal/models"
)

// ComputeNextDue returns the next due date after base according to rule.
func ComputeNextDue(base time.Time, rule models.RecurrenceRule) time.Time {
	switch rule {
	case models.RecurrenceDaily:
		return base.AddDate(0, 0, 1)
	case models.RecurrenceWeekly:
		return base.AddDate(0, 0, 7)
	case models.RecurrenceMonthly:
		return base.AddDate(0, 1, 0)
	default:
		return base
	}
}

// AdvanceToFutureDue advances from base until the next due is strictly after now.
func AdvanceToFutureDue(base time.Time, rule models.RecurrenceRule, now time.Time) time.Time {
	next := ComputeNextDue(base, rule)
	for !next.After(now) {
		next = ComputeNextDue(next, rule)
	}
	return next
}

// IsValidRecurrenceRule reports whether rule is a supported recurrence interval.
func IsValidRecurrenceRule(rule models.RecurrenceRule) bool {
	switch rule {
	case models.RecurrenceDaily, models.RecurrenceWeekly, models.RecurrenceMonthly:
		return true
	default:
		return false
	}
}
