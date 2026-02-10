package ws

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
	"github.com/Harish-Naruto/Space-Striker-Server/internal/services"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 1 * time.Minute
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	roomId string
	gs *services.GameService
	playerID string
}

// readPump read message from the client and broadcast them into hub
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if err := MessageHandler(message,c.roomId,c.gs,c.playerID); err!=nil{
			log.Print(err)
		}

	}
}

// writePump takes message from hub (client send channel) and send that message to client
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}

func ServerWs(h *Hub, gs *services.GameService, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}


	roomID := r.URL.Query().Get("roomID")
	playerID := r.URL.Query().Get("playerID")

	if roomID=="" || playerID=="" {
		log.Println("roomID or PlayerID is missing")
		return
	}
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// throw error
		log.Println(err)
		return
	}

	// create new Client
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
		roomId: roomID,
		gs: gs,
		playerID: playerID,
	}
	// this spawns the read and write thread for each client to send and receive messages
	go client.readPump()
	go client.writePump()
	
	syncPayload := &struct {
    	ServerTime int64 `json:"serverTime"`
	}{
    	ServerTime: time.Now().UnixMilli(),
	}
	p,err := json.Marshal(syncPayload)

	syncMsg := models.MessageWs{
		Type: "SYNC_TIME",
		Payload: json.RawMessage(p),
	}

	temp,errr := json.Marshal(syncMsg)

	if errr!= nil {
		log.Println(errr)
	}

	client.send <- temp

	// adding new client to hub
	h.Register <- client
}

