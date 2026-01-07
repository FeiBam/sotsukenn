package middleware

import (
	"net/http"
	"strings"

	"sotsukenn/go/utils"

	"github.com/gin-gonic/gin"
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
