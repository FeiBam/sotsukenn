package routes

import (
	"sotsukenn/go/handlers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(prefix string, r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
	}
}

func UserRoutes(prefix string, r *gin.RouterGroup) {
	users := r.Group("/users")
	users.Use(handlers.AuthMiddleware())
	{
		users.GET("/profile", handlers.GetProfile)
		users.PUT("/profile", handlers.UpdateProfile)
	}
}

func HealthRoutes(prefix string, r *gin.RouterGroup) {
	r.GET("/health", handlers.HealthCheck)
}

func CameraRoutes(prefix string, r *gin.RouterGroup) {
	cameras := r.Group("/camera")
	cameras.Use(handlers.AuthMiddleware())
	{
		cameras.GET("/:name/snapshot", handlers.GetCameraSnapshot)
		cameras.GET("/:name/stream", handlers.GetCameraStream)
	}
}

func CamerasRoutes(prefix string, r *gin.RouterGroup) {
	cameras := r.Group(prefix)
	cameras.Use(handlers.AuthMiddleware())
	{
		cameras.GET("", handlers.GetCameras)
	}
}

func MqttRoutes(prefix string, r *gin.RouterGroup) {
	mqtt := r.Group("/mqtt")
	mqtt.Use(handlers.AuthMiddleware())
	{
		mqtt.POST("/start", handlers.StartMQTT)
		mqtt.POST("/stop", handlers.StopMQTT)
		mqtt.GET("/status", handlers.GetMQTTStatus)
	}
}

func FcmRoutes(prefix string, r *gin.RouterGroup) {
	fcm := r.Group("/fcm")
	fcm.Use(handlers.AuthMiddleware())
	{
		fcm.GET("/status", handlers.GetFirebaseStatus)
		fcm.POST("/test", handlers.SendTestNotification)
		fcm.GET("/tokens", handlers.GetFCMTokens)
		fcm.POST("/tokens", handlers.RegisterFCMToken)
		fcm.PUT("/tokens/:id", handlers.UpdateFCMToken)
		fcm.DELETE("/tokens/:id", handlers.DeleteFCMToken)
	}
}

func ZabbixRoutes(prefix string, r *gin.RouterGroup) {
	zabbix := r.Group("/zabbix")
	zabbix.Use(handlers.AuthMiddleware())
	{
		zabbix.GET("/all", handlers.GetZabbixAllStats)
		zabbix.GET("/status", handlers.GetZabbixStatus)
		zabbix.GET("/events/last", handlers.GetZabbixLastEvent)
		zabbix.GET("/cameras", handlers.GetZabbixCameras)
		zabbix.GET("/stats/person", handlers.GetZabbixPersonStats)
	}
}
