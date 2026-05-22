package services

import (
	"encoding/json"
	stderrors "errors"
	"strconv"
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/database/migrations"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"gorm.io/gorm"
)

type GroupService interface {
	AssertCanCollaborateWith(actorID, targetID uint) error
	AddUserToDefaultGroup(userID uint) error
	ListGroups(userID uint) ([]GroupListItem, error)
	CreateGroup(userID uint, name string) (*models.Group, error)
	GetGroup(userID, groupID uint) (*GroupDetail, error)
	UpdateGroup(userID, groupID uint, name string) (*models.Group, error)
	DeleteGroup(userID, groupID uint) error
	InviteUser(actorID, groupID, invitedUserID uint) (*models.GroupInvitation, error)
	CancelInvitation(actorID, groupID, invitationID uint) error
	RemoveMember(actorID, groupID, memberUserID uint) error
	ListReceivedInvitations(userID uint) ([]models.GroupInvitation, error)
	AcceptInvitation(userID, invitationID uint) error
	DeclineInvitation(userID, invitationID uint) error
	ListUsers(actorID uint, scope string, groupID *uint, page, limit int) ([]models.UserPublic, int64, error)
}

type GroupListItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	IsDefault   bool   `json:"is_default"`
	MemberCount int64  `json:"member_count"`
	CreatedBy   uint   `json:"created_by"`
}

type GroupDetail struct {
	ID                 uint                     `json:"id"`
	Name               string                   `json:"name"`
	IsDefault          bool                     `json:"is_default"`
	CreatedBy          uint                     `json:"created_by"`
	Members            []models.UserPublic      `json:"members"`
	PendingInvitations []models.GroupInvitation `json:"pending_invitations"`
}

type groupService struct {
	groupRepo        repositories.GroupRepository
	invitationRepo   repositories.GroupInvitationRepository
	notificationRepo repositories.UserNotificationRepository
	userRepo         repositories.UserRepository
}

func NewGroupService(
	groupRepo repositories.GroupRepository,
	invitationRepo repositories.GroupInvitationRepository,
	notificationRepo repositories.UserNotificationRepository,
	userRepo repositories.UserRepository,
) GroupService {
	return &groupService{
		groupRepo:        groupRepo,
		invitationRepo:   invitationRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

func (s *groupService) AssertCanCollaborateWith(actorID, targetID uint) error {
	if actorID == targetID {
		return nil
	}
	ok, err := s.groupRepo.UsersShareAtLeastOneGroup(actorID, targetID)
	if err != nil {
		return errors.NewInternalServerError(err)
	}
	if !ok {
		return errors.NewInvalidInputError("Só é possível colaborar com usuários do mesmo grupo")
	}
	return nil
}

func (s *groupService) AddUserToDefaultGroup(userID uint) error {
	return migrations.AddUserToDefaultGroup(database.DB, userID)
}

func (s *groupService) ListGroups(userID uint) ([]GroupListItem, error) {
	groups, err := s.groupRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := make([]GroupListItem, 0, len(groups))
	for _, g := range groups {
		count, err := s.groupRepo.CountMembers(g.ID)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		result = append(result, GroupListItem{
			ID:          g.ID,
			Name:        g.Name,
			IsDefault:   g.IsDefault,
			MemberCount: count,
			CreatedBy:   g.CreatedBy,
		})
	}
	return result, nil
}

func (s *groupService) CreateGroup(userID uint, name string) (*models.Group, error) {
	group := &models.Group{
		Name:      name,
		CreatedBy: userID,
		IsDefault: false,
	}
	if err := s.groupRepo.Create(group); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := s.groupRepo.AddMember(group.ID, userID); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return group, nil
}

func (s *groupService) GetGroup(userID, groupID uint) (*GroupDetail, error) {
	if err := s.requireMember(groupID, userID); err != nil {
		return nil, err
	}
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	members, err := s.groupRepo.ListMembers(groupID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	publicMembers := make([]models.UserPublic, 0, len(members))
	for _, m := range members {
		publicMembers = append(publicMembers, models.UserPublic{
			ID:        m.ID,
			Username:  m.Username,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	pending, err := s.invitationRepo.ListPendingByGroup(groupID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &GroupDetail{
		ID:                 group.ID,
		Name:               group.Name,
		IsDefault:          group.IsDefault,
		CreatedBy:          group.CreatedBy,
		Members:            publicMembers,
		PendingInvitations: pending,
	}, nil
}

func (s *groupService) UpdateGroup(userID, groupID uint, name string) (*models.Group, error) {
	if err := s.requireMember(groupID, userID); err != nil {
		return nil, err
	}
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	group.Name = name
	if err := s.groupRepo.Update(group); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return group, nil
}

func (s *groupService) DeleteGroup(userID, groupID uint) error {
	if err := s.requireMember(groupID, userID); err != nil {
		return err
	}
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return errors.NewInternalServerError(err)
	}
	if group.IsDefault {
		return errors.NewInvalidInputError("O grupo padrão não pode ser excluído")
	}
	return s.groupRepo.Delete(groupID)
}

func (s *groupService) InviteUser(actorID, groupID, invitedUserID uint) (*models.GroupInvitation, error) {
	if err := s.requireMember(groupID, actorID); err != nil {
		return nil, err
	}
	if invitedUserID == actorID {
		return nil, errors.NewInvalidInputError("Não é possível convidar a si mesmo")
	}
	if _, err := s.userRepo.FindByID(invitedUserID); err != nil {
		return nil, errors.NewInvalidInputError("Usuário não encontrado")
	}
	isMember, err := s.groupRepo.IsMember(groupID, invitedUserID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if isMember {
		return nil, errors.NewInvalidInputError("Usuário já é membro do grupo")
	}

	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	inviter, err := s.userRepo.FindByID(actorID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	var created *models.GroupInvitation
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var pendingCount int64
		if err := tx.Model(&models.GroupInvitation{}).
			Where("group_id = ? AND invited_user_id = ? AND status = ?", groupID, invitedUserID, models.GroupInvitationPending).
			Count(&pendingCount).Error; err != nil {
			return err
		}
		if pendingCount > 0 {
			return errors.NewInvalidInputError("Já existe um convite pendente para este usuário")
		}

		inv := &models.GroupInvitation{
			GroupID:         groupID,
			InvitedUserID:   invitedUserID,
			InvitedByUserID: actorID,
			Status:          models.GroupInvitationPending,
		}
		if err := tx.Create(inv).Error; err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]interface{}{
			"invitation_id":       inv.ID,
			"group_id":            group.ID,
			"group_name":          group.Name,
			"invited_by_username": inviter.Username,
		})
		if err != nil {
			return err
		}
		if err := tx.Create(&models.UserNotification{
			UserID:  invitedUserID,
			Type:    models.UserNotificationTypeGroupInvite,
			Payload: string(payload),
			Read:    false,
		}).Error; err != nil {
			return err
		}
		created = inv
		return nil
	})
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.NewInternalServerError(err)
	}
	return created, nil
}

func (s *groupService) CancelInvitation(actorID, groupID, invitationID uint) error {
	if err := s.requireMember(groupID, actorID); err != nil {
		return err
	}
	inv, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return errors.NewInvalidInputError("Convite não encontrado")
	}
	if inv.GroupID != groupID || inv.Status != models.GroupInvitationPending {
		return errors.NewInvalidInputError("Convite inválido")
	}
	return s.invitationRepo.UpdateStatus(invitationID, models.GroupInvitationCancelled)
}

func (s *groupService) RemoveMember(actorID, groupID, memberUserID uint) error {
	if err := s.requireMember(groupID, actorID); err != nil {
		return err
	}
	return s.groupRepo.RemoveMember(groupID, memberUserID)
}

func (s *groupService) ListReceivedInvitations(userID uint) ([]models.GroupInvitation, error) {
	list, err := s.invitationRepo.ListPendingReceived(userID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return list, nil
}

func (s *groupService) AcceptInvitation(userID, invitationID uint) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var inv models.GroupInvitation
		if err := tx.First(&inv, invitationID).Error; err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return errors.NewInvalidInputError("Convite não encontrado")
			}
			return err
		}
		if inv.InvitedUserID != userID || inv.Status != models.GroupInvitationPending {
			return errors.NewForbiddenError()
		}
		member := models.GroupMember{
			GroupID:  inv.GroupID,
			UserID:   userID,
			JoinedAt: time.Now(),
		}
		if err := tx.Where(models.GroupMember{GroupID: inv.GroupID, UserID: userID}).
			FirstOrCreate(&member).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&models.GroupInvitation{}).Where("id = ?", invitationID).
			Updates(map[string]interface{}{
				"status":       models.GroupInvitationAccepted,
				"responded_at": now,
			}).Error; err != nil {
			return err
		}
		pattern := "%\"invitation_id\":" + strconv.FormatUint(uint64(invitationID), 10) + "%"
		return tx.Model(&models.UserNotification{}).
			Where("user_id = ? AND type = ? AND `read` = ? AND payload LIKE ?",
				userID, models.UserNotificationTypeGroupInvite, false, pattern).
			Update("`read`", true).Error
	})
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return appErr
		}
		return errors.NewInternalServerError(err)
	}
	return nil
}

func (s *groupService) DeclineInvitation(userID, invitationID uint) error {
	inv, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return errors.NewInvalidInputError("Convite não encontrado")
	}
	if inv.InvitedUserID != userID || inv.Status != models.GroupInvitationPending {
		return errors.NewForbiddenError()
	}
	if err := s.invitationRepo.UpdateStatus(invitationID, models.GroupInvitationDeclined); err != nil {
		return errors.NewInternalServerError(err)
	}
	_ = s.notificationRepo.MarkReadByInvitationID(userID, invitationID)
	return nil
}

func (s *groupService) ListUsers(actorID uint, scope string, groupID *uint, page, limit int) ([]models.UserPublic, int64, error) {
	if scope == "invite" {
		var excludeIDs []uint
		excludeIDs = append(excludeIDs, actorID)
		if groupID != nil {
			if err := s.requireMember(*groupID, actorID); err != nil {
				return nil, 0, err
			}
			memberIDs, err := s.groupRepo.ListMemberIDs(*groupID)
			if err != nil {
				return nil, 0, errors.NewInternalServerError(err)
			}
			excludeIDs = append(excludeIDs, memberIDs...)
			pending, err := s.invitationRepo.ListPendingByGroup(*groupID)
			if err != nil {
				return nil, 0, errors.NewInternalServerError(err)
			}
			for _, inv := range pending {
				excludeIDs = append(excludeIDs, inv.InvitedUserID)
			}
		}
		return s.userRepo.FindAllPaginatedExcluding(actorID, excludeIDs, page, limit)
	}
	return s.userRepo.FindCoGroupUsersPaginated(actorID, page, limit)
}

func (s *groupService) requireMember(groupID, userID uint) error {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return errors.NewInternalServerError(err)
	}
	if !ok {
		return errors.NewForbiddenError()
	}
	return nil
}
