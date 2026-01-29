package services

import (
	"context"

	"github.com/Harish-Naruto/Space-Striker-Server/pkg/domain"
	"github.com/redis/go-redis/v9"
)

type HttpService struct {
	rdb *redis.Client
}

func CreateHttpService(r *redis.Client) *HttpService {
	return &HttpService{
		rdb: r,
	}
}

func (hs HttpService) RoomValidator(roomId string) bool {
	return hs.rdb.SIsMember(context.Background(),"rooms",roomId).Val()
}

func (hs HttpService) RoomGenerator() (string,error) {
	roomID, err := domain.GenerateRoomID(5)
	if err != nil {
		return "",err
	}

	hs.rdb.SAdd(context.Background(),"rooms",roomID)
	return roomID,nil
}