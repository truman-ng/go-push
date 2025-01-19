package server

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine) {
	router.GET("/get/token", GetToken)
	router.GET("/heartbeat", HeartBeatHandle)
	router.GET("/push/new/score/message", PushNewScoreMessage)
}
