package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"sotsukenn/go/models"
	"sotsukenn/go/services"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
)

var (
	firebaseClient *services.FirebaseClient
	firebaseOnce   sync.Once
	firebaseMutex  sync.RWMutex
)

// initFirebaseClient initializes the Firebase client singleton
func initFirebaseClient() {
	client, err := services.NewFirebaseClient()
	if err != nil {
		// Don't panic, just log the error
		// This allows the server to start even if Firebase is not configured
		log.Printf("Failed to initialize Firebase client: %v", err)
		firebaseClient = nil
	} else {
		firebaseClient = client
	}
}

// getFirebaseClient returns the Firebase client instance (lazy initialization)
func getFirebaseClient() *services.FirebaseClient {
	firebaseMutex.Lock()
	defer firebaseMutex.Unlock()

	if firebaseClient == nil {
		firebaseOnce.Do(initFirebaseClient)
	}

	return firebaseClient
}

// RegisterFCMTokenRequest represents a request to register an FCM token
type RegisterFCMTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	DeviceName string `json:"device_name"`
	ReplaceAll bool   `json:"replace_all"` // If true, deactivate all old tokens for this user
}

// RegisterFCMToken registers a new FCM token for the authenticated user
// POST /api/fcm/tokens
func RegisterFCMToken(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req RegisterFCMTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// If replace_all is true, deactivate all existing tokens for this user
	if req.ReplaceAll {
		log.Printf("[FCM] Deactivating all old tokens for user %v", userID)
		db.Model(&models.FCMToken{}).
			Where("user_id = ?", userID).
			Update("is_active", false)
	}

	// Check if token already exists for this user
	var existingToken models.FCMToken
	err = db.Where("user_id = ? AND token = ?", userID, req.Token).First(&existingToken).Error

	if err == nil {
		// Token exists, reactivate it if needed and update device name
		if !existingToken.IsActive {
			existingToken.IsActive = true
		}
		if req.DeviceName != "" && req.DeviceName != existingToken.DeviceName {
			existingToken.DeviceName = req.DeviceName
		}
		db.Save(&existingToken)

		log.Printf("[FCM] Token reactivated for user %v, device: %s", userID, existingToken.DeviceName)
		ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Token reactivated successfully", "", gin.H{
			"id":          existingToken.ID,
			"token":       req.Token[:20] + "...",
			"device_name": existingToken.DeviceName,
			"is_active":   existingToken.IsActive,
		}))
		return
	}

	// Create new token
	fcmToken := models.FCMToken{
		UserID:     userID.(uint),
		Token:      req.Token,
		DeviceName: req.DeviceName,
		IsActive:   true,
	}

	if err := db.Create(&fcmToken).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to register token", err.Error())
		return
	}

	log.Printf("[FCM] New token registered for user %v, device: %s", userID, fcmToken.DeviceName)
	ctx.JSON(http.StatusCreated, utils.JsonResponse("success", http.StatusCreated, "Token registered successfully", "", gin.H{
		"id":          fcmToken.ID,
		"token":       fcmToken.Token[:20] + "...",
		"device_name": fcmToken.DeviceName,
		"is_active":   fcmToken.IsActive,
	}))
}

// GetFCMTokens retrieves all FCM tokens for the authenticated user
// GET /api/fcm/tokens
func GetFCMTokens(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var tokens []models.FCMToken
	err = db.Where("user_id = ?", userID).Find(&tokens).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to retrieve tokens", err.Error())
		return
	}

	// Format response
	response := make([]gin.H, len(tokens))
	for i, token := range tokens {
		response[i] = gin.H{
			"id":         token.ID,
			"token":      token.Token[:20] + "...", // Only show first 20 chars
			"device_name": token.DeviceName,
			"is_active":  token.IsActive,
			"created_at": token.CreatedAt,
		}
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Tokens retrieved", "", response))
}

// DeleteFCMToken deletes an FCM token
// DELETE /api/fcm/tokens/:id
func DeleteFCMToken(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	tokenID := ctx.Param("id")
	if tokenID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Token ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(tokenID, 10, 32)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid token ID", err.Error())
		return
	}

	// Check if token belongs to user
	var fcmToken models.FCMToken
	err = db.Where("id = ? AND user_id = ?", id, userID).First(&fcmToken).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Token not found", nil)
		return
	}

	// Delete token
	if err := db.Delete(&fcmToken).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to delete token", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Token deleted successfully", "", nil))
}

// UpdateFCMTokenRequest represents a request to update an FCM token
type UpdateFCMTokenRequest struct {
	DeviceName string `json:"device_name"`
	IsActive   *bool  `json:"is_active"`
}

// UpdateFCMToken updates an FCM token
// PUT /api/fcm/tokens/:id
func UpdateFCMToken(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	tokenID := ctx.Param("id")
	if tokenID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Token ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(tokenID, 10, 32)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid token ID", err.Error())
		return
	}

	var req UpdateFCMTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Check if token belongs to user
	var fcmToken models.FCMToken
	err = db.Where("id = ? AND user_id = ?", id, userID).First(&fcmToken).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Token not found", nil)
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.DeviceName != "" {
		updates["device_name"] = req.DeviceName
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := db.Model(&fcmToken).Updates(updates).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to update token", err.Error())
		return
	}

	// Refresh from database
	db.First(&fcmToken, fcmToken.ID)

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Token updated successfully", "", gin.H{
		"id":         fcmToken.ID,
		"device_name": fcmToken.DeviceName,
		"is_active":  fcmToken.IsActive,
	}))
}

// TestNotificationRequest represents a test notification request
type TestNotificationRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// SendTestNotification sends a test notification to all user's devices
// POST /api/fcm/test
func SendTestNotification(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get Firebase client
	client := getFirebaseClient()
	if client == nil {
		utils.RespondWithError(ctx, http.StatusServiceUnavailable, "Firebase not configured", "Please configure Firebase credentials")
		return
	}

	// Get request body
	var req TestNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Use default values if request body is empty
		req.Title = "测试通知"
		req.Body = "这是一条测试通知"
	}

	// Get all active tokens for user
	var tokens []models.FCMToken
	err = db.Where("user_id = ? AND is_active = ?", userID, true).Find(&tokens).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to retrieve tokens", err.Error())
		return
	}

	if len(tokens) == 0 {
		utils.RespondWithError(ctx, http.StatusNotFound, "No active tokens found", nil)
		return
	}

	// Extract token strings
	tokenList := make([]string, len(tokens))
	for i, token := range tokens {
		tokenList[i] = token.Token
	}

	// Send multicast notification
	data := map[string]string{
		"type":      "test",
		"timestamp": strconv.FormatInt(ctx.GetInt64("now"), 10),
	}

	invalidTokens, err := client.SendMulticastNotification(tokenList, req.Title, req.Body, data)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to send notification", err.Error())
		return
	}

	// Remove invalid tokens from database
	if len(invalidTokens) > 0 {
		log.Printf("[FCM] Removing %d invalid tokens from database", len(invalidTokens))
		result := db.Where("token IN ?", invalidTokens).Delete(&models.FCMToken{})
		if result.Error != nil {
			log.Printf("[FCM] Failed to remove invalid tokens: %v", result.Error)
		} else {
			log.Printf("[FCM] Removed %d invalid tokens", result.RowsAffected)
		}
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Test notification sent", "", gin.H{
		"devices_count":    len(tokens),
		"success_count":    len(tokens) - len(invalidTokens),
		"failed_count":     len(invalidTokens),
	}))
}

// GetFirebaseStatus returns Firebase configuration status
// GET /api/fcm/status
func GetFirebaseStatus(ctx *gin.Context) {
	client := getFirebaseClient()
	enabled := os.Getenv("FCM_NOTIFICATIONS_ENABLED") == "true"
	eventTypes := os.Getenv("FCM_NOTIFY_ON_EVENT_TYPE")
	labels := os.Getenv("FCM_NOTIFY_LABELS")

	status := gin.H{
		"configured": client != nil,
		"enabled":    enabled,
	}

	if client != nil {
		status["event_types"] = strings.Split(eventTypes, ",")
		status["labels"] = strings.Split(labels, ",")
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Firebase status retrieved", "", status))
}
