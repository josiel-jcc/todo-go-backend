package repositories

import (
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

type PushSubscriptionRepository interface {
	Upsert(userID uint, endpoint, p256dh, auth, userAgent string) error
	DeleteByEndpoint(userID uint, endpoint string) error
	ListByUserID(userID uint) ([]models.PushSubscription, error)
	DeleteByUserID(userID uint) error
}

type pushSubscriptionRepository struct{}

func NewPushSubscriptionRepository() PushSubscriptionRepository {
	return &pushSubscriptionRepository{}
}

func (r *pushSubscriptionRepository) Upsert(userID uint, endpoint, p256dh, auth, userAgent string) error {
	var sub models.PushSubscription
	err := database.DB.Where("user_id = ? AND endpoint = ?", userID, endpoint).First(&sub).Error
	if err == gorm.ErrRecordNotFound {
		sub = models.PushSubscription{
			UserID:    userID,
			Endpoint:  endpoint,
			P256dh:    p256dh,
			Auth:      auth,
			UserAgent: userAgent,
		}
		return database.DB.Create(&sub).Error
	}
	if err != nil {
		return err
	}
	sub.P256dh = p256dh
	sub.Auth = auth
	sub.UserAgent = userAgent
	return database.DB.Save(&sub).Error
}

func (r *pushSubscriptionRepository) DeleteByEndpoint(userID uint, endpoint string) error {
	return database.DB.Where("user_id = ? AND endpoint = ?", userID, endpoint).
		Delete(&models.PushSubscription{}).Error
}

func (r *pushSubscriptionRepository) ListByUserID(userID uint) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := database.DB.Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *pushSubscriptionRepository) DeleteByUserID(userID uint) error {
	return database.DB.Where("user_id = ?", userID).Delete(&models.PushSubscription{}).Error
}
