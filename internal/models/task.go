package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskType represents the type of a task
type TaskType string

const (
	// TaskTypeCasa represents tasks related to home
	TaskTypeCasa TaskType = "casa"
	// TaskTypeTrabalho represents tasks related to work
	TaskTypeTrabalho TaskType = "trabalho"
	// TaskTypeLazer represents tasks related to leisure
	TaskTypeLazer TaskType = "lazer"
	// TaskTypeSaude represents tasks related to health
	TaskTypeSaude TaskType = "saude"
)

// Priority represents the priority level of a task
type Priority string

const (
	// PriorityBaixa represents low priority
	PriorityBaixa Priority = "baixa"
	// PriorityMedia represents medium priority
	PriorityMedia Priority = "media"
	// PriorityAlta represents high priority
	PriorityAlta Priority = "alta"
	// PriorityUrgente represents urgent priority
	PriorityUrgente Priority = "urgente"
)

// Task represents a task in the system
// A task belongs to a user and can be assigned by another user.
// Tasks can be shared with other users (many-to-many); when a user creates a task for another, both have access.
type Task struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Title            string         `json:"title" gorm:"type:varchar(200);not null"`
	Description      string         `json:"description" gorm:"type:text"`
	Type             TaskType       `json:"type" gorm:"type:varchar(20);not null"`
	Priority         Priority       `json:"priority" gorm:"type:varchar(20);default:'media'"` // Task priority
	DueDate               *time.Time `json:"due_date" gorm:"index:idx_tasks_reminder_candidates"` // Deadline for task completion
	Completed             bool       `json:"completed" gorm:"default:false;index:idx_tasks_reminder_candidates"`
	ReminderMinutesBefore *int       `json:"reminder_minutes_before,omitempty"` // NULL = inherit user default
	UserID           uint           `json:"user_id" gorm:"not null;index"` // ID of the user responsible for the task (owner)
	AssignedBy       *uint          `json:"assigned_by"`                   // ID of the user who created/assigned the task (nil if created by the user themselves)
	User             User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AssignedByUser   *User          `json:"assigned_by_user,omitempty" gorm:"foreignKey:AssignedBy"`
	SharedWithUsers  []User         `json:"shared_with,omitempty" gorm:"many2many:task_shared_with;"` // Users with whom the task is shared (no limit)
	Tags             []Tag          `json:"tags,omitempty" gorm:"many2many:task_tags;"`             // Tags associated with the task
	Comments         []Comment      `json:"comments,omitempty" gorm:"foreignKey:TaskID"`           // Comments on the task
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// TaskSharedWith is the join table for sharing tasks with users (task_id, user_id).
// Used for FirstOrCreate/Delete; the same table is used by Task.SharedWithUsers many2many.
type TaskSharedWith struct {
	TaskID uint `gorm:"primaryKey"`
	UserID uint `gorm:"primaryKey"`
}

// TableName returns the table name for TaskSharedWith
func (TaskSharedWith) TableName() string {
	return "task_shared_with"
}

// Tag represents a custom tag that can be associated with tasks
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(50);not null"`
	Color     string         `json:"color" gorm:"type:varchar(7)"`  // Hex color code (e.g., #FF5733)
	UserID    uint           `json:"user_id" gorm:"not null;index"` // Tags are user-specific
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tasks     []Task         `json:"tasks,omitempty" gorm:"many2many:task_tags;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
