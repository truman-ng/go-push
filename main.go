package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-push/internal/config"
	"go-push/internal/server"
	"go-push/pkg/redis"
	"io"
	"log"
	"net/http"
	"os"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	//initConfig()
	cfg := config.LoadConfig("config/config.yaml")
	initLog()
	initRedis(cfg)
	route := gin.Default()
	server.RegisterRoutes(route)

	go server.DefaultClientManager.Start()

	go server.StartWebSocketServer(cfg.Server.WSPort)
	go server.ClientTimeoutChecker()

	log.Printf("HTTP server started on port %s", cfg.Server.HttpPort)

	route.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	if err := route.Run(":" + cfg.Server.HttpPort); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

//	func initConfig() {
//		viper.SetConfigName("config")
//		viper.AddConfigPath("config")
//		if err := viper.ReadInConfig(); err != nil {
//			panic(fmt.Errorf("fatal error config file: %s", err))
//		}
//		fmt.Println("config data: ", viper.AllSettings())
//	}
func initLog() {
	gin.DisableConsoleColor()
	logFile := viper.GetString("log.file")
	f, _ := os.Create(logFile)
	gin.DefaultWriter = io.MultiWriter(f)
}
func initRedis(cfg *config.Config) {
	redis.NewRedisClient(cfg)
}
