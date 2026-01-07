package services

import (
	"fmt"
	"sotsukenn/go/models"
	"sotsukenn/go/types"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MonitoringService 处理Frigate健康检查和摄像头状态
type MonitoringService struct {
	db            *gorm.DB
	frigateClient *FrigateClient
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(db *gorm.DB, baseURL string) *MonitoringService {
	return &MonitoringService{
		db:            db,
		frigateClient: NewFrigateClient(baseURL),
	}
}

// GetFrigateStatus 获取Frigate状态
func (ms *MonitoringService) GetFrigateStatus(userID uint) (*types.FrigateStatus, error) {
	// 获取用户的Frigate配置
	var frigateConnect models.FrigateConnect
	err := ms.db.Where("user_id = ? AND is_active = ?", userID, true).
		First(&frigateConnect).Error
	if err != nil {
		return nil, fmt.Errorf("frigate configuration not found: %w", err)
	}

	status := &types.FrigateStatus{
		LastCheckTime: time.Now(),
	}

	// 测试连接并测量响应时间
	start := time.Now()

	// 使用VerifyToken检查连接
	tokenValid, err := ms.frigateClient.VerifyToken(frigateConnect.TokenCookie)

	elapsed := time.Since(start)
	status.ResponseTime = elapsed.Milliseconds()

	if err != nil {
		status.IsOnline = false
		status.LastError = err.Error()
		return status, nil
	}

	status.IsOnline = tokenValid

	// 如果token无效
	if !tokenValid {
		status.LastError = "token invalid"
	}

	return status, nil
}

// GetCameraStatus 获取摄像头在线状态
func (ms *MonitoringService) GetCameraStatus(userID uint) (*types.CameraStatus, error) {
	// 获取用户的Frigate配置
	var frigateConnect models.FrigateConnect
	err := ms.db.Where("user_id = ? AND is_active = ?", userID, true).
		First(&frigateConnect).Error
	if err != nil {
		return nil, fmt.Errorf("frigate configuration not found: %w", err)
	}

	// 获取go2rtc流信息
	streams, err := ms.frigateClient.GetGo2RTCStreamsWithToken(frigateConnect.TokenCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams: %w", err)
	}

	status := &types.CameraStatus{
		OnlineList:  make([]string, 0),
		OfflineList: make([]string, 0),
	}

	// 用于去重的map
	seenCameras := make(map[string]bool)

	// 分析每个流的状态
	for streamName, streamData := range streams {
		// 移除_WebRTC后缀
		cameraName := streamName
		if strings.HasSuffix(streamName, "_WebRTC") {
			cameraName = streamName[:len(streamName)-7]
		}

		// 如果已经处理过这个摄像头，跳过
		if seenCameras[cameraName] {
			continue
		}
		seenCameras[cameraName] = true

		// 检查是否有producer
		if len(streamData.Producers) > 0 {
			// 摄像头在线
			status.OnlineCount++
			status.OnlineList = append(status.OnlineList, cameraName)
		} else {
			// 摄像头离线
			status.OfflineCount++
			status.OfflineList = append(status.OfflineList, cameraName)
		}
	}

	status.TotalCount = status.OnlineCount + status.OfflineCount

	return status, nil
}
