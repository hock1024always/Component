package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"online_meeting/helper"
)

// Auth函数用于验证用户身份
func Auth() gin.HandlerFunc {
	// 返回一个gin.HandlerFunc类型的函数
	return func(c *gin.Context) {
		// 获取请求头中的Authorization字段
		auth := c.GetHeader("Authorization")
		// 解析Authorization字段中的token
		userClaim, err := helper.AnalyseToken(auth)
		// 如果解析失败
		if err != nil {
			// 终止请求
			c.Abort()
			// 返回错误信息
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Authorization",
			})
			// 返回
			return
		}
		// 如果解析的token为空
		if userClaim == nil {
			// 终止请求
			c.Abort()
			// 返回错误信息
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Admin",
			})
			// 返回
			return
		}
		// 将解析的token存入上下文中
		c.Set("user_claims", userClaim)
		// 继续处理请求
		c.Next()
	}
}
