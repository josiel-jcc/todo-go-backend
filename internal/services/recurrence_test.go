package services

import (
	"testing"
	"time"

	"todo-go-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestComputeNextDue(t *testing.T) {
	base := time.Date(2026, 5, 29, 18, 0, 0, 0, time.UTC)

	assert.Equal(t, time.Date(2026, 5, 30, 18, 0, 0, 0, time.UTC), ComputeNextDue(base, models.RecurrenceDaily))
	assert.Equal(t, time.Date(2026, 6, 5, 18, 0, 0, 0, time.UTC), ComputeNextDue(base, models.RecurrenceWeekly))
	assert.Equal(t, time.Date(2026, 6, 29, 18, 0, 0, 0, time.UTC), ComputeNextDue(base, models.RecurrenceMonthly))
}

func TestAdvanceToFutureDue(t *testing.T) {
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)

	next := AdvanceToFutureDue(base, models.RecurrenceWeekly, now)
	assert.True(t, next.After(now))
	assert.Equal(t, time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), next)
}

func TestIsValidRecurrenceRule(t *testing.T) {
	assert.True(t, IsValidRecurrenceRule(models.RecurrenceDaily))
	assert.False(t, IsValidRecurrenceRule(models.RecurrenceRule("yearly")))
}
