package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func XResponseTime(ctx *gin.Context) {
	startTime := time.Now()
	ctx.Next()

	responseTime := time.Since(startTime)
	ctx.Header("X-Response-Time", responseTime.String())
}

func SecurityHeaders(ctx *gin.Context) {
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("X-Frame-Options", "DENY")
	ctx.Header("X-XSS-Protection", "1; mode=block")
	ctx.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	ctx.Next()
}
