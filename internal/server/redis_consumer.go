package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go-push/internal/models"
	"log"
	"time"
)

var (
	streamKey   = "score-data:push_stream"
	groupName   = "score-data:push_group"
	consumer    = "go-" + generateUUID() // 或使用 pod IP/hostname
	redisClient *redis.ClusterClient
)

func generateUUID() string {
	return uuid.New().String()
}

func InitRedisConsumer(rdb *redis.ClusterClient) {
	redisClient = rdb

	ctx := context.Background()

	// 创建消费组，忽略已存在的错误
	_ = redisClient.XGroupCreateMkStream(ctx, streamKey, groupName, "$").Err()

	go startConsuming()
}

func startConsuming() {
	ctx := context.Background()

	for {
		streams, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumer,
			Streams:  []string{streamKey, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("❌ Redis Stream Read Error: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {

				payloadRaw, ok := msg.Values["payload"]
				if !ok {
					log.Println("❌ Redis Stream 中未包含 payload 字段")
					continue
				}
				
				payloadStr, ok := payloadRaw.(string)
				if !ok {
					log.Println("❌ payload 字段不是字符串类型")
					continue
				}
				
				// payload 是个嵌套 JSON 字符串，需要先反序列化
				var finalPayload map[string]interface{}
				err := json.Unmarshal([]byte(payloadStr), &finalPayload)
				if err != nil {
					log.Printf("❌ payload 字符串解析失败: %v", err)
					continue
				}
				 msgRoom, ok := finalPayload["roomId"].(string) 
				if !ok || msgRoom == "" { 	
					continue 
				}

				jsonBytes, err := json.Marshal(finalPayload)
				if err != nil {
					log.Printf("❌ 最终 JSON 序列化失败: %v", err)
					continue
				}
				for _, client := range clientManager.Clients {
					if !clientInRoom(client, msgRoom) {
						continue
					}
				
					select {
						case clientManager.Broadcast <- jsonBytes:
						default:
							log.Printf("⚠️ Broadcast 通道已满，丢弃消息: %s", jsonBytes)
						}
				}

				
				
				redisClient.XAck(ctx, streamKey, groupName, msg.ID)
			}
		}
	}
}
func clientInRoom(client *models.Client, room string) bool {
	for _, r := range client.RoomIds {
		if r == room {
			return true
		}
	}
	return false
}


