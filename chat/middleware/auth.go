package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Cookie中的Session Token
		sessionToken := c.GetCookie("session_token")
		if sessionToken == "" || sessionToken != "some_token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 设置超时时间（可选）
		expiration := time.Now().Add(1 * time.Hour)
		c.SetCookie("session_token", sessionToken, int(expiration.Unix()), "/", "localhost", false, true)

		c.Next()
	}
}
