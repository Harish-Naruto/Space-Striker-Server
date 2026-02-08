package game

import (
	"context"
	"log"
	"strings"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"
	"github.com/redis/go-redis/v9"
)

func ListenForTimeOut(ctx context.Context,rdb *redis.Client,gs *services.GameService)  {
	pubsub := rdb.Subscribe(ctx,"__keyevent@0__:expired")
	defer pubsub.Close()
	log.Println("listening for Redis move Timeout")

	for {
		msg,err := pubsub.ReceiveMessage(ctx)
		if err!=nil {
			log.Printf("error while listening to Timeout Message : %v",err)
			continue
		}
		key := msg.Payload
		parts := strings.Split(key,":")

		switch parts[0] {
		case "turn":
			gameID := parts[1]
			gs.HandleTurnTimeOut(gameID)
		case "disconnect":
			gameID := parts[1]
			playerID := parts[2]
			gs.HandleDisconnect(gameID,playerID)
		case "place":
			gameID := parts[1]
			gs.HandlePlaceTimeOut(gameID)
		case "game":
			gameID := parts[1]
			gs.HandleGameLimit(gameID)
		}
		
	}
}