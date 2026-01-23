package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/repository/redis"
	"github.com/Harish-Naruto/Space-Striker-Server/pkg/domain"
)

type HubInterface interface {
	BroadcastMessage (roomID string, payload []byte)
	SoloMessage (channel string, payload []byte)
}


type GameService struct {
	repo redis.RedisGameRepository
	hub  HubInterface
}

func NewGameService(r redis.RedisGameRepository, h HubInterface) *GameService {
	return &GameService{
		repo: r,
		hub:  h,
	}
}

// Handlers

func (gs *GameService) HandleMove(ctx context.Context, playerId string, roomID string, payload json.RawMessage) {
	var move models.MovePayload
	if err := json.Unmarshal(payload, &move); err != nil {
		gs.sendError(roomID, "Invalid move data",playerId)
		return
	}

	game, err := gs.repo.GetGame(ctx, roomID)
	if err != nil {
		gs.sendError(roomID, "Game Not Found",playerId)
		return
	}
	// handle shot 
	result, err := game.HandleShot(playerId, domain.Point(move))
	if err != nil {
		gs.sendError(roomID, err.Error(),playerId)
		return
	}
	// check for winner
	IsWinner := game.CheckWinner(playerId)
	
	// Switch ActivePlayer
	game.SwitchActivePlayer(playerId)

	if err := gs.repo.SaveGame(ctx, game); err != nil {
		gs.sendError(roomID, "Failed to save the game",playerId)
		return
	}
	// send payload
	gs.BroadcastMoveResult(result, roomID, move, game, playerId, IsWinner)
}

func (gs *GameService) HandlePlace(ctx context.Context, playerId string, RoomID string, payload json.RawMessage) {
	defer gs.repo.DeleteLock(ctx, RoomID)

	var ships models.PlacePayload

	if err := json.Unmarshal(payload, &ships); err != nil {
		gs.sendError(RoomID, "Invalid Ship payload",playerId)
		return
	}

	if err := gs.repo.LockGame(ctx, RoomID); err != nil {
		gs.sendError(RoomID, err.Error(),playerId)
		return
	}

	game, err := gs.repo.GetGame(ctx, RoomID)
	if err != nil {
		gs.sendError(RoomID, err.Error(),playerId)
		return
	}
	// add Ship for a player
	game.AddShip(playerId, ships.Ships)
	if err := gs.repo.SaveGame(ctx, game); err != nil {
		gs.sendError(RoomID, "Failed to save Game",playerId)
		return
	}

	// Place Payload
	gs.SendToRoom(RoomID, models.TypeGameUpdate, models.UpdatePayload{Message: "Ship Added For Player :" + playerId})
}

func (gs *GameService) HandleJoin(ctx context.Context, playerId string,roomID string) error {

	defer gs.repo.DeleteLock(ctx,roomID)
	if err := gs.repo.LockGame(ctx,roomID); err!=nil{
		gs.SendToSolo(ctx,playerId,models.TypeGameUpdate,models.UpdatePayload{Message: "Player here is your game update"})
		return nil
	}

	if gs.repo.FindPlayer(ctx,roomID,playerId) {
		// send game updated state
		gs.SendToSolo(ctx,playerId,models.TypeGameUpdate,models.UpdatePayload{Message: "Player here is your game update"})
		return nil
	}

	number,err := gs.repo.AddPlayerToGame(ctx,roomID,playerId)
	if err != nil{
		return err
	}
	if number == 2 {
		players,err := gs.repo.GetPlayers(ctx,roomID)
		if len(players) == 0 || err!= nil {
			return errors.New("Game doesnot Exist")
		}
		game := domain.NewGame(players[0],players[1],roomID)
		if err := gs.repo.SaveGame(ctx,game);err != nil {
			return errors.New("Failed to save Game")
		}

		  gs.SendToRoom(roomID, models.TypeGameUpdate,models.UpdatePayload{Message: "Game Created for roomID:"+roomID})
	}
	 gs.SendToRoom(roomID,models.TypeGameUpdate,models.UpdatePayload{Message: "Player:"+ playerId+ " added to roomID:"+roomID})
	return nil
}

// Helpers

func (gs *GameService) BroadcastMoveResult(result domain.CellState, roomId string, move models.MovePayload, game *domain.Game, playerId string, IsWinner bool) {
	resultPayload := models.HitPayload{
		X:        move.X,
		Y:        move.Y,
		Result:   result,
		NextTurn: game.ActivePlayer,
		By:       playerId,
	}

	gs.SendToRoom(roomId, models.TypeMove, resultPayload)

	if IsWinner {
		gs.SendToRoom(roomId, models.TypeGameOver, models.GameOverPayload{Winner: playerId})
	}
}

func (gs *GameService) sendError(roomId string, err string, playerId string) {
	gs.SendToRoom(roomId, models.TypeError, models.ErrorPayload{
		Message: err,
		To: playerId,
	})
}

func (gs *GameService) SendToRoom(roomId string, msgType models.MessageType, payload interface{}) {
	response := models.MessageWs{
		Type:    msgType,
		Payload: toRawMessage(payload),
	}

	msg, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal response :%v", response)
		return
	}
	
	gs.hub.BroadcastMessage(roomId,msg)
	
}

func (gs *GameService) SendToSolo(ctx context.Context,playerID string,msgType models.MessageType,payload interface{})  {
	// get serverId first
	ServerID := gs.repo.GetPlayerServer(ctx,playerID)

	if ServerID == "" {
		return
	}

	channel := fmt.Sprintf("solo:%s:%s",ServerID,playerID)

	// Parse the message
	response := models.MessageWs{
		Type:    msgType,
		Payload: toRawMessage(payload),
	}

	msg, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal response :%v", response)
		return
	}

	// publish message to the channel solo:serverID:playerID using
	gs.hub.SoloMessage(channel,msg)
	
}

func toRawMessage(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("failed to marshal payload: %v", err)
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}
