package controllers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gomail/config"
	"gomail/models"
	"net/http"
	"strings"
)

func Verify(c *gin.Context) {
	var data struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查验证码
	code, ok := verificationCodes[data.Email]
	if !ok || code != data.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	// 提取邮箱的用户名部分（假设用户名是邮箱的前缀）
	atIndex := strings.Index(data.Email, "@")
	if atIndex == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邮箱地址"})
		return
	}
	username := data.Email[:atIndex]

	// 注册用户
	user := models.User{
		Username: username,
		Password: "hashed_password", // 实际开发中应加密存储
		Email:    data.Email,
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user.Password = string(hashedPassword)

	// 存储用户到数据库
	if result := config.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户注册失败"})
		return
	}

	// 清理验证码
	delete(verificationCodes, data.Email)

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}
