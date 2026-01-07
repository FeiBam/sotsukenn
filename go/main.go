package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"sotsukenn/go/database"
	"sotsukenn/go/handlers"
	"sotsukenn/go/middleware"
	"sotsukenn/go/migrate"
	"sotsukenn/go/models"
	"sotsukenn/go/routes"
	"sotsukenn/go/services"
	"sotsukenn/go/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gorm.io/gorm/logger"
)

var (
	tokenStore *models.TokenStore
	once       sync.Once
)

func initTokenStore() {
	tokenStore = models.NewTokenStore()
	go func() {
		for {
			time.Sleep(60 * time.Second)
			tokenStore.Clean()
		}
	}()
}

func getTokenStore() *models.TokenStore {
	once.Do(initTokenStore)
	return tokenStore
}

func addTokenStore(ctx *gin.Context) {
	ctx.Set("token_store", getTokenStore())
	ctx.Next()
}

func addDB(ctx *gin.Context) {
	db, err := database.GetDBWithLogger(logger.Silent)
	if err != nil {
		panic(fmt.Sprintf("fail get databases.... error : %v", err))
	}
	ctx.Set("db", db)
	ctx.Next()
}

func runServer() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		if port == "" {
			port = ":8080"
		}

		fmt.Printf("Running server on port %s\n", port)

		r := gin.Default()

		r.TrustedPlatform = gin.PlatformCloudflare

		r.Use(gin.Logger())
		r.Use(gin.Recovery())
		r.Use(middleware.XResponseTime)
		r.Use(middleware.SecurityHeaders)
		r.Use(addDB)
		r.Use(addTokenStore)

		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		}))

		// Default routes
		defaultRoute := r.Group("")
		{
			defaultRoute.Any("/teapot", func(ctx *gin.Context) {
				ctx.Redirect(http.StatusTemporaryRedirect, "https://www.google.com/teapot")
			})
		}

		// API routes
		api := r.Group("/api")
		{
			utils.RegisterRoutes("/auth", api, routes.AuthRoutes)
			utils.RegisterRoutes("/users", api, routes.UserRoutes)
			utils.RegisterRoutes("/health", api, routes.HealthRoutes)
			utils.RegisterRoutes("/cameras", api, routes.CamerasRoutes)
			utils.RegisterRoutes("", api, routes.CameraRoutes)
			utils.RegisterRoutes("", api, routes.MqttRoutes)
			utils.RegisterRoutes("", api, routes.FcmRoutes)
			utils.RegisterRoutes("", api, routes.ZabbixRoutes)
		}

		// Auto-start MQTT if enabled
		if os.Getenv("MQTT_AUTO_START") == "true" {
			log.Println("MQTT: Auto-start enabled, connecting to broker...")

			// Initialize database for notification service
			db, err := database.GetDBWithLogger(logger.Silent)
			if err != nil {
				log.Printf("Failed to initialize database for notifications: %v", err)
			} else {
				// Initialize Firebase client
				var firebaseClient *services.FirebaseClient
				if os.Getenv("FIREBASE_PROJECT_ID") != "" {
					firebaseClient, err = services.NewFirebaseClient()
					if err != nil {
						log.Printf("Failed to initialize Firebase: %v", err)
						log.Println("FCM notifications will be disabled")
					} else {
						log.Println("Firebase: Client initialized for FCM notifications")
					}
				}

				// Initialize notification service
				notificationService := services.NewNotificationService(db, firebaseClient)

				// Initialize event service for Zabbix monitoring
				eventService := services.NewEventService(db)
				log.Println("Event Service: Initialized for detection event persistence")

				// Get MQTT client and set services
				client := handlers.GetMQTTClient()
				client.SetNotificationService(notificationService)
				client.SetEventService(eventService)

				// Connect MQTT
				if err := client.Connect(); err != nil {
					log.Printf("MQTT: Failed to auto-start: %v", err)
					log.Println("MQTT: Will retry via API if broker becomes available")
				} else {
					if err := client.Subscribe(); err != nil {
						log.Printf("MQTT: Failed to subscribe: %v", err)
					} else {
						log.Println("MQTT: Auto-start successful")
						if firebaseClient != nil {
							log.Println("FCM: Notifications enabled for Frigate events")
						}
						log.Println("Event Persistence: Detection events will be saved to database")
					}
				}
			}
		} else {
			log.Println("MQTT: Auto-start disabled, use API to start manually")
		}

		if err := r.Run(port); err != nil {
			panic(fmt.Sprintf("failed to start server: %v", err))
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	var rootCmd = &cobra.Command{
		Use:   "sotsukenn-server",
		Short: "Sotsukenn Go Server",
		Long:  "A powerful REST API server built with Gin and GORM",
	}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run commands",
	}

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run the server",
		Run:   runServer(),
	}
	serverCmd.Flags().StringP("port", "p", ":8080", "Port to run the server on")

	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate model or markdown to databases",
	}

	var migrateModelCmd = migrate.MigrateModelCmd()
	var migrateMarkdownCmd = migrate.MigrateMarkdownCmd()

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(migrateCmd)

	runCmd.AddCommand(serverCmd)

	migrateCmd.AddCommand(migrateMarkdownCmd)
	migrateCmd.AddCommand(migrateModelCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
