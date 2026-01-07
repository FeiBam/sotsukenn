package utils

import (
	"fmt"
	"os"

	"sotsukenn/go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func JsonResponse(status string, code int, message string, err string, body any) gin.H {
	return gin.H{
		"status":  status,
		"code":    code,
		"error":   err,
		"message": message,
		"body":    body,
	}
}

func RegisterRoutes(relativePath string, target *gin.RouterGroup, source func(prefix string, target *gin.RouterGroup)) {
	source(relativePath, target)
}

func GetDBFromContext(ctx *gin.Context) (*gorm.DB, error) {
	value, exists := ctx.Get("db")
	if !exists {
		return nil, fmt.Errorf("database instance not found in context")
	}
	db, ok := value.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("failed to assert database instance")
	}
	return db, nil
}

func RespondWithError(ctx *gin.Context, statusCode int, message string, details interface{}) {
	ctx.JSON(statusCode, JsonResponse("err", statusCode, "", message, details))
}

func GetTokenStoreFromContext(ctx *gin.Context) *models.TokenStore {
	return ctx.MustGet("token_store").(*models.TokenStore)
}

func GenerateJWT(jwtClaims *jwt.MapClaims) (string, error) {
	var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// 断言为 jwt.MapClaims 并验证 token 是否有效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token or claims")
}
