package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cors() 函数返回一个 gin.HandlerFunc，用于处理跨域请求
func Cors() gin.HandlerFunc {
	// 返回一个匿名函数，该函数接收一个 *gin.Context 参数
	return func(c *gin.Context) {
		// 获取请求的方法
		method := c.Request.Method
		// 设置允许跨域请求的源
		c.Header("Access-Control-Allow-Origin", "*")
		// 设置允许跨域请求的方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		// 设置允许跨域请求的头部
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, AccessToken")
		// 设置允许跨域请求暴露的头部
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		// 设置允许跨域请求携带凭证
		c.Header("Access-Control-Allow-Credentials", "true")
		// 如果请求的方法是 OPTIONS，则返回 http.StatusNoContent 状态码
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 继续处理请求
		c.Next()
	}
}
