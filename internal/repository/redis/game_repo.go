package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Harish-Naruto/Space-Striker-Server/pkg/domain"
	"github.com/redis/go-redis/v9"
)

var (
	ErrGameFull = errors.New("Game Full") 
)

type RedisGameRepository struct {
	RedisClient *redis.Client
}

func (R *RedisGameRepository) SaveGame(ctx context.Context,g *domain.Game)	error {
	data, err := json.Marshal(g)
	if err!=nil {
		return err
	}
	return R.RedisClient.Set(ctx,"game:"+g.ID,data,0).Err()
}

func (R *RedisGameRepository) GetGame(ctx context.Context,id string) (*domain.Game,error) {
	data,err := R.RedisClient.Get(ctx,"game:"+id).Bytes()
	if err!=nil {
		return nil,err
	}

	var g domain.Game

	if err := json.Unmarshal(data,&g); err!=nil {
		return nil,err
	}

	return &g, nil
}

func (R *RedisGameRepository) LockGame(ctx context.Context, gameID string) error {
	lock := "lock:game-"+gameID
	ok, err := R.RedisClient.SetNX(ctx,lock,"locked",5*time.Second).Result()
	if err!=nil || !ok {
		return errors.New("faild to lock game")
	}
	return nil
}

func (R *RedisGameRepository) DeleteLock(ctx context.Context, gameID string) error {
	lock := "lock:game"+gameID
	return R.RedisClient.Del(ctx,lock).Err()
}

func (R *RedisGameRepository) AddPlayerToGame(ctx context.Context, gameId string, playerID string) ( int64 ,error) {
	activeGameKey := "Active:game-"+gameId
	numberPlayers := R.RedisClient.SCard(ctx,activeGameKey).Val()

	if numberPlayers >=2 {
		return 0,ErrGameFull
	}
	R.RedisClient.SAdd(ctx,activeGameKey,playerID)
	numberPlayers = R.RedisClient.SCard(ctx,activeGameKey).Val()
	return numberPlayers,nil
}

func (R *RedisGameRepository) GetPlayers(ctx context.Context,gameID string) ([]string,error) {
	activeGameKey := "Active:game-"+gameID
	data := R.RedisClient.SMembers(ctx,activeGameKey)
	var players []string = data.Val();
	return players,nil
}

func (R *RedisGameRepository) FindPlayer(ctx context.Context,gameID string, playerID string) bool {
	activeKey:="Active:game-"+gameID
	check :=R.RedisClient.SIsMember(ctx,activeKey,playerID).Val()
	return check
}

func (R *RedisGameRepository) GetPlayerServer(ctx context.Context,playerID string) string {
	
	serverID,err := R.RedisClient.HGet(ctx,"presence",playerID).Result()

	if err == redis.Nil {
		log.Printf("playerID , %s , not found in redis",playerID)
		return ""
	}

	if err != nil {
		log.Printf("Error while finding ServerID for PlayerID %s : %v",playerID,err)
		return ""
	}

	return serverID
}