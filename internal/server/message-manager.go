package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Message struct {
	Message string `json:"message" binding:"required"`
}

func PushNewScoreMessage(c *gin.Context) {
	var msg Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clientManager.Broadcast <- []byte(msg.Message)
	c.JSON(http.StatusOK, gin.H{"message": "message sent"})
}
