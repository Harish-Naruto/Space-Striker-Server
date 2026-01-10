package infra

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func CreateRedisClient(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB: 0,
		MinIdleConns: 2,
	})
	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	err:= rdb.Ping(ctx).Err()

	if err!=nil {
		log.Fatalf("Redis connection failed: %v",err)
	}
	log.Printf("Redis connected")

	return rdb
}