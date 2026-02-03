package domain

import (
	"errors"
)

var BoardSize = 5; //remember to take this from env file

type CellState int

const (
	Empty CellState = iota
	Ship
	Hit
	Miss
)

type GameStatus string

const (
	StatusWait GameStatus = "WAITING_FOR_SHIP"
	StatusActive GameStatus = "ACTIVE"
	StatusHold GameStatus = "HOLD"
	StatusOver GameStatus = "OVER"
)

type Game struct {
	ID				string						`json:"id"`
	Boards			map[string][][]CellState    `json:"boards"`
	Players			[2]string					`json:"players"`
	ActivePlayer	string						`json:"active_player"`
	Winner 			string						`json:"winner"`
	Status			GameStatus					`json:"status"`  //Current Status of A Game
}

type Point struct {
	X  	int `json:"x"`
	Y	int `json:"y"`
}

var (
	ErrNotYourTurn = errors.New("Invalid turn, not your turn mat")
	ErrOutOfBound = errors.New("Points are out of bound")
	ErrBoardNotFound = errors.New("board not found")
	ErrGameNotStarted = errors.New("Game has not started yet")
	ErrShipPlaced = errors.New("Not Allowed to place Ship")
	ErrShipLimitExceed = errors.New("Ship max/min limit is 5")
	ErrInvalidShipPlacement = errors.New("Invalid Ship Placement")
	ErrGameOver = errors.New("Game is Finished")
	ErrInvalidMove = errors.New("Invalid Move")
)

func NewGame(P1 , P2, roomId string) (*Game) {

	g := &Game{
		ID: roomId,
		Players: [2]string{P1,P2},
		ActivePlayer: P1,
		Winner: "",
		Status: StatusWait,
	}
	BoardsTemp := make(map[string][][]CellState)
	for _ ,i := range g.Players {
		
		PlayerBoard :=  make([][]CellState,BoardSize)
		for j:=range PlayerBoard {
			PlayerBoard[j]  = make([]CellState, BoardSize)
			for k:= range PlayerBoard[j]{
				PlayerBoard[j][k] = Empty
			}
		}
		BoardsTemp[i] = PlayerBoard

	}
	g.Boards = BoardsTemp
	return g
}


func (g *Game) HandleShot(playerID string, p Point) (CellState,error) {

	if g.Status == StatusOver {
		return Empty,ErrGameOver
	}

	if g.Status != StatusActive {
		return Empty, ErrGameNotStarted
	}

	if g.ActivePlayer != playerID {
		return Empty, ErrNotYourTurn
	}

	if p.X < 0 || p.X >= BoardSize || p.Y < 0 || p.Y >= BoardSize {
		return Empty ,ErrOutOfBound
	}

	opponentID := g.GetOpponent(playerID)
	board,ok := g.Boards[opponentID]

	if !ok {
		return Empty, ErrBoardNotFound
	}

	if board[p.X][p.Y] == Hit || board[p.X][p.Y] == Miss {
		return Empty,ErrInvalidMove
	}

	if board[p.X][p.Y] == Ship {
		board[p.X][p.Y] = Hit
		return Hit,nil
	}

	board[p.X][p.Y] = Miss
	g.ActivePlayer = opponentID

	return Miss , nil

}

func (g  *Game) CheckWinner(playerID string) bool {
	opponentID := g.GetOpponent(playerID)
	board:= g.Boards[opponentID]
	countShip := 0
	countHit := 0
	for i:= range board {
		for j := range board[i] {
			if board[i][j] == Hit {
				countHit++
			}
			if board[i][j] == Ship {
				countShip++
			}
		}
	}

	if countShip!=0 || countHit == 0 {
		return  false
	}
	
	g.Status = StatusOver
	return true

}

func (g *Game) GetOpponent(Player string) string {
	if Player != g.Players[0] {
		return g.Players[0]
	}
	return g.Players[1]
}

func (g *Game) AddShip(playerID string, Ships []Point) error {

	if g.Status!=StatusWait {
		return ErrShipPlaced
	}
	board := g.Boards[playerID]

	if len(Ships)!=BoardSize{
		return ErrShipLimitExceed
	}

	for i:= range Ships {
		x := Ships[i].X
		y := Ships[i].Y

		if x < 0 || x>= BoardSize || y < 0 || y>=BoardSize {
			return ErrOutOfBound
		}

		if board[x][y] == Ship {
			return ErrInvalidShipPlacement
		}

		board[x][y] = Ship
	}

	g.Boards[playerID] = board

	return nil
}

func (g *Game) SwitchActivePlayer(playerID string)  {
	opponentID := g.GetOpponent(playerID)
	g.ActivePlayer = opponentID
}

func (g *Game) HideOpponentShips(playerID string) [][]CellState {
	
	opponentId := g.GetOpponent(playerID)

	board := g.Boards[opponentId]  

	for i := range board {
		for j:= range board[i] {
			if board[i][j] == Ship {
				board[i][j] =  Empty
			}
		}
	}
	return board

}


// disconnect user should keep game on hold and add timeout feature

// better option for error handling (if possible)