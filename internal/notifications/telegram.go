package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"todo-go-backend/internal/models"
)

// TelegramService handles Telegram notifications
type TelegramService struct {
	botToken string
	apiURL   string
}

// NewTelegramService creates a new Telegram service
func NewTelegramService(botToken string) *TelegramService {
	return &TelegramService{
		botToken: botToken,
		apiURL:   "https://api.telegram.org/bot" + botToken,
	}
}

// SendNotification sends a notification via Telegram.
// minutesBefore is used for task_reminder messages (effective offset before due_date).
func (s *TelegramService) SendNotification(chatID string, task *models.Task, notificationType models.NotificationType, minutesBefore int) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	if chatID == "" {
		return fmt.Errorf("user telegram chat ID not configured")
	}

	message := s.buildMessage(task, notificationType, minutesBefore)

	url := fmt.Sprintf("%s/sendMessage", s.apiURL)
	
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)
		
		// Parse error response for better error messages
		var errorResp struct {
			OK          bool   `json:"ok"`
			ErrorCode   int    `json:"error_code"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			switch errorResp.ErrorCode {
			case 400:
				if errorResp.Description == "Bad Request: chat not found" {
					return fmt.Errorf("chat not found: user needs to send a message to the bot first (chat_id: %s)", chatID)
				}
				return fmt.Errorf("telegram API error (400): %s", errorResp.Description)
			case 401:
				return fmt.Errorf("telegram API error (401): invalid bot token")
			case 403:
				return fmt.Errorf("telegram API error (403): bot was blocked by user")
			default:
				return fmt.Errorf("telegram API error (%d): %s", errorResp.ErrorCode, errorResp.Description)
			}
		}
		
		return fmt.Errorf("telegram API error: %s", errorMsg)
	}

	return nil
}

// buildMessage builds Telegram message based on notification type
func (s *TelegramService) buildMessage(task *models.Task, notificationType models.NotificationType, minutesBefore int) string {
	var emoji string
	var title string

	switch notificationType {
	case models.NotificationTypeTaskReminder:
		return fmt.Sprintf(
			"🔔 Lembrete: a tarefa <b>%s</b> vence em %d minutos",
			task.Title,
			minutesBefore,
		)
	case models.NotificationTypeDueSoon:
		emoji = "⏰"
		title = "Tarefa vence amanhã!"
	case models.NotificationTypeDueToday:
		emoji = "📅"
		title = "Tarefa vence hoje!"
	case models.NotificationTypeOverdue:
		emoji = "⚠️"
		title = "Tarefa atrasada!"
	}

	dueDateStr := ""
	if task.DueDate != nil {
		dueDateStr = task.DueDate.Format("02/01/2006")
	}

	message := fmt.Sprintf(
		"%s <b>%s</b>\n\n"+
			"<b>%s</b>\n"+
			"%s\n\n"+
			"<b>Prioridade:</b> %s\n"+
			"<b>Data de vencimento:</b> %s",
		emoji,
		title,
		task.Title,
		task.Description,
		task.Priority,
		dueDateStr,
	)

	return message
}

