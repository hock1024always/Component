package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gomail/config"
	"gomail/models"
	"gomail/services"
	"net/http"
	"time"
)

var verificationCodes = make(map[string]string)

func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"错误": err.Error()})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if result := config.DB.Where("email = ? OR username = ?", user.Email, user.Username).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或邮箱已注册"})
		return
	}

	// 生成验证码
	code := services.GenerateVerificationCode()
	verificationCodes[user.Email] = code

	// 发送验证码
	err := services.SendMail(user.Email, "您的验证码", fmt.Sprintf("您的验证码是：%s，有效期5分钟。", code))
	fmt.Println(err)
	fmt.Println(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码发送失败"})
		return
	}

	// 设置验证码有效期
	time.AfterFunc(5*time.Minute, func() {
		delete(verificationCodes, user.Email)
	})

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送，请在5分钟内完成验证"})
}
