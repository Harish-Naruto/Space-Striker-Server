package redis

import (
	"context"
	"encoding/json"
	"github.com/Harish-Naruto/Space-Striker-Server/pkg/domain"
	"github.com/redis/go-redis/v9"
)

type RedisGameRepository struct {
	redisClient *redis.Client
}

func (R *RedisGameRepository) SaveGame(ctx context.Context,g *domain.Game)	error {
	data, err := json.Marshal(g)
	if err!=nil {
		return err
	}
	return R.redisClient.Set(ctx,"game:"+g.ID,data,0).Err()
}

func (R *RedisGameRepository) GetGame(ctx context.Context,id string) (*domain.Game,error) {
	data,err := R.redisClient.Get(ctx,"game:"+id).Bytes()
	if err!=nil {
		return nil,err
	}

	var g domain.Game

	if err := json.Unmarshal(data,&g); err!=nil {
		return nil,err
	}

	return &g, nil
}