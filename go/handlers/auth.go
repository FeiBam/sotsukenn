package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"sotsukenn/go/models"
	"sotsukenn/go/services"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondWithError(ctx, http.StatusUnauthorized, "Authorization header required", nil)
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.RespondWithError(ctx, http.StatusUnauthorized, "Invalid authorization format", nil)
			ctx.Abort()
			return
		}

		_, err := utils.ValidateJWT(tokenString)
		if err != nil {
			utils.RespondWithError(ctx, http.StatusUnauthorized, "Invalid token", err.Error())
			ctx.Abort()
			return
		}

		tokenStore := utils.GetTokenStoreFromContext(ctx)
		tokenInfo, exists := tokenStore.Get(tokenString)
		if !exists {
			utils.RespondWithError(ctx, http.StatusUnauthorized, "Token not found or expired", nil)
			ctx.Abort()
			return
		}

		ctx.Set("user_id", tokenInfo.UserID)
		ctx.Set("token", tokenString)
		ctx.Next()
	}
}

type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	FrigateURL string `json:"frigate_url,omitempty"` // Required for new Frigate-only users
}

func Login(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Scenario A: Local user exists
	var user models.User
	err = db.Where("username = ?", req.Username).First(&user).Error

	if err == nil {
		// User found locally - proceed with Scenario A
		handleLocalUserLogin(ctx, db, &user, &req)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Scenario B: User not found locally
		// Determine Frigate URL from request or environment
		frigateURL := req.FrigateURL
		if frigateURL == "" {
			frigateURL = os.Getenv("FRIGATE_URL")
		}

		if frigateURL != "" {
			// Update req.FrigateURL directly (it's fine since we own this request)
			req.FrigateURL = frigateURL
			handleFrigateOnlyLogin(ctx, db, &req)
			return
		}

		utils.RespondWithError(ctx, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	// Other database errors
	utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
}

// handleLocalUserLogin handles authentication for existing local users
func handleLocalUserLogin(ctx *gin.Context, db *gorm.DB, user *models.User, req *LoginRequest) {
	// Step 1: Verify local password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	if !user.IsActive {
		utils.RespondWithError(ctx, http.StatusForbidden, "Account is inactive", nil)
		return
	}

	// Step 2: Check for Frigate integration
	var frigateConnect models.FrigateConnect
	err := db.Where("user_id = ? AND is_active = ?", user.ID, true).First(&frigateConnect).Error

	if err == nil {
		// Frigate integration exists - verify/refresh token
		verifyOrRefreshFrigateToken(ctx, db, user, &frigateConnect, req.Password)
	}

	// Step 3: Generate local JWT and respond
	generateLocalToken(ctx, user)
}

// verifyOrRefreshFrigateToken verifies and refreshes Frigate tokens if needed
func verifyOrRefreshFrigateToken(ctx *gin.Context, db *gorm.DB, user *models.User, fc *models.FrigateConnect, password string) {
	client := services.NewFrigateClient(fc.FrigateURL)

	// If recently verified, skip verification
	if !fc.ShouldVerifyToken() {
		return
	}

	// Verify token
	valid, err := client.VerifyToken(fc.TokenCookie)
	if err == nil && valid {
		// Token valid, update last verified time
		now := time.Now()
		fc.LastVerifiedAt = &now
		db.Save(fc)
		return
	}

	// Token invalid or expired, refresh it
	fmt.Printf("Refreshing Frigate token for user %d\n", user.ID)

	username := fc.FrigateUsername
	if username == "" {
		username = user.Username
	}

	newToken, err := client.Login(username, password)
	if err != nil {
		fmt.Printf("Failed to refresh Frigate token: %v\n", err)
		return
	}

	// Update database
	now := time.Now()
	fc.TokenCookie = newToken
	fc.LastVerifiedAt = &now
	db.Save(fc)

	fmt.Printf("Frigate token refreshed for user %d\n", user.ID)
}

// handleFrigateOnlyLogin handles authentication for Frigate-only users
func handleFrigateOnlyLogin(ctx *gin.Context, db *gorm.DB, req *LoginRequest) {
	client := services.NewFrigateClient(req.FrigateURL)

	// Step 1: Authenticate with Frigate API
	token, err := client.Login(req.Username, req.Password)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "Frigate authentication failed", err.Error())
		return
	}

	// Step 2: Create local user record
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to hash password", nil)
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    "", // Optional, leave empty
		Password: string(hashedPassword),
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to create user", nil)
		return
	}

	// Step 3: Create FrigateConnect mapping
	now := time.Now()
	frigateConnect := models.FrigateConnect{
		UserID:          user.ID,
		FrigateURL:      req.FrigateURL,
		FrigateUsername: req.Username,
		TokenCookie:     token,
		LastVerifiedAt:  &now,
		IsActive:        true,
	}

	if err := db.Create(&frigateConnect).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to create Frigate integration", nil)
		return
	}

	// Step 4: Generate local JWT and respond
	generateLocalToken(ctx, &user)
}

// generateLocalToken generates JWT token and returns success response
func generateLocalToken(ctx *gin.Context, user *models.User) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	tokenString, err := utils.GenerateJWT(&claims)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to generate token", nil)
		return
	}

	tokenStore := utils.GetTokenStoreFromContext(ctx)
	tokenStore.Set(tokenString, &models.TokenInfo{
		Token:     tokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24),
		UserID:    user.ID,
	})

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Login successful", "", gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}))
}

func Logout(ctx *gin.Context) {
	token, exists := ctx.Get("token")
	if !exists {
		utils.RespondWithError(ctx, http.StatusBadRequest, "No token found", nil)
		return
	}

	tokenStore := utils.GetTokenStoreFromContext(ctx)
	tokenStore.Delete(token.(string))

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Logout successful", "", nil))
}
