package services

import (
	"fmt"
	"sotsukenn/go/models"

	"gorm.io/gorm"
)

// EventService 处理事件存储和统计查询
type EventService struct {
	db *gorm.DB
}

// NewEventService 创建事件服务
func NewEventService(db *gorm.DB) *EventService {
	return &EventService{db: db}
}

// SaveDetectionEvent 保存检测事件
func (es *EventService) SaveDetectionEvent(event models.FrigateEvent) error {
	// 只处理 "new" 类型的事件
	if event.Type != models.EventTypeNew {
		return nil
	}

	// 检查事件是否已存在
	var existing models.DetectionEvent
	err := es.db.Where("event_id = ?", event.After.ID).First(&existing).Error
	if err == nil {
		// 事件已存在，跳过
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing event: %w", err)
	}

	// 解析sub_label
	subLabel := ""
	if event.After.SubLabel != nil {
		switch v := event.After.SubLabel.(type) {
		case string:
			subLabel = v
		case []interface{}:
			if len(v) > 0 {
				if str, ok := v[0].(string); ok {
					subLabel = str
				}
			}
		}
	}

	// 创建新的检测事件记录
	detectionEvent := models.DetectionEvent{
		EventID:    event.After.ID,
		Camera:     event.After.Camera,
		Label:      event.After.Label,
		SubLabel:   subLabel,
		StartTime:  event.After.StartTime,
		EndTime:    event.After.EndTime,
		TopScore:   event.After.TopScore,
		Score:      event.After.Score,
		Active:     event.After.Active,
		Stationary: event.After.Stationary,
		IsCurrent:  true,
	}

	// 开始事务
	tx := es.db.Begin()

	// 将相同摄像头和标签的旧事件的is_current设置为false
	tx.Model(&models.DetectionEvent{}).
		Where("camera = ? AND label = ? AND is_current = ?", event.After.Camera, event.After.Label, true).
		Update("is_current", false)

	// 保存新事件
	if err := tx.Create(&detectionEvent).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save detection event: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// GetLastEventTime 获取最后事件时间
func (es *EventService) GetLastEventTime(camera, label string) (float64, error) {
	var event models.DetectionEvent
	query := es.db.Where("is_current = ?", true)

	if camera != "" {
		query = query.Where("camera = ?", camera)
	}
	if label != "" {
		query = query.Where("label = ?", label)
	}

	err := query.Order("start_time DESC").First(&event).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return event.StartTime, nil
}

// GetPersonDetectionCount 获取人类检测总数和识别到的人员列表
func (es *EventService) GetPersonDetectionCount() (int64, []string, error) {
	var count int64
	var subLabels []string

	// 统计person事件总数
	if err := es.db.Model(&models.DetectionEvent{}).
		Where("label = ?", "person").
		Count(&count).Error; err != nil {
		return 0, nil, err
	}

	// 获取所有不同的sub_label (Re-ID结果)
	if err := es.db.Model(&models.DetectionEvent{}).
		Where("label = ? AND sub_label != '' AND sub_label IS NOT NULL", "person").
		Distinct("sub_label").
		Order("sub_label ASC").
		Pluck("sub_label", &subLabels).Error; err != nil {
		return 0, nil, err
	}

	return count, subLabels, nil
}

// GetEventsByTimeRange 获取指定时间范围内的事件
func (es *EventService) GetEventsByTimeRange(startTime, endTime int64) ([]models.DetectionEvent, error) {
	var events []models.DetectionEvent
	err := es.db.Where("start_time >= ? AND start_time <= ?",
		startTime, endTime).
		Order("start_time DESC").
		Find(&events).Error
	return events, err
}
