package database

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func InitRedis(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis at %s: %v", addr, err)
		log.Printf("Some features may be limited without Redis")
	}

	return client
}