package ws

import (
	"github.com/JscorpTech/websocket/internal/metrics"
	"go.uber.org/zap"
)

type Hub struct {
	Rooms map[string]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	Logger     *zap.Logger
	// MaxConnsPerUser: bitta xona (user_<id>) uchun ulanish chegarasi.
	// 0 yoki manfiy bo'lsa cheklanmaydi.
	MaxConnsPerUser int
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
			// Per-user ulanish chegarasi — bitta foydalanuvchi cheksiz ulanish
			// ochib resursni tugata olmasin (DoS). Chegaradan oshsa yangi
			// ulanishni rad etamiz: Send'ni yopamiz (WritePump tugaydi) va
			// Conn'ni yopamiz (ReadPump tugaydi). Xonaga qo'shilmaydi.
			if h.MaxConnsPerUser > 0 && len(h.Rooms[client.Room]) >= h.MaxConnsPerUser {
				h.Logger.Warn("Per-user ulanish chegarasi oshdi, rad etildi",
					zap.String("room", client.Room),
					zap.Int("limit", h.MaxConnsPerUser))
				close(client.Send)
				client.Conn.Close()
				continue
			}
			h.Rooms[client.Room][client] = true
			metrics.ActiveConnections.Inc()
		case client := <-h.Unregister:
			h.Logger.Info("Client uzuldi", zap.String("address", client.Conn.RemoteAddr().String()))
			if _, ok := h.Rooms[client.Room][client]; ok {
				delete(h.Rooms[client.Room], client)
				close(client.Send)
				if len(h.Rooms[client.Room]) == 0 {
					delete(h.Rooms, client.Room)
				}
				metrics.ActiveConnections.Dec()
			}
		case msg := <-h.Broadcast:
			for room := range h.Rooms[msg.Room] {
				// Collab op'ni yuboruvchining o'ziga qaytarmaymiz (echo yo'q).
				if msg.Sender != nil && room == msg.Sender {
					continue
				}
				select {
				case room.Send <- msg:
					metrics.BroadcastMessages.Inc()
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
