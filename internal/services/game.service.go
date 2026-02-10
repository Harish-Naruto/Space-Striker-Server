package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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
		gs.sendError("Invalid move data",playerId)
		return
	}

	if err := gs.repo.LockGame(ctx,roomID);err!=nil{
		gs.sendError(err.Error(),playerId)
		return
	}
	defer gs.repo.DeleteLock(ctx,roomID)

	// delete timer for current player
	gs.repo.RedisClient.Del(ctx,"turn:"+roomID)

	game, err := gs.repo.GetGame(ctx, roomID)
	if err != nil {
		gs.sendError("Game Not Found",playerId)
		return
	}
	// handle shot 
	result, err := game.HandleShot(playerId, domain.Point(move))
	if err != nil {
		gs.sendError(err.Error(),playerId)
		return
	}
	// check for winner
	IsWinner := game.CheckWinner(playerId)
	
	// Switch ActivePlayer
	game.SwitchActivePlayer(playerId)
	game.AddEndAt(40*time.Second)


	if err := gs.repo.SaveGame(ctx, game); err != nil {
		gs.sendError( "Failed to save the game",playerId)
		return
	}
	
	gs.BroadcastMoveResult(result, roomID, move, game, playerId, IsWinner)

	//start timer for next player
	gs.StartTimer(roomID)
}

func (gs *GameService) HandlePlace(ctx context.Context, playerId string, RoomID string, payload json.RawMessage) {
	
	var ships models.PlacePayload
	
	if err := json.Unmarshal(payload, &ships); err != nil {
		gs.sendError("Invalid Ship payload",playerId)
		return
	}
	
	if err := gs.repo.LockGame(ctx, RoomID); err != nil {
		gs.sendError(err.Error(),playerId)
		return
	}
	
	defer gs.repo.DeleteLock(ctx, RoomID)

	game, err := gs.repo.GetGame(ctx, RoomID)
	
	if err != nil {
		gs.sendError(err.Error(),playerId)
		return
	}
	
	size,errShip := gs.repo.AddPlayerShip(ctx,RoomID,playerId)
	
	if errShip != nil {
		gs.sendError(errShip.Error(),playerId)
		return
	}
	// add Ship for a player
	
	game.AddShip(playerId, ships.Ships)
	
	if size == 2 {
		key := "place:"+game.ID
		gs.repo.RedisClient.Del(ctx,key)
		game.Status = domain.StatusActive
		game.AddEndAt(40*time.Second)
	}

	if err := gs.repo.SaveGame(ctx, game); err != nil {
		gs.sendError("Failed to save Game",playerId)
		return
	}

	// Place Payload
	if size == 2 {
		gs.SendGameHistoryToRoom( ctx,RoomID);
		gs.StartTimer(game.ID)
	}
}

func (gs *GameService) HandleJoin(ctx context.Context, playerId string,roomID string) error {
	
	if gs.repo.FindPlayer(ctx,roomID,playerId) {
		// send game updated state / previous state
		gs.SendGameHistory(ctx,playerId,roomID)
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
		game.AddEndAt(60*time.Second)
		if err := gs.repo.SaveGame(ctx,game);err != nil {
			return errors.New("Failed to save Game")
		}
		key := "place:"+game.ID
		gs.repo.SetTimeOut(ctx,key,60*time.Second)
		
		gs.SendGameHistoryToRoom(ctx,roomID);
	}
	
	return nil
}

func (gs *GameService) HandleTurnTimeOut(gameID string)  {
	
	//get game
	game, err := gs.repo.GetGame(context.Background(),gameID);
	if err!=nil {
		log.Println(err)
		return
	}

	//swtich activePlayer and add new timer
	game.SwitchActivePlayer(game.ActivePlayer)
	game.AddEndAt(40*time.Second)

	errGame := gs.repo.SaveGame(context.Background(),game);

	if errGame != nil{
		log.Println(errGame)
		return
	}

	timeOutPayload := &models.TimeOutPayload{
		NextTurn: game.ActivePlayer,
		EndAt: game.EndAt,
	}

	gs.SendToRoom(gameID,models.TypeTimeOut,timeOutPayload)
}

func (gs *GameService) HandleDisconnect(gameID string,playerID string)  {
	//get game
	game,err := gs.repo.GetGame(context.Background(),gameID)
	
	if err != nil {
		log.Println(err)
		return
	}

	//winner
	game.Winner = game.GetOpponent(playerID)

	if err := gs.repo.SaveGame(context.Background(),game); err!= nil{
		log.Println(err)
	}
	
	//send game over
	gs.SendToRoom(gameID,models.TypeGameOver,&models.GameOverPayload{Winner: game.Winner})
}

func (gs *GameService) HandlePlaceTimeOut(gameID string)  {
	gs.SendToRoom(gameID,models.TypeGameOver,&models.GameOverPayload{Winner: ""})
}

func (gs *GameService) HandleGameLimit(gameID string)  {
	gs.SendToRoom(gameID,models.TypeGameOver,&models.GameOverPayload{Winner: ""})
}

// Helpers

func (gs *GameService) SendGameHistory(ctx context.Context,playerId string,roomID string)  {

	// get game
	game,err := gs.repo.GetGame(ctx,roomID)
	if err!= nil{
		log.Printf("Failed To get game %s, err : %v",roomID,err)
		return
	}

	// Process the game to hide oppoent ships
	opponentBoard := game.HideOpponentShips(playerId)
	yourBoard := game.Boards[playerId]

	gameState := models.GameStateResponse{
		Id: roomID,
		YourBoard: yourBoard,
		OpponentBoard: opponentBoard,
		ActivePlayer: game.ActivePlayer,
		Winner: game.Winner,
		Status: game.Status,
		EndAt: game.EndAt,
	}

	// Send it to Solo send
	gs.SendToSolo(ctx,playerId,models.TypeGameState,gameState)
}

func (gs *GameService) BroadcastMoveResult(result domain.CellState, roomId string, move models.MovePayload, game *domain.Game, playerId string, IsWinner bool) {
	resultPayload := models.HitPayload{
		X:        move.X,
		Y:        move.Y,
		Result:   result,
		NextTurn: game.ActivePlayer,
		By:       playerId,
		EndAt: game.EndAt,
	}

	gs.SendToRoom(roomId, models.TypeMove, resultPayload)

	if IsWinner {
		gs.SendToRoom(roomId, models.TypeGameOver, models.GameOverPayload{Winner: playerId})
	}
}

func (gs *GameService) sendError(err string, playerId string) {
	gs.SendToSolo(context.Background(),playerId, models.TypeError, models.ErrorPayload{
		Message: err,
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

func (gs *GameService) SendGameHistoryToRoom(ctx context.Context,roomId string)  {
	players,err := gs.repo.GetPlayers(ctx,roomId);
	if err != nil {
		log.Println("error recieved: ",err);
	}
	for _, p := range players {
		gs.SendGameHistory(ctx,p,roomId);
	}
}

func (gs *GameService) StartTimer(gameID string)  {
	key := "turn:"+gameID
	limit := 40 * time.Second

	gs.repo.SetTimeOut(context.Background(),key,limit)
}