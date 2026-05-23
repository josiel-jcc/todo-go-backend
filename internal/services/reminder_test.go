package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidReminderMinutes(t *testing.T) {
	for _, m := range AllowedReminderMinutes {
		assert.True(t, IsValidReminderMinutes(m), "expected %d to be valid", m)
	}
	assert.False(t, IsValidReminderMinutes(7))
	assert.False(t, IsValidReminderMinutes(0))
}
