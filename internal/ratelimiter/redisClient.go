package ratelimiter

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

// InitRedis initializes the Redis client and tests the connection
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("UPSTASH_REDIS_REST_URL"), // Replace with your Redis host and port
		Password: os.Getenv("UPSTASH_REDIS_REST_TOKEN"), // Set if your Redis instance has a password
		DB:       0,                // Default DB
	})

	// Ping to test the connection
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("✅ Redis connection established")
}

// GetClient provides a direct handle to the initialized client
func GetClient() *redis.Client {
	if RedisClient == nil {
		InitRedis()
	}
	return RedisClient
}
