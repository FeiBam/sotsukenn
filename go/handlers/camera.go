package handlers

import (
	"net/http"
	"strings"

	"sotsukenn/go/models"
	"sotsukenn/go/services"
	"sotsukenn/go/types"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
)

// GetCameras retrieves all camera names from go2rtc streams
// GET /api/cameras
// Requires authentication (JWT token)
func GetCameras(ctx *gin.Context) {
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

	// Get user's Frigate configuration
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	// Create Frigate client
	client := services.NewFrigateClient(frigateConnect.FrigateURL)

	// Get streams using stored Frigate token (Bearer authentication)
	streams, err := client.GetGo2RTCStreamsWithToken(frigateConnect.TokenCookie)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusBadGateway, "Failed to retrieve streams from Frigate", err.Error())
		return
	}

	// Extract camera names and deduplicate (remove _WebRTC suffix)
	cameras := deduplicateCameraNames(streams)

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Cameras retrieved successfully", "", gin.H{
		"cameras": cameras,
	}))
}

// GetCameraSnapshot returns the URL for the latest snapshot from a camera
// GET /api/camera/:name/snapshot
// Requires authentication (JWT token)
func GetCameraSnapshot(ctx *gin.Context) {
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

	cameraName := ctx.Param("name")
	if cameraName == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Camera name is required", nil)
		return
	}

	// Get user's Frigate configuration
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	// Create Frigate client and get snapshot URL
	client := services.NewFrigateClient(frigateConnect.FrigateURL)
	snapshotURL := client.GetLatestSnapshotURL(cameraName)

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Snapshot URL retrieved", "", gin.H{
		"camera_name":   cameraName,
		"url":           snapshotURL,
		"frigate_token": frigateConnect.TokenCookie,
	}))
}

// GetCameraStream returns the stream URL for a specific camera
// GET /api/camera/:name/stream?type=mjpeg
// Requires authentication (JWT token)
func GetCameraStream(ctx *gin.Context) {
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

	cameraName := ctx.Param("name")
	if cameraName == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Camera name is required", nil)
		return
	}

	// Get stream type from query parameter (default: mjpeg)
	streamType := ctx.Query("type")
	if streamType == "" {
		streamType = "mjpeg"
	}

	// Currently only support mjpeg
	if streamType != "mjpeg" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Only mjpeg stream type is currently supported", nil)
		return
	}

	// Get user's Frigate configuration
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	// Create Frigate client and get stream URL
	client := services.NewFrigateClient(frigateConnect.FrigateURL)
	streamURL := client.GetMJPEGStreamURL(cameraName)

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Stream URL retrieved", "", gin.H{
		"camera_name":   cameraName,
		"stream_type":   streamType,
		"url":           streamURL,
		"frigate_token": frigateConnect.TokenCookie,
	}))
}

// normalizeCameraName removes the _WebRTC suffix from camera names
func normalizeCameraName(name string) string {
	return strings.TrimSuffix(name, "_WebRTC")
}

// deduplicateCameraNames extracts camera names from go2rtc streams and removes duplicates
// e.g., "Tp-Link" and "Tp-Link_WebRTC" are treated as the same camera
func deduplicateCameraNames(streams types.Go2RTCStreamsResponse) []string {
	seen := make(map[string]bool)
	result := []string{}

	for name := range streams {
		normalized := normalizeCameraName(name)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}
