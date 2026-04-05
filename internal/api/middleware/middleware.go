package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Auth-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// Auth 认证中间件（如果配置了 auth_key）
func Auth(authKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未配置 auth_key，跳过认证
		if authKey == "" {
			c.Next()
			return
		}

		// 检查 Header
		key := c.GetHeader("X-Auth-Key")
		if key != authKey {
			c.AbortWithStatusJSON(401, gin.H{"error": "未授权访问", "code": "UNAUTHORIZED"})
			return
		}

		c.Next()
	}
}