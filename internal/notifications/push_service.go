package notifications

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/repositories"
)

// PushPayload is the JSON body sent to Web Push endpoints.
type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// PushService sends Web Push notifications via VAPID.
type PushService struct {
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
	repo            repositories.PushSubscriptionRepository
}

// NewPushService creates a Web Push service.
func NewPushService(cfg *config.Config, pushSubscriptionRepo repositories.PushSubscriptionRepository) *PushService {
	return &PushService{
		vapidPublicKey:  cfg.VAPIDPublicKey,
		vapidPrivateKey: cfg.VAPIDPrivateKey,
		vapidSubject:    cfg.VAPIDSubject,
		repo:            pushSubscriptionRepo,
	}
}

// PublicKey returns the VAPID public key for client subscription.
func (s *PushService) PublicKey() string {
	return s.vapidPublicKey
}

// SendToUser sends a push notification to all subscriptions for the user.
// If VAPID keys are not configured, this is a no-op (returns nil).
func (s *PushService) SendToUser(userID uint, payload PushPayload) error {
	if s.vapidPublicKey == "" || s.vapidPrivateKey == "" {
		log.Println("[PushService] VAPID keys not configured, skipping web push")
		return nil
	}

	subs, err := s.repo.ListByUserID(userID)
	if err != nil {
		return fmt.Errorf("list push subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	var lastErr error
	successCount := 0
	for _, sub := range subs {
		if err := s.sendToSubscription(userID, sub.Endpoint, sub.P256dh, sub.Auth, body); err != nil {
			log.Printf("[PushService] failed to send to user %d: %v", userID, err)
			lastErr = err
			continue
		}
		successCount++
	}

	if successCount > 0 {
		return nil
	}
	return lastErr
}

func (s *PushService) sendToSubscription(userID uint, endpoint, p256dh, auth string, body []byte) error {
	sub := &webpush.Subscription{
		Endpoint: endpoint,
		Keys: webpush.Keys{
			P256dh: p256dh,
			Auth:   auth,
		},
	}

	resp, err := webpush.SendNotification(body, sub, &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             86400,
	})
	if err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	case http.StatusNotFound, http.StatusGone:
		if delErr := s.repo.DeleteByEndpoint(userID, endpoint); delErr != nil {
			log.Printf("[PushService] failed to remove expired subscription: %v", delErr)
		} else {
			log.Printf("[PushService] removed expired push subscription for user %d (status %d)", userID, resp.StatusCode)
		}
		return nil
	default:
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push endpoint returned %d: %s", resp.StatusCode, string(respBody))
	}
}
