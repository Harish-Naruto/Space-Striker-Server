package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"
)


func MessageHandler(raw []byte,roomId string,gs *services.GameService) error {
	var msg models.MessageWs;
	
	if err := json.Unmarshal(raw,&msg); err!=nil {
		return err
	}
	
	switch msg.Type {
	case models.TypeMove:
		gs.HandleMove(context.Background(),msg.Sender,roomId,msg.Payload)
	case models.TypePlaceShip:
		gs.HandlePlace(context.Background(),msg.Sender,roomId,msg.Payload)
	default:
		log.Println("Invalid Type of Message")
	}

	return nil
}

