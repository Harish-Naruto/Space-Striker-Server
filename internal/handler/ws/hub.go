package ws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Harish-Naruto/Space-Striker-Server/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	ServerID	string
	Rooms	   map[string]map[*Client]bool
	Clients		map[string]*Client
	RoomsCancels map[string]context.CancelFunc	
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan  models.Message
	rdb		   *redis.Client
	mu   			sync.Mutex
}


func NewHub(rbd *redis.Client) *Hub {
	
	Id := uuid.NewString()

	return &Hub{
		ServerID: Id,
		Rooms:    make(map[string]map[*Client]bool),
		RoomsCancels: make(map[string]context.CancelFunc),
		Clients: make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan models.Message),
		rdb: rbd,
	}
}

func (h *Hub) Run()  {

	go h.ListenToSolo(context.Background())

	for {
		select {
		case client:= <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.roomId] == nil {
				h.Rooms[client.roomId] = make(map[*Client]bool)
				ctx,cancel := context.WithCancel(context.Background())
				h.RoomsCancels[client.roomId] = cancel
				go h.SubscribeToRoom(ctx,client.roomId)
			}

			h.rdb.HSet(context.Background(),"presence",client.playerID,h.ServerID)
			h.Rooms[client.roomId][client] = true
			h.Clients[client.playerID] = client
			h.mu.Unlock()

			go func() {
				if err := client.gs.HandleJoin(context.Background(),client.playerID,client.roomId); err!= nil {
					log.Println("player got removed due to err : ",err)
					h.Unregister<-client
				}
			}()
			log.Println("user : "+client.playerID+" Joined")

		case client := <-h.Unregister:
			h.mu.Lock()			
			if clients ,ok := h.Rooms[client.roomId]; ok {
				if _,ok := clients[client]; ok{
					delete(clients,client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.Rooms,client.roomId)
						if cancel,ok := h.RoomsCancels[client.roomId]; ok {
							cancel()
							delete(h.RoomsCancels,client.roomId)
						}
					}
					delete(h.Clients,client.playerID)
					h.rdb.HDel(context.Background(),"presence",client.playerID)
					log.Println("user : "+client.playerID+" Removed")

				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			err := h.rdb.Publish(context.Background(),message.RoomID,message.Payload).Err()
			
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
		for client := range h.Rooms[roomId] {
			select{
			case client.send <- []byte(msg.Payload):
			default:
				close(client.send)
				delete(h.Rooms[roomId],client)
			}
		}
		h.mu.Unlock()
	}
}

func (h *Hub) ListenToSolo(ctx context.Context)  {
	pubsub := h.rdb.PSubscribe(ctx,fmt.Sprintf("solo:%s:*",h.ServerID))
	defer pubsub.Close()

	ch := pubsub.Channel()

	for message := range ch {
		parts := strings.Split(message.Channel, ":")
		playerID := parts[2]
		h.mu.Lock()
		if client, ok := h.Clients[playerID]; ok {
			client.send <- []byte(message.Payload)
		}
		h.mu.Unlock()
	}
}

func (h *Hub) BroadcastMessage(roomID string, payload []byte)  {
	
	h.Broadcast <- models.Message{
		RoomID:  roomID,
		Payload: payload,
	}
}

func (h *Hub) SoloMessage(channel string,payload []byte)  {
	if err := h.rdb.Publish(context.Background(),channel,payload).Err(); err!= nil {
		log.Printf("Redis publish error : %v",err)
	}
}
