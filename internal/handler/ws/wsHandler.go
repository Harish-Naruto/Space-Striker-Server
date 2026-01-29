package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"
)


func MessageHandler(raw []byte,roomId string,gs *services.GameService, clientID string) error {
	var msg models.MessageWs;
	
	if err := json.Unmarshal(raw,&msg); err!=nil {
		return err
	}
	
	switch msg.Type {
	case models.TypeMove:
		gs.HandleMove(context.Background(),clientID,roomId,msg.Payload)
	case models.TypePlaceShip:
		gs.HandlePlace(context.Background(),clientID,roomId,msg.Payload)
	case models.TypeChat:
		gs.HandleChat(context.Background(),clientID,roomId,msg.Payload)
	default:
		log.Println("Invalid Type of Message")
	}

	return nil
}

