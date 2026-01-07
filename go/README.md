# Sotsukenn Go Server

基于 Gin 框架的 Go 服务器，包含用户认证、数据库操作等功能。使用 Cobra CLI 工具进行管理。

## 功能特性

- ✅ 用户注册/登录/登出
- ✅ JWT 身份验证
- ✅ GORM 数据库操作
- ✅ 中间件支持（CORS、认证、安全头）
- ✅ RESTful API 设计
- ✅ 统一 JSON 响应格式
- ✅ CLI 命令行工具（Cobra）
- ✅ 数据库迁移命令
- ✅ Frigate 摄像头流转发
- ✅ MQTT 事件监听（Frigate events）

## 技术栈

- **框架**: Gin Web Framework
- **CLI**: Cobra
- **ORM**: GORM
- **数据库**: SQLite
- **认证**: JWT (golang-jwt/jwt)
- **密码加密**: bcrypt
- **配置**: godotenv
- **MQTT**: Eclipse Paho MQTT

## 项目结构

```
.
├── main.go           # CLI 主入口文件
├── database/         # 数据库配置
│   └── database.go
├── models/           # 数据模型
│   ├── user.go
│   └── token.go
├── handlers/         # 路由处理器
│   ├── auth.go
│   └── user.go
├── middleware/       # 中间件
│   ├── auth.go
│   ├── cors.go
│   └── common.go
├── routes/           # 路由定义
│   └── routes.go
├── migrate/          # 数据库迁移
│   └── migrate.go
└── utils/            # 工具函数
    └── utils.go
```

## 快速开始

### 1. 配置环境变量

复制 `.env.example` 为 `.env` 并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置 JWT 密钥（可选修改数据库文件路径）。

### 2. 安装依赖

```bash
go mod download
```

### 3. 数据库迁移

运行数据库迁移命令创建表结构：

```bash
go run main.go migrate db
```

或者使用编译后的二进制文件：

```bash
./sotsukenn-server migrate db
```

### 4. 运行服务器

使用默认端口（:8080）：

```bash
go run main.go run server
```

或使用二进制文件：

```bash
./sotsukenn-server run server
```

指定端口：

```bash
./sotsukenn-server run server -p :3000
```

服务器将在指定端口启动。

## CLI 命令

### 查看帮助

```bash
./sotsukenn-server --help
```

### 运行服务器

```bash
./sotsukenn-server run server [flags]

Flags:
  -p, --port string   Port to run the server on (default ":8080")
```

### 数据库迁移

```bash
# 迁移模型到数据库
./sotsukenn-server migrate db

# 迁移 Markdown 文件（功能开发中）
./sotsukenn-server migrate md -p /path/to/markdown --force --update
```

## API 端点

### 健康检查

```
GET /api/health
```

### 认证

```
POST /api/auth/register  # 用户注册
POST /api/auth/login      # 用户登录
POST /api/auth/logout     # 用户登出（需要认证）
```

### 用户（需要认证）

```
GET  /api/users/profile  # 获取用户信息
PUT  /api/users/profile  # 更新用户信息
```

### 摄像头流（需要认证）

```
GET /api/camera/streams             # 获取所有摄像头流
GET /api/camera/streams/:name/url   # 获取指定摄像头的流 URL
```

**摄像头流参数说明：**

- `GET /api/camera/streams/:name?url?format=mp4`
  - `name`: 摄像头名称（路径参数）
  - `format`: 流格式（查询参数，可选，默认为 `mp4`）
    - 支持的格式：`mp4`、`mjpg`、`webrtc`、`rtsp`

### MQTT 服务（需要认证）

```
POST /api/mqtt/start   # 启动 MQTT 连接
POST /api/mqtt/stop    # 停止 MQTT 连接
GET  /api/mqtt/status  # 获取 MQTT 状态
```

**MQTT 服务说明：**

- 连接到 Frigate MQTT broker 订阅 `frigate/events` 主题
- 接收事件后自动输出 `camera` 和 `label` 到日志
- 支持的事件类型：`new`（新建）、`update`（更新）、`end`（结束）
- **自动启动**：通过环境变量 `MQTT_AUTO_START=true` 可在服务器启动时自动连接 MQTT
- **手动控制**：即使设置了自动启动，仍可通过 API 随时停止或重新启动

## API 使用示例

### 注册用户

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 登录

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 获取用户信息（需要 token）

```bash
curl -X GET http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 获取所有摄像头流

```bash
curl -X GET http://localhost:8080/api/camera/streams \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**响应示例：**

```json
{
  "status": "success",
  "code": 200,
  "error": "",
  "message": "Streams retrieved successfully",
  "body": {
    "Tp-Link": {
      "producers": [...],
      "consumers": [...]
    },
    "Tp-Link_WebRTC": {
      "producers": [...],
      "consumers": null
    }
  }
}
```

### 获取指定摄像头的流 URL

```bash
# 获取 MP4 格式流（默认）
curl -X GET "http://localhost:8080/api/camera/streams/Tp-Link/url" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 获取 MJPEG 格式流
curl -X GET "http://localhost:8080/api/camera/streams/Tp-Link/url?format=mjpg" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 获取 WebRTC 格式流
curl -X GET "http://localhost:8080/api/camera/streams/Tp-Link/url?format=webrtc" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**响应示例：**

```json
{
  "status": "success",
  "code": 200,
  "error": "",
  "message": "Stream URL generated",
  "body": {
    "stream_name": "Tp-Link",
    "format": "mp4",
    "url": "https://frigate.example.com/api/go2rtc/Tp-Link.mp4?token=xxxxx"
  }
}
```

**支持的流格式：**

- `mp4` - H.264 视频流（推荐用于大多数播放器）
- `mjpg` - MJPEG 视频流（适用于浏览器直接播放）
- `webrtc` - WebRTC 流（低延迟实时流）
- `rtsp` - RTSP 流（适用于专业播放器）

### 启动 MQTT 连接

```bash
curl -X POST http://localhost:8080/api/mqtt/start \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**响应示例：**

```json
{
  "status": "success",
  "code": 200,
  "error": "",
  "message": "MQTT client started and subscribed",
  "body": {
    "status": "connected",
    "connected": true
  }
}
```

### 获取 MQTT 状态

```bash
curl -X GET http://localhost:8080/api/mqtt/status \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**响应示例：**

```json
{
  "status": "success",
  "code": 200,
  "error": "",
  "message": "MQTT status retrieved",
  "body": {
    "connected": true,
    "broker": "localhost:1883",
    "client_id": "sotsukenn-server",
    "topic": "frigate/events"
  }
}
```

### 停止 MQTT 连接

```bash
curl -X POST http://localhost:8080/api/mqtt/stop \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**事件日志输出示例：**

当 MQTT 收到 Frigate 事件时，会在日志中输出：

```
[Frigate Event] Type: new, Camera: front_door, Label: person
[Frigate Event] Type: update, Camera: front_door, Label: person
[Frigate Event] Type: end, Camera: front_door, Label: person
```

### MQTT 自动启动配置

在 `.env` 文件中设置 `MQTT_AUTO_START` 来控制 MQTT 服务的启动行为：

```bash
# .env 文件
MQTT_AUTO_START=true   # 服务器启动时自动连接 MQTT
# 或
MQTT_AUTO_START=false  # 需要通过 API 手动启动（默认）
```

**MQTT_AUTO_START=true 时：**

```
Running server on port :8080
MQTT: Auto-start enabled, connecting to broker...
MQTT: Connected to tcp://localhost:1883
MQTT: Subscribed to topic: frigate/events
MQTT: Auto-start successful
[GIN] Listening and serving HTTP on :8080
```

**MQTT_AUTO_START=false 时（默认）：**

```
Running server on port :8080
MQTT: Auto-start disabled, use API to start manually
[GIN] Listening and serving HTTP on :8080
```

**注意**：无论 `MQTT_AUTO_START` 设置如何，都可以通过 API 随时控制 MQTT 连接状态。

## 响应格式

所有 API 响应遵循统一格式：

```json
{
  "status": "success",
  "code": 200,
  "error": "",
  "message": "操作成功",
  "body": {}
}
```

## 开发说明

### 添加新的路由

1. 在 `handlers/` 目录创建或编辑处理器文件
2. 在 `main.go` 中注册路由
3. 如需认证，添加 `middleware.AuthMiddleware()`

### 数据库迁移

服务器启动时会自动执行数据库迁移，创建必要的表。

## License

MIT

## Firebase FCM 推送通知

### 功能说明

集成 Google Firebase Cloud Messaging (FCM)，当 MQTT 接收到 Frigate 事件时自动发送推送到手机 APP：

- **人脸识别通知**：检测到人脸时，通知显示具体人名
- **普通检测通知**：只检测到人时，显示"检测到：person"
- **事件类型过滤**：只在事件开始（new）和结束（end）时发送
- **去重机制**：30 秒内相同事件只发送一次

### API 端点

```
POST /api/fcm/tokens        # 注册设备 token
GET  /api/fcm/tokens        # 获取设备列表
PUT  /api/fcm/tokens/:id    # 更新设备信息
DELETE /api/fcm/tokens/:id  # 删除设备 token
POST /api/fcm/test          # 发送测试通知
GET  /api/fcm/status        # 获取 FCM 状态
```

### 配置步骤

1. 创建 Firebase 项目并生成服务账号密钥
2. 将密钥文件保存为 `firebase-service-account.json`（项目根目录）
3. 在 `.env` 中配置 `FIREBASE_PROJECT_ID` 和相关参数

完整配置示例请参考项目内的文档。

