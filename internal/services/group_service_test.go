package services

import (
	"testing"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/database/migrations"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGroupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMember{},
		&models.GroupInvitation{},
		&models.UserNotification{},
	))
	database.DB = db
	return db
}

func TestEnsureDefaultGroup_Idempotent(t *testing.T) {
	db := setupGroupTestDB(t)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	require.NoError(t, migrations.EnsureDefaultGroup(db))
	require.NoError(t, migrations.EnsureDefaultGroup(db))

	var groupCount int64
	db.Model(&models.Group{}).Where("is_default = ?", true).Count(&groupCount)
	assert.Equal(t, int64(1), groupCount)

	var memberCount int64
	db.Model(&models.GroupMember{}).Count(&memberCount)
	assert.Equal(t, int64(2), memberCount)
}

func TestAssertCanCollaborateWith_SameGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()

	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	g := &models.Group{Name: "Test", CreatedBy: u1.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))
	require.NoError(t, groupRepo.AddMember(g.ID, u2.ID))

	assert.NoError(t, svc.AssertCanCollaborateWith(u1.ID, u2.ID))
	assert.Error(t, svc.AssertCanCollaborateWith(u1.ID, 999))
}

func TestListUsers_InviteScopeRequiresMembership(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()
	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	u3 := &models.User{Username: "c", Email: "c@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)
	require.NoError(t, db.Create(u3).Error)

	g := &models.Group{Name: "Private", CreatedBy: u1.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))

	gid := g.ID
	_, _, err := svc.ListUsers(u2.ID, "invite", &gid, 1, 10)
	assert.Error(t, err)

	users, _, err := svc.ListUsers(u1.ID, "invite", &gid, 1, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
}

func TestInviteUser_CreatesNotification(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()
	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	g := &models.Group{Name: "Team", CreatedBy: u1.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))

	inv, err := svc.InviteUser(u1.ID, g.ID, u2.ID)
	require.NoError(t, err)
	require.NotNil(t, inv)

	var notifCount int64
	db.Model(&models.UserNotification{}).Where("user_id = ?", u2.ID).Count(&notifCount)
	assert.Equal(t, int64(1), notifCount)

	_, err = svc.InviteUser(u1.ID, g.ID, u2.ID)
	assert.Error(t, err)
}

func TestAcceptInvitation_AddsMember(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()
	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	g := &models.Group{Name: "Team", CreatedBy: u1.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))

	inv, err := svc.InviteUser(u1.ID, g.ID, u2.ID)
	require.NoError(t, err)

	require.NoError(t, svc.AcceptInvitation(u2.ID, inv.ID))

	ok, err := groupRepo.IsMember(g.ID, u2.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	var invStatus models.GroupInvitation
	require.NoError(t, db.First(&invStatus, inv.ID).Error)
	assert.Equal(t, models.GroupInvitationAccepted, invStatus.Status)
}

func TestDeclineInvitation(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()
	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	u2 := &models.User{Username: "b", Email: "b@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	g := &models.Group{Name: "Team", CreatedBy: u1.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))

	inv, err := svc.InviteUser(u1.ID, g.ID, u2.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeclineInvitation(u2.ID, inv.ID))

	ok, err := groupRepo.IsMember(g.ID, u2.ID)
	require.NoError(t, err)
	assert.False(t, ok)

	var invStatus models.GroupInvitation
	require.NoError(t, db.First(&invStatus, inv.ID).Error)
	assert.Equal(t, models.GroupInvitationDeclined, invStatus.Status)
}

func TestDeleteGroup_DefaultNotAllowed(t *testing.T) {
	db := setupGroupTestDB(t)
	groupRepo := repositories.NewGroupRepository()
	invRepo := repositories.NewGroupInvitationRepository()
	notifRepo := repositories.NewUserNotificationRepository()
	userRepo := repositories.NewUserRepository()
	svc := NewGroupService(groupRepo, invRepo, notifRepo, userRepo)

	u1 := &models.User{Username: "a", Email: "a@test.com", Password: "hash"}
	require.NoError(t, db.Create(u1).Error)

	g := &models.Group{Name: "Os de casa", CreatedBy: u1.ID, IsDefault: true}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, u1.ID))

	err := svc.DeleteGroup(u1.ID, g.ID)
	assert.Error(t, err)
}
