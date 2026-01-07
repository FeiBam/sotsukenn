package services

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sotsukenn/go/models"

	"gorm.io/gorm"
)

// NotificationService handles sending FCM notifications for Frigate events
type NotificationService struct {
	db               *gorm.DB
	firebaseClient   *FirebaseClient
	debounceCache    map[string]time.Time
	debounceMutex    sync.RWMutex
	debounceDuration time.Duration
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, fbClient *FirebaseClient) *NotificationService {
	duration := 30 * time.Second
	if d := os.Getenv("FCM_DEBOUNCE_DURATION"); d != "" {
		if seconds, err := strconv.Atoi(d); err == nil {
			duration = time.Duration(seconds) * time.Second
		}
	}

	return &NotificationService{
		db:               db,
		firebaseClient:   fbClient,
		debounceCache:    make(map[string]time.Time),
		debounceDuration: duration,
	}
}

// ShouldSendNotification checks if a notification should be sent for this event
func (ns *NotificationService) ShouldSendNotification(event models.FrigateEvent) bool {
	// Check if FCM notifications are enabled
	if os.Getenv("FCM_NOTIFICATIONS_ENABLED") != "true" {
		return false
	}

	// Check event type
	allowedTypes := strings.Split(os.Getenv("FCM_NOTIFY_ON_EVENT_TYPE"), ",")
	typeAllowed := false
	for _, t := range allowedTypes {
		if strings.TrimSpace(t) == event.Type {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return false
	}

	// Check label
	allowedLabels := strings.Split(os.Getenv("FCM_NOTIFY_LABELS"), ",")
	labelAllowed := false
	for _, l := range allowedLabels {
		if strings.TrimSpace(l) == event.After.Label {
			labelAllowed = true
			break
		}
	}
	if !labelAllowed {
		return false
	}

	return true
}

// GenerateNotificationContent creates notification title and body from event
func (ns *NotificationService) GenerateNotificationContent(event models.FrigateEvent) (title, body string, data map[string]string) {
	camera := event.After.Camera
	label := event.After.Label
	eventType := event.Type

	// Check for face recognition (sub_label)
	if event.After.SubLabel != nil {
		if personName := ns.extractPersonName(event.After.SubLabel); personName != "" {
			if eventType == models.EventTypeNew {
				title = "检测到人脸"
				body = camera + " 检测到：" + personName
			} else if eventType == models.EventTypeEnd {
				title = personName + " 已离开"
				body = camera + " 检测结束"
			} else {
				title = personName
				body = camera + " " + label
			}
		} else {
			if eventType == models.EventTypeNew {
				title = "检测到人"
				body = camera + " 检测到：" + label
			} else if eventType == models.EventTypeEnd {
				title = "人已离开"
				body = camera + " 检测结束"
			} else {
				title = "人员更新"
				body = camera + " " + label
			}
		}
	} else {
		if eventType == models.EventTypeNew {
			title = "检测到人"
			body = camera + " 检测到：" + label
		} else if eventType == models.EventTypeEnd {
			title = "人已离开"
			body = camera + " 检测结束"
		} else {
			title = "人员更新"
			body = camera + " " + label
		}
	}

	data = map[string]string{
		"camera":    camera,
		"label":     label,
		"event_id":  event.After.ID,
		"event_type": eventType,
		"timestamp": strconv.FormatFloat(event.After.StartTime, 'f', 0, 64),
	}

	return title, body, data
}

// extractPersonName extracts person name from sub_label
// sub_label can be: ["John Smith", 0.79] or null
func (ns *NotificationService) extractPersonName(subLabel interface{}) string {
	if subLabel == nil {
		return ""
	}

	// Try to parse as array
	switch v := subLabel.(type) {
	case []interface{}:
		if len(v) >= 1 {
			if name, ok := v[0].(string); ok {
				return name
			}
		}
	case []string:
		if len(v) >= 1 {
			return v[0]
		}
	case string:
		return v
	}

	return ""
}

// SendNotification sends FCM notification to all users for a Frigate event
func (ns *NotificationService) SendNotification(event models.FrigateEvent) error {
	// Check if we should send notification
	if !ns.ShouldSendNotification(event) {
		return nil
	}

	// Check debounce
	debounceKey := event.After.ID + "_" + event.Type
	if ns.isDebounced(debounceKey) {
		log.Printf("[FCM] Notification debounced: %s", debounceKey)
		return nil
	}

	// Generate notification content
	title, body, data := ns.GenerateNotificationContent(event)

	// Get all active FCM tokens from all users
	var tokens []models.FCMToken
	err := ns.db.Where("is_active = ?", true).Find(&tokens).Error
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		log.Println("[FCM] No active tokens found")
		return nil
	}

	// Extract token strings
	tokenList := make([]string, len(tokens))
	for i, token := range tokens {
		tokenList[i] = token.Token
	}

	// Send multicast notification and get invalid tokens
	invalidTokens, err := ns.firebaseClient.SendMulticastNotification(tokenList, title, body, data)
	if err != nil {
		log.Printf("[FCM] Failed to send notification: %v", err)
		return err
	}

	// Remove invalid tokens from database
	if len(invalidTokens) > 0 {
		log.Printf("[FCM] Removing %d invalid tokens from database", len(invalidTokens))
		result := ns.db.Where("token IN ?", invalidTokens).Delete(&models.FCMToken{})
		if result.Error != nil {
			log.Printf("[FCM] Failed to remove invalid tokens: %v", result.Error)
		} else {
			log.Printf("[FCM] Removed %d invalid tokens", result.RowsAffected)
		}
	}

	log.Printf("[FCM] Notification sent: %s - %s", title, body)

	// Mark as sent in debounce cache
	ns.markAsSent(debounceKey)

	return nil
}

// isDebounced checks if a notification should be debounced
func (ns *NotificationService) isDebounced(key string) bool {
	ns.debounceMutex.RLock()
	defer ns.debounceMutex.RUnlock()

	if lastSent, exists := ns.debounceCache[key]; exists {
		return time.Since(lastSent) < ns.debounceDuration
	}
	return false
}

// markAsSent marks a notification as sent in the debounce cache
func (ns *NotificationService) markAsSent(key string) {
	ns.debounceMutex.Lock()
	defer ns.debounceMutex.Unlock()
	ns.debounceCache[key] = time.Now()

	// Clean up old entries periodically
	go ns.cleanDebounceCache()
}

// cleanDebounceCache removes old entries from the debounce cache
func (ns *NotificationService) cleanDebounceCache() {
	time.Sleep(ns.debounceDuration)

	ns.debounceMutex.Lock()
	defer ns.debounceMutex.Unlock()

	now := time.Now()
	for key, lastSent := range ns.debounceCache {
		if now.Sub(lastSent) > ns.debounceDuration {
			delete(ns.debounceCache, key)
		}
	}
}
