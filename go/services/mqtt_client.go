package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"sotsukenn/go/models"
	"sotsukenn/go/types"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient handles MQTT connections and subscriptions
type MQTTClient struct {
	client              mqtt.Client
	brokerURL           string
	brokerPort          string
	clientID            string
	username            string
	password            string
	topic               string
	connected           bool
	mu                  sync.RWMutex
	onMessage           func(models.FrigateEvent)
	notificationService *NotificationService
	eventService        *EventService // New: Event service for persisting events
}

// NewMQTTClient creates a new MQTT client
func NewMQTTClient(config types.MQTTConfig) *MQTTClient {
	// Set default values
	if config.BrokerURL == "" {
		config.BrokerURL = os.Getenv("MQTT_BROKER_URL")
		if config.BrokerURL == "" {
			config.BrokerURL = "localhost"
		}
	}

	if config.BrokerPort == "" {
		config.BrokerPort = os.Getenv("MQTT_BROKER_PORT")
		if config.BrokerPort == "" {
			config.BrokerPort = "1883"
		}
	}

	if config.ClientID == "" {
		config.ClientID = os.Getenv("MQTT_CLIENT_ID")
		if config.ClientID == "" {
			config.ClientID = "sotsukenn-server"
		}
	}

	if config.Topic == "" {
		config.Topic = os.Getenv("MQTT_TOPIC")
		if config.Topic == "" {
			config.Topic = "frigate/events"
		}
	}

	broker := fmt.Sprintf("tcp://%s:%s", config.BrokerURL, config.BrokerPort)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(os.Getenv("MQTT_USERNAME"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	// Set connection handler
	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("MQTT: Connected to broker %s", broker)
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("MQTT: Connection lost: %v", err)
	}

	client := &MQTTClient{
		client:     mqtt.NewClient(opts),
		brokerURL:  config.BrokerURL,
		brokerPort: config.BrokerPort,
		clientID:   config.ClientID,
		topic:      config.Topic,
		connected:  false,
	}

	// Set default message handler
	client.onMessage = func(event models.FrigateEvent) {
		// Default: log camera and label
		camera := event.After.Camera
		label := event.After.Label
		eventType := event.Type
		log.Printf("[Frigate Event] Type: %s, Camera: %s, Label: %s", eventType, camera, label)
	}

	return client
}

// Connect connects to the MQTT broker
func (mc *MQTTClient) Connect() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.connected {
		return fmt.Errorf("already connected")
	}

	token := mc.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect: %w", token.Error())
	}

	mc.connected = true
	log.Printf("MQTT: Connected to %s:%s as %s", mc.brokerURL, mc.brokerPort, mc.clientID)
	return nil
}

// Disconnect disconnects from the MQTT broker
func (mc *MQTTClient) Disconnect() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.connected {
		return
	}

	mc.client.Disconnect(250)
	mc.connected = false
	log.Println("MQTT: Disconnected from broker")
}

// Subscribe subscribes to the configured topic
func (mc *MQTTClient) Subscribe() error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if !mc.connected {
		return fmt.Errorf("not connected")
	}

	topic := mc.topic

	token := mc.client.Subscribe(topic, 0, mc.messageHandler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe: %w", token.Error())
	}

	log.Printf("MQTT: Subscribed to topic: %s", topic)
	return nil
}

// SetMessageHandler sets a custom message handler
func (mc *MQTTClient) SetMessageHandler(handler func(models.FrigateEvent)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.onMessage = handler
}

// SetNotificationService sets the notification service for sending FCM notifications
func (mc *MQTTClient) SetNotificationService(ns *NotificationService) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.notificationService = ns
}

// SetEventService sets the event service for persisting detection events
func (mc *MQTTClient) SetEventService(es *EventService) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.eventService = es
}

// messageHandler handles incoming MQTT messages
func (mc *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()

	// Parse JSON payload
	var event models.FrigateEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("MQTT: Failed to parse message: %v", err)
		log.Printf("MQTT: Raw payload: %s", string(payload))
		return
	}

	// Call the message handler (with read lock for safety)
	mc.mu.RLock()
	handler := mc.onMessage
	notificationSvc := mc.notificationService
	eventSvc := mc.eventService
	mc.mu.RUnlock()

	// Save event to database if event service is configured
	if eventSvc != nil {
		if err := eventSvc.SaveDetectionEvent(event); err != nil {
			log.Printf("MQTT: Failed to save detection event: %v", err)
		}
	}

	if handler != nil {
		handler(event)
	}

	// Send FCM notification if service is configured
	if notificationSvc != nil {
		if err := notificationSvc.SendNotification(event); err != nil {
			log.Printf("MQTT: Failed to send FCM notification: %v", err)
		}
	}
}

// IsConnected returns the connection status
func (mc *MQTTClient) IsConnected() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.connected
}

// GetStatus returns the current status information
func (mc *MQTTClient) GetStatus() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"connected": mc.connected,
		"broker":    fmt.Sprintf("%s:%s", mc.brokerURL, mc.brokerPort),
		"client_id": mc.clientID,
		"topic":     mc.topic,
	}
}
