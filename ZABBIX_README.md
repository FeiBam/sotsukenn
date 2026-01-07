# Zabbix监控集成指南

本指南说明如何使用Zabbix监控Frigate系统。

## 功能概述

系统提供以下监控指标：

1. **最后事件时间** - 最后一次检测事件的时间戳
2. **Frigate状态** - Frigate API连接状态和响应时间
3. **在线摄像头数量** - 通过go2rtc流状态判断
4. **离线摄像头数量** - 没有producer的摄像头
5. **人类被触发次数** - 统计label="person"的事件总数
6. **哪些人** - Re-ID识别结果（sub_label字段）

## 架构说明

```
Frigate → MQTT Broker → MQTT Client → EventService → SQLite数据库
                                                  ↓
                                          Zabbix API ← Zabbix Agent
```

**优势**：
- 实时事件捕获，无需轮询Frigate API
- 降低Frigate API负载
- 事件数据持久化，支持历史查询
- 统一的API接口供Zabbix调用

## API端点

所有端点都需要JWT认证（在请求头中添加 `Authorization: Bearer <token>`）

### 统一监控端点（推荐）

**GET** `/api/zabbix/all`

返回所有监控指标：

```json
{
  "status": "success",
  "code": 200,
  "body": {
    "last_event_time": 1704556800.123,
    "frigate_status": {
      "is_online": true,
      "response_time_ms": 45,
      "last_error": "",
      "last_check_time": "2026-01-06T15:30:00Z"
    },
    "camera_status": {
      "online_count": 3,
      "offline_count": 1,
      "total_count": 4,
      "online_list": ["front_door", "backyard", "garage"],
      "offline_list": ["driveway"]
    },
    "person_stats": {
      "count": 156,
      "recognized": ["john_doe", "jane_smith", "unknown_person_1"]
    }
  }
}
```

### 单独监控端点

#### 1. Frigate状态

**GET** `/api/zabbix/status`

返回Frigate连接状态和响应时间。

#### 2. 最后事件时间

**GET** `/api/zabbix/events/last?camera=xxx&label=xxx`

- `camera`（可选）: 摄像头名称
- `label`（可选）: 检测类型（如person）

#### 3. 摄像头状态

**GET** `/api/zabbix/cameras`

返回摄像头在线/离线统计。

#### 4. 人类检测统计

**GET** `/api/zabbix/stats/person`

返回人类检测次数和识别到的人员列表。

## Zabbix集成步骤

### 步骤1：部署监控脚本

1. 复制监控脚本到服务器：
```bash
cp zabbix_frigate_monitor.sh /usr/local/bin/
chmod +x /usr/local/bin/zabbix_frigate_monitor.sh
```

2. 配置JWT Token环境变量，或直接在脚本中设置。

### 步骤2：获取JWT Token

```bash
curl -X POST http://your-server:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}'
```

响应示例：
```json
{
  "status": "success",
  "body": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 步骤3：配置Zabbix Agent

1. 复制配置文件：
```bash
cp zabbix_agentd.conf.example /etc/zabbix_agentd.conf.d/frigate.conf
```

2. 编辑配置文件，设置JWT Token：
```bash
vi /etc/zabbix_agentd.conf.d/frigate.conf
```

3. 重启Zabbix Agent：
```bash
systemctl restart zabbix-agent
```

### 步骤4：测试监控

在Zabbix Server上测试：
```bash
zabbix_get -s <agent_ip> -k "frigate.status"
zabbix_get -s <agent_ip> -k "frigate.cameras.online"
```

### 步骤5：创建Zabbix监控项

在Zabbix Web界面中创建以下监控项：

| 名称 | 键值 | 类型 | 信息类型 |
|------|------|------|----------|
| Frigate状态 | frigate.status | Zabbix agent | Numeric (unsigned) |
| Frigate响应时间 | frigate.response_time | Zabbix agent | Numeric (unsigned) |
| 在线摄像头数量 | frigate.cameras.online | Zabbix agent | Numeric (unsigned) |
| 离线摄像头数量 | frigate.cameras.offline | Zabbix agent | Numeric (unsigned) |
| 最后事件时间 | frigate.last_event | Zabbix agent | Numeric (unsigned) |
| 人类检测总数 | frigate.person.count | Zabbix agent | Numeric (unsigned) |
| 识别人员数量 | frigate.person.recognized | Zabbix agent | Numeric (unsigned) |

### 步骤6：配置触发器

示例触发器配置：

- **Frigate离线告警**：`frigate.status=0`
- **摄像头离线告警**：`frigate.cameras.offline>0`
- **响应时间告警**：`frigate.response_time>1000`

## 数据库维护

### 清理历史事件

定期清理旧事件以节省数据库空间：

```sql
-- 删除30天前的事件
DELETE FROM detection_events WHERE start_time < strftime('%s', 'now', '-30 days');

-- 或者只保留最近1000条记录
DELETE FROM detection_events WHERE id NOT IN (
  SELECT id FROM detection_events ORDER BY start_time DESC LIMIT 1000
);
```

### 查看事件统计

```sql
-- 按摄像头统计事件数量
SELECT camera, label, COUNT(*) as count
FROM detection_events
WHERE deleted_at IS NULL
GROUP BY camera, label
ORDER BY count DESC;

-- 查看识别到的人员
SELECT DISTINCT sub_label, COUNT(*) as count
FROM detection_events
WHERE label = 'person' AND sub_label IS NOT NULL AND sub_label != ''
GROUP BY sub_label
ORDER BY count DESC;

-- 查看最近的事件
SELECT camera, label, sub_label, datetime(start_time, 'unixepoch') as event_time
FROM detection_events
WHERE is_current = 1
ORDER BY start_time DESC;
```

## 故障排除

### 问题：API返回401错误

**解决方案**：检查JWT Token是否有效，重新登录获取新token。

### 问题：事件没有保存到数据库

**解决方案**：
1. 检查MQTT是否连接：`GET /api/mqtt/status`
2. 查看日志中是否有 "Failed to save detection event" 错误
3. 确认数据库迁移已运行：`./sotsukenn-server migrate db`

### 问题：Zabbix Agent无法获取数据

**解决方案**：
1. 检查脚本权限：`ls -la /usr/local/bin/zabbix_frigate_monitor.sh`
2. 手动测试脚本：`JWT_TOKEN=xxx /usr/local/bin/zabbix_frigate_monitor.sh status`
3. 查看Zabbix Agent日志：`tail -f /var/log/zabbix/zabbix_agentd.log`

## 性能优化

1. **数据库索引**：已为常用查询字段添加索引
2. **并发查询**：统一监控端点使用goroutine并发获取数据
3. **is_current标记**：快速查找最后事件，无需全表扫描

## 安全建议

1. **保护JWT Token**：不要在日志中记录token
2. **使用HTTPS**：生产环境建议使用SSL/TLS加密
3. **限制API访问**：配置防火墙规则限制访问
4. **定期更换Token**：建议定期更新JWT token

## 监控示例输出

```bash
# 测试所有监控指标
JWT_TOKEN=your_token /usr/local/bin/zabbix_frigate_monitor.sh all

# 输出示例：
{
  "last_event_time": 1704556800.123,
  "frigate_online": true,
  "response_time_ms": 45,
  "cameras_online": 3,
  "cameras_offline": 0,
  "person_detections": 156,
  "recognized_people": 3
}
```
