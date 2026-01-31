package ws

import (
	"go.uber.org/zap"
)

type Hub struct {
	Rooms map[string]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	Logger     *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message),
		Logger:     logger,
	}
}

func (h *Hub) Run() {
	h.Logger.Info("Hub ishga tushdi")
	for {
		select {
		case client := <-h.Register:
			h.Logger.Info("Client", zap.String("address", client.Conn.RemoteAddr().String()))
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
		case client := <-h.Unregister:
			h.Logger.Info("Client uzuldi", zap.String("address", client.Conn.RemoteAddr().String()))
			if _, ok := h.Rooms[client.Room][client]; ok {
				delete(h.Rooms[client.Room], client)
				close(client.Send)
				if len(h.Rooms[client.Room]) == 0 {
					delete(h.Rooms, client.Room)
				}
			}
		case msg := <-h.Broadcast:
			for room := range h.Rooms[msg.Room] {
				select {
				case room.Send <- msg:
				default:
					close(room.Send)
					delete(h.Rooms[msg.Room], room)
					if len(h.Rooms[msg.Room]) == 0 {
						delete(h.Rooms, msg.Room)
					}
				}
			}
		}
	}
}
