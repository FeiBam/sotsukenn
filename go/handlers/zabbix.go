package handlers

import (
	"net/http"
	"sotsukenn/go/models"
	"sotsukenn/go/services"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
)

// GetZabbixStatus 返回Frigate状态和响应时间
// GET /api/zabbix/status
func GetZabbixStatus(ctx *gin.Context) {
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

	// 获取Frigate配置
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	monitoringSvc := services.NewMonitoringService(db, frigateConnect.FrigateURL)
	status, err := monitoringSvc.GetFrigateStatus(userID.(uint))
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get Frigate status", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Frigate status retrieved", "", status))
}

// GetZabbixLastEvent 返回最后事件时间
// GET /api/zabbix/events/last?camera=xxx&label=xxx
func GetZabbixLastEvent(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	camera := ctx.Query("camera")
	label := ctx.Query("label")

	eventSvc := services.NewEventService(db)
	lastTime, err := eventSvc.GetLastEventTime(camera, label)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get last event time", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Last event time retrieved", "", gin.H{
		"last_event_time": lastTime,
		"camera":          camera,
		"label":           label,
	}))
}

// GetZabbixCameras 返回摄像头在线/离线统计
// GET /api/zabbix/cameras
func GetZabbixCameras(ctx *gin.Context) {
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

	// 获取Frigate配置
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	monitoringSvc := services.NewMonitoringService(db, frigateConnect.FrigateURL)
	status, err := monitoringSvc.GetCameraStatus(userID.(uint))
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get camera status", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Camera status retrieved", "", status))
}

// GetZabbixPersonStats 返回人类检测统计
// GET /api/zabbix/stats/person
func GetZabbixPersonStats(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	eventSvc := services.NewEventService(db)
	count, people, err := eventSvc.GetPersonDetectionCount()
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get person stats", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Person detection stats retrieved", "", gin.H{
		"total_detections":  count,
		"recognized_people": people,
		"unique_count":      len(people),
	}))
}

// GetZabbixAllStats 返回所有监控指标的统一端点
// GET /api/zabbix/all
func GetZabbixAllStats(ctx *gin.Context) {
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

	// 获取Frigate配置
	var frigateConnect models.FrigateConnect
	err = db.Where("user_id = ? AND is_active = ?", userID, true).First(&frigateConnect).Error
	if err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "Frigate configuration not found", nil)
		return
	}

	// 初始化服务
	eventSvc := services.NewEventService(db)
	monitoringSvc := services.NewMonitoringService(db, frigateConnect.FrigateURL)

	// 1. 最后事件时间
	lastTime, err := eventSvc.GetLastEventTime("", "")
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get last event time", err.Error())
		return
	}

	// 2. Frigate状态
	frigateStatus, err := monitoringSvc.GetFrigateStatus(userID.(uint))
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get Frigate status", err.Error())
		return
	}

	// 3. 摄像头状态
	cameraStatus, err := monitoringSvc.GetCameraStatus(userID.(uint))
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get camera status", err.Error())
		return
	}

	// 4. 人类检测统计
	personCount, people, err := eventSvc.GetPersonDetectionCount()
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to get person stats", err.Error())
		return
	}

	// 组装所有统计数据
	allStats := gin.H{
		"last_event_time": lastTime,
		"frigate_status":  frigateStatus,
		"camera_status":   cameraStatus,
		"person_stats": gin.H{
			"count":      personCount,
			"recognized": people,
		},
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Zabbix stats retrieved", "", allStats))
}
