package redis

import "github.com/redis/go-redis/v9"

type ChatRedisRepository struct {
	redisClient *redis.Client
}

func (c *ChatRedisRepository) SaveChat()  {
	
}