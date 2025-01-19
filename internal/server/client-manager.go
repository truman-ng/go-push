package server

import (
	"github.com/gin-gonic/gin"
	"go-push/common/utils"
	"go-push/internal/models"
	"go-push/pkg/redis"
	"log"
	"net/http"
	"sync"
	"time"
)

type ClientManager struct {
	Clients    map[string]*models.Client
	Register   chan *models.Client
	Unregister chan *models.Client
	HeartBeat  chan *models.Client
	Broadcast  chan []byte
	mutex      sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:    make(map[string]*models.Client),
		Register:   make(chan *models.Client),
		Unregister: make(chan *models.Client),
		HeartBeat:  make(chan *models.Client),
		Broadcast:  make(chan []byte),
	}
}
func (cm *ClientManager) Start() {
	for {
		select {
		case client := <-cm.Register:
			cm.mutex.Lock()
			cm.Clients[client.Token] = client
			cm.mutex.Unlock()
			log.Println("client register success, client:", client)
		case client := <-cm.HeartBeat:
			cm.mutex.Lock()
			client.UpdatePingTime()
			cm.mutex.Unlock()
		case client := <-cm.Unregister:
			cm.mutex.Lock()
			if _, ok := cm.Clients[client.Token]; ok {
				close(client.Send)
				delete(cm.Clients, client.Token)
			}
			cm.mutex.Unlock()
			log.Println("client unregister success, client:", client)
		case message := <-cm.Broadcast:
			cm.mutex.RLock()
			for _, client := range cm.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(cm.Clients, client.Token)
				}
				log.Println("client broadcast success, client:", client)
			}
			cm.mutex.RUnlock()
		}
	}
}
func RemoveStaleClient() {
	clientManager.mutex.RLock()
	defer clientManager.mutex.RUnlock()
	for key, client := range clientManager.Clients {
		now := time.Now()
		client.Lock.Lock()
		timeSinceLastPing := now.Sub(client.LastPingTime)
		client.Lock.Unlock()

		if timeSinceLastPing > 30*time.Second {
			log.Printf("Removing stale client: %s, inactive for %v", key, timeSinceLastPing)
			delete(clientManager.Clients, key)
		}
	}
}
func HeartBeatHandle(c *gin.Context) {
	token := c.Query("token")
	decodeToken, err := utils.DecodeBase64(token)
	if err != nil {
		log.Printf("decode token error: %v, token: %v", err, token)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"data": "decode token error"})
	}
	client := &models.Client{}
	err = redis.GetStructValue(decodeToken, client)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"data": "token error"})
	}
	clientManager.HeartBeat <- client
	c.JSON(http.StatusOK, gin.H{})
}
func GetToken(c *gin.Context) {
	deviceId := c.Query("deviceId")
	userId := c.Query("userId")
	isLogin := c.Query("isLogin") == "true"
	if (deviceId == "" && isLogin == false) || (userId == "" && isLogin == true) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"data": "Invalid request parameters"})
		return
	}
	token := utils.EncodeBase64(deviceId + "-" + userId)
	client := &models.Client{
		UserId:       userId,
		DeviceId:     deviceId,
		IsLogin:      isLogin,
		LastPingTime: time.Now(),
		Token:        token,
	}
	err := redis.GetOrSetStructWithExpiration(clientKey+token, client, 30*24*time.Hour)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"data": "json marshal error"})
	}
	c.JSON(http.StatusOK, gin.H{"data": token})
}
