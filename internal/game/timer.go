package game

import (
	"context"
	"log"
	"strings"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"
	"github.com/redis/go-redis/v9"
)

func ListenForTimeOut(ctx context.Context, rdb *redis.Client, gs *services.GameService) {
    pubsub := rdb.Subscribe(ctx, "__keyevent@0__:expired")
    defer pubsub.Close()
    log.Println("Listening for Redis move/game timeouts...")

    ch := pubsub.Channel()

    for {
        select {
        case <-ctx.Done():
            log.Println("Stopping Redis timeout listener...")
            return
        case msg := <-ch:
            key := msg.Payload

            parts := strings.Split(key, ":")
            if len(parts) < 2 {
                continue 
            }

            prefix := parts[0]
            gameID := parts[1]

            switch prefix {
            case "turn":
                gs.HandleTurnTimeOut(gameID)
            case "disconnect":
                if len(parts) >= 3 {
                    playerID := parts[2]
                    gs.HandleDisconnect(gameID, playerID)
                }
            case "place":
                gs.HandlePlaceTimeOut(gameID)
            case "game":
                gs.HandleGameLimit(gameID)
            default:
                // Ignore keys we don't care about
                continue
            }
        }
    }
}