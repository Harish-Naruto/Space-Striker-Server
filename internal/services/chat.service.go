package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
)

func (gs *GameService) HandleChat(ctx context.Context,playerId string, roomId string,payload json.RawMessage)  {
	var msgpayload models.ChatPayload

	if err := json.Unmarshal(payload,&msgpayload); err != nil {
		log.Println("error while parsing chat message : ",err)
	}

	msgpayload.Sender = playerId

	gs.SendToRoom(roomId,models.TypeChat,msgpayload)

}