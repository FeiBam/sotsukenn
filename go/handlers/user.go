package handlers

import (
	"net/http"

	"sotsukenn/go/models"
	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty"`
}

func GetProfile(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "User not found", nil)
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Profile retrieved", "", gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}))
}

func UpdateProfile(ctx *gin.Context) {
	db, err := utils.GetDBFromContext(ctx)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Database error", nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusNotFound, "User not found", nil)
		return
	}

	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to update profile", nil)
		return
	}

	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Profile updated", "", gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"is_active":  user.IsActive,
		"updated_at": user.UpdatedAt,
	}))
}

func HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, utils.JsonResponse("success", http.StatusOK, "Server is running", "", gin.H{
		"status": "healthy",
	}))
}
