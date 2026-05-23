package notifications

import (
	"testing"
	"time"
	"todo-go-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestEffectiveReminderMinutes_taskOverride(t *testing.T) {
	taskMinutes := 30
	task := models.Task{ReminderMinutesBefore: &taskMinutes}
	user := models.User{ReminderMinutesBefore: 10}

	assert.Equal(t, 30, effectiveReminderMinutes(task, user))
}

func TestEffectiveReminderMinutes_userDefault(t *testing.T) {
	task := models.Task{}
	user := models.User{ReminderMinutesBefore: 15}

	assert.Equal(t, 15, effectiveReminderMinutes(task, user))
}

func TestShouldSendInWindow(t *testing.T) {
	now := time.Date(2026, 5, 22, 14, 10, 0, 0, time.UTC)

	tests := []struct {
		name       string
		reminderAt time.Time
		want       bool
	}{
		{
			name:       "inside window",
			reminderAt: time.Date(2026, 5, 22, 14, 9, 30, 0, time.UTC),
			want:       true,
		},
		{
			name:       "at window start",
			reminderAt: time.Date(2026, 5, 22, 14, 9, 0, 0, time.UTC),
			want:       true,
		},
		{
			name:       "before window (missed)",
			reminderAt: time.Date(2026, 5, 22, 14, 8, 59, 0, time.UTC),
			want:       false,
		},
		{
			name:       "at now (not yet)",
			reminderAt: now,
			want:       false,
		},
		{
			name:       "after now (not yet)",
			reminderAt: time.Date(2026, 5, 22, 14, 10, 1, 0, time.UTC),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldSendInWindow(tt.reminderAt, now))
		})
	}
}

func TestTaskDueDateFingerprint_truncatesToUTCSeconds(t *testing.T) {
	due := time.Date(2026, 5, 22, 15, 0, 0, 500_000_000, time.FixedZone("BRT", -3*3600))
	fp := taskDueDateFingerprint(due)

	assert.Equal(t, time.UTC, fp.Location())
	assert.Equal(t, 0, fp.Nanosecond())
	assert.Equal(t, time.Date(2026, 5, 22, 18, 0, 0, 0, time.UTC), fp)
}
