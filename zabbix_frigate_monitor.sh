#!/bin/bash

# Zabbix监控脚本 - Frigate监控系统
# 此脚本通过API获取Frigate监控数据供Zabbix使用

# 配置
API_URL="${API_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

# 检查JWT Token
if [ -z "$JWT_TOKEN" ]; then
    echo "Error: JWT_TOKEN environment variable is required"
    echo "Usage: JWT_TOKEN=your_token $0 [all|status|cameras|events|person]"
    exit 1
fi

# 获取所有监控指标
get_all_stats() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/all" | jq '
{
  "last_event_time": .body.last_event_time,
  "frigate_online": .body.frigate_status.is_online,
  "response_time_ms": .body.frigate_status.response_time_ms,
  "cameras_online": .body.camera_status.online_count,
  "cameras_offline": .body.camera_status.offline_count,
  "person_detections": .body.person_stats.count,
  "recognized_people": .body.person_stats.recognized | length
}'
}

# 获取Frigate状态
get_frigate_status() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/status" | jq -r '.body.is_online'
}

# 获取响应时间
get_response_time() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/status" | jq -r '.body.response_time_ms'
}

# 获取在线摄像头数量
get_cameras_online() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/cameras" | jq -r '.body.online_count'
}

# 获取离线摄像头数量
get_cameras_offline() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/cameras" | jq -r '.body.offline_count'
}

# 获取最后事件时间
get_last_event_time() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/events/last" | jq -r '.body.last_event_time'
}

# 获取人类检测总数
get_person_count() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/stats/person" | jq -r '.body.total_detections'
}

# 获取识别到的人员数量
get_recognized_people_count() {
    curl -s -H "Authorization: Bearer $JWT_TOKEN" \
        "$API_URL/api/zabbix/stats/person" | jq -r '.body.unique_count'
}

# 主函数
case "$1" in
    all)
        get_all_stats
        ;;
    status)
        get_frigate_status
        ;;
    response_time)
        get_response_time
        ;;
    cameras_online)
        get_cameras_online
        ;;
    cameras_offline)
        get_cameras_offline
        ;;
    last_event)
        get_last_event_time
        ;;
    person_count)
        get_person_count
        ;;
    recognized_count)
        get_recognized_people_count
        ;;
    *)
        echo "Usage: $0 {all|status|response_time|cameras_online|cameras_offline|last_event|person_count|recognized_count}"
        exit 1
        ;;
esac
