package handlers

import (
	"net/http"

	"chatroom/db"
	"github.com/gin-gonic/gin"
)

func SendMessage(c *gin.Context) {
	var message db.Message
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}

func GetMessages(c *gin.Context) {
	var messages []db.Message
	receiverID := c.Param("receiverID")

	if err := db.DB.Where("sender_id = ? OR receiver_id = ?", receiverID, receiverID).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
