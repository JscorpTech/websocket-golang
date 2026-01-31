package ws

import (
	"fmt"
)

type Hub struct {
	Rooms map[string]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message),
	}
}

func (h *Hub) Run() {
	fmt.Printf("Hub ishga tushdi\n")
	for {
		select {
		case client := <-h.Register:
			// mijozni ro'yxatga olish
			fmt.Printf("Client: %v\n", client.Conn.RemoteAddr())
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true
		case client := <-h.Unregister:
			// mijozni ro'yxatdan o'tkazish
			fmt.Printf("Client uzuldi: %s\n", client.Conn.RemoteAddr())
			if _, ok := h.Rooms[client.Room][client]; ok {
				delete(h.Rooms[client.Room], client)
				close(client.Send)
				if len(h.Rooms[client.Room]) == 0 {
					delete(h.Rooms, client.Room)
				}
			}
		case msg := <-h.Broadcast:
			// xabarni barcha mijozlarga yuborish
			fmt.Printf("Xabar qabul qilindi: %s\n", msg.Data)
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
