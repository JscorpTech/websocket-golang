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
		frameType := websocket.TextMessage
		if msg.Binary {
			frameType = websocket.BinaryMessage
		}
		c.Conn.WriteMessage(frameType, msg.Data)
	}
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(1024 * 1024)
	for {
		msgType, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Client-originated xabar (collab op) — xona ichidagi BOSHQA
		// klientlarga relay qilinadi (Sender o'ziga qaytmaydi). Frame turi
		// (binary/text) saqlanadi.
		hub.Broadcast <- &Message{
			Data:   msg,
			Room:   c.Room,
			Binary: msgType == websocket.BinaryMessage,
			Sender: c,
		}
	}
}
