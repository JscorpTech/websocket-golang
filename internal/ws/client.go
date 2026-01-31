package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan *Message
	Room string
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, msg.Data)
	}
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(1024 * 1024)
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		hub.Broadcast <- &Message{
			Data: msg,
			Room: c.Room,
		}
	}
}
