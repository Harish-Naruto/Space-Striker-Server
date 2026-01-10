package ws

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Hub struct {
	rooms	   map[string]map[*Client]bool
	roomsCancels map[string]context.CancelFunc	
	register   chan *Client
	unregister chan *Client
	broadcast  chan  Message
	rdb		   *redis.Client
	mu   			sync.Mutex
	
}

type Message struct {
	roomID string
	payload []byte
}

func NewHub(rbd *redis.Client) *Hub {
	return &Hub{
		rooms:    make(map[string]map[*Client]bool),
		roomsCancels: make(map[string]context.CancelFunc),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
		rdb: rbd,
	}
}

func (h *Hub) Run()  {
	for {
		select {
		case client:= <-h.register:
			if h.rooms[client.roomId] == nil {
				h.rooms[client.roomId] = make(map[*Client]bool)
				//subscribe to that room 
				ctx,cancel := context.WithCancel(context.Background())
				h.roomsCancels[client.roomId] = cancel
				go h.SubscribeToRoom(ctx,client.roomId)
			}
			h.rooms[client.roomId][client] = true
		case client := <-h.unregister:
			if clients ,ok := h.rooms[client.roomId]; ok {
				if _,ok := clients[client]; ok{
					delete(clients,client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.rooms,client.roomId)
						if cancel,ok := h.roomsCancels[client.roomId]; ok {
							cancel()
							delete(h.roomsCancels,client.roomId)
						}
					}
				}
			}
		case message := <-h.broadcast:
			err := h.rdb.Publish(context.Background(),message.roomID,message.payload).Err()
			if err!= nil {
				log.Printf("Redis publish error : %v", err)
			}
		}
	}
}

func (h *Hub) SubscribeToRoom(ctx context.Context,roomId string){
	
	pubsub := h.rdb.Subscribe(ctx,roomId)
	defer pubsub.Close()
	
	ch := pubsub.Channel()
	for msg := range ch {
		h.mu.Lock()
		for client := range h.rooms[roomId] {
			select{
			case client.send <- []byte(msg.Payload):
			default:
				close(client.send)
				delete(h.rooms[roomId],client)
			}
		}
		h.mu.Unlock()
	}
}
