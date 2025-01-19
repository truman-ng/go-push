package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-push/internal/config"
	"log"
	"time"
)

var (
	client *redis.ClusterClient
	ctx    = context.Background()
)

func NewRedisClient(cfg *config.Config) {
	client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:      cfg.Redis.Addrs,
		Password:   cfg.Redis.Password,
		ClientName: cfg.Redis.ClientName,
	})

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("redis ping err:", err)
	}
	fmt.Println("redis ping result:", pong)
}

func GetOrSetStructWithExpiration(key string, value interface{}, expiration time.Duration) error {
	_, err := client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		jsonData, err := json.Marshal(value)
		if err != nil {
			log.Printf("Failed to serialize struct: %v", err) // 打印日志
			return fmt.Errorf("failed to serialize struct: %v", err)
		}

		// 存储到 Redis
		err = client.Set(ctx, key, jsonData, expiration).Err()
		if err != nil {
			log.Printf("Failed to store struct in Redis: %v", err) // 打印日志
			return fmt.Errorf("failed to store struct in Redis: %v", err)
		}
		return nil
	} else if err != nil {
		log.Printf("Failed to store struct in Redis: %v", err) // 打印日志
		// 查询发生其他错误
		fmt.Printf("Error querying key: %v\n", err)
	}
	return nil
}

func GetStructValue(token string, dest interface{}) error {
	jsonData, err := client.Get(ctx, token).Result()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("key does not exist: %s", token)
	} else if err != nil {
		return fmt.Errorf("failed to get value from Redis: %v", err)
	}

	err = json.Unmarshal([]byte(jsonData), dest)
	if err != nil {
		return fmt.Errorf("failed to deserialize struct: %v", err)
	}
	return nil
}
