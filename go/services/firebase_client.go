package services

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FirebaseClient handles Firebase operations and FCM notifications
type FirebaseClient struct {
	app *firebase.App
	ctx context.Context
}

// NewFirebaseClient creates a new Firebase client
func NewFirebaseClient() (*FirebaseClient, error) {
	ctx := context.Background()

	// Get project ID from environment
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID not configured")
	}

	// Try to get service account key path from environment
	keyPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")

	// If path is not set, try JSON content
	keyJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_JSON")

	var opts []option.ClientOption

	if keyPath != "" {
		// Use service account file
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("service account key file not found: %s", keyPath)
		}

		// Set GOOGLE_APPLICATION_CREDENTIALS environment variable
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", keyPath)

		opts = append(opts, option.WithCredentialsFile(keyPath))
		log.Printf("Firebase: Using service account file: %s", keyPath)
		log.Printf("Firebase: GOOGLE_APPLICATION_CREDENTIALS set to: %s", keyPath)
	} else if keyJSON != "" {
		// Use service account JSON from environment
		opts = append(opts, option.WithCredentialsJSON([]byte(keyJSON)))
		log.Println("Firebase: Using service account from environment variable")
	} else {
		// No credentials provided, return error
		return nil, fmt.Errorf("Firebase credentials not configured. Set FIREBASE_SERVICE_ACCOUNT_KEY_PATH or FIREBASE_SERVICE_ACCOUNT_KEY_JSON")
	}

	// Initialize Firebase app with explicit project ID
	config := &firebase.Config{
		ProjectID: projectID,
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	log.Printf("Firebase: Client initialized successfully for project: %s", projectID)

	return &FirebaseClient{
		app: app,
		ctx: ctx,
	}, nil
}

// SendNotification sends a push notification to a specific device
func (fc *FirebaseClient) SendNotification(token string, title, body string, data map[string]string) error {
	if fc == nil || fc.app == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	// Get FCM messaging client
	messagingClient, err := fc.app.Messaging(fc.ctx)
	if err != nil {
		return fmt.Errorf("error getting messaging client: %w", err)
	}

	// Build the message
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "frigate_events",
				Title:     title,
				Body:      body,
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
					Badge: func() *int { i := 1; return &i }(),
				},
			},
		},
	}

	// Send the message
	_, err = messagingClient.Send(fc.ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("[FCM] Notification sent to token: %s...", token[:20])
	return nil
}

// SendMulticastNotification sends a notification to multiple devices
// Returns list of invalid tokens that should be removed from database
// Note: SendMulticast is deprecated, so we implement our own batch sending
func (fc *FirebaseClient) SendMulticastNotification(tokens []string, title, body string, data map[string]string) ([]string, error) {
	if fc == nil || fc.app == nil {
		return nil, fmt.Errorf("firebase client not initialized")
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	// Get FCM messaging client
	messagingClient, err := fc.app.Messaging(fc.ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %w", err)
	}

	log.Printf("[FCM] Sending notification to %d devices (batch mode)", len(tokens))

	var invalidTokens []string
	successCount := 0
	failureCount := 0

	// Send to each token individually
	for _, token := range tokens {
		// Build message for this token
		message := &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "frigate_events",
					Title:     title,
					Body:      body,
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: title,
							Body:  body,
						},
						Sound: "default",
						Badge: func() *int { i := 1; return &i }(),
					},
				},
			},
		}

		// Send message
		_, err := messagingClient.Send(fc.ctx, message)
		if err != nil {
			failureCount++
			invalidTokens = append(invalidTokens, token)
			log.Printf("[FCM] Failed to send to token %s...: %v", token[:20], err)
		} else {
			successCount++
		}
	}

	log.Printf("[FCM] Batch notification sent: %d successful, %d failed", successCount, failureCount)

	return invalidTokens, nil
}
