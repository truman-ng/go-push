package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

				// ✅ 将 map[string]interface{} 转成 JSON
				jsonBytes, err := json.Marshal(msg.Values)
				if err != nil {
					log.Printf("❌ 消息 JSON 序列化失败: %v", err)
					continue
				}
				select {
				case clientManager.Broadcast <- jsonBytes:
					// OK
				default:
					log.Printf("⚠️ Broadcast 通道已满，丢弃消息: %s", jsonBytes)
				}

				// ✅ 手动 ACK
				redisClient.XAck(ctx, streamKey, groupName, msg.ID)
			}
		}
	}
}
