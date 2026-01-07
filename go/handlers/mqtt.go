package handlers

import (
	"net/http"
	"sync"

	"sotsukenn/go/services"
	"sotsukenn/go/types"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
)

var (
	mqttClient *services.MQTTClient
	mqttOnce   sync.Once
	mqttMutex  sync.RWMutex
)

// initMQTTClient initializes the MQTT client singleton
func initMQTTClient() {
	config := types.MQTTConfig{}
	mqttClient = services.NewMQTTClient(config)
}

// getMQTTClient returns the MQTT client instance (lazy initialization)
func getMQTTClient() *services.MQTTClient {
	mqttMutex.Lock()
	defer mqttMutex.Unlock()

	if mqttClient == nil {
		mqttOnce.Do(initMQTTClient)
	}

	return mqttClient
}

// GetMQTTClient returns the MQTT client instance (exported for use in main.go)
func GetMQTTClient() *services.MQTTClient {
	return getMQTTClient()
}

// StartMQTT starts the MQTT connection and subscription
// POST /api/mqtt/start
func StartMQTT(ctx *gin.Context) {
	client := getMQTTClient()

	// Check if already connected
	if client.IsConnected() {
		utils.RespondWithError(ctx, http.StatusConflict, "MQTT client is already connected", nil)
		return
	}

	// Connect to broker
	if err := client.Connect(); err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to connect to MQTT broker", err.Error())
		return
	}

	// Subscribe to topic
	if err := client.Subscribe(); err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to subscribe to MQTT topic", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "MQTT client started and subscribed", "", gin.H{
		"status":    "connected",
		"connected": true,
	}))
}

// StopMQTT stops the MQTT connection
// POST /api/mqtt/stop
func StopMQTT(ctx *gin.Context) {
	client := getMQTTClient()

	if !client.IsConnected() {
		utils.RespondWithError(ctx, http.StatusConflict, "MQTT client is not connected", nil)
		return
	}

	client.Disconnect()

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "MQTT client stopped", "", gin.H{
		"status":    "disconnected",
		"connected": false,
	}))
}

// GetMQTTStatus returns the current MQTT client status
// GET /api/mqtt/status
func GetMQTTStatus(ctx *gin.Context) {
	client := getMQTTClient()
	status := client.GetStatus()

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "MQTT status retrieved", "", status))
}
