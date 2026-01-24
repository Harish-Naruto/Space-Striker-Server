package models

import (
	"encoding/json"

	"github.com/Harish-Naruto/Space-Striker-Server/pkg/domain"
)

type Message struct {
	RoomID string
	Payload []byte
}

type MessageType string

const (
	TypeMove MessageType = "MOVE"
	TypeChat MessageType = "CHAT"
	TypeGameState MessageType = "GAME_STATE"
	TypeGameOver MessageType = "GAME_OVER"
	TypeError MessageType = "ERROR"
	// TypeHistory MessageType = "HISTORY" not using this rn
	TypePlaceShip MessageType = "PLACE_SHIP"
	TypeGameUpdate MessageType = "GAME_UPDATE"

)

type MessageWs struct {
	Type MessageType `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GameOverPayload struct {
	Winner string `json:"winner"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type HitPayload struct {
	X        int              `json:"x"`
	Y        int              `json:"y"`
	Result   domain.CellState `json:"result"`
	NextTurn string           `json:"nextTurn"`
	By       string           `json:"by"`
}

type PlacePayload struct {
	Ships []domain.Point `json:"ships"`
}

type UpdatePayload struct {
	Message string `json:"message"`
}

type GameStateResponse struct {
	Id string `json:"id"`
	YourBoard [][]domain.CellState `json:"yourBoard"`
	OpponentBoard [][]domain.CellState `json:"opponentBoard"`
	ActivePlayer string	`json:"activePlayer"`
	Winner	string	`json:"winner"`
	Status	domain.GameStatus	`json:"status"` 
}