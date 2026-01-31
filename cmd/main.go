package main

import (
	"fmt"
	"net/http"

	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // productionda domain tekshiriladi
	},
}

func echo(r *http.Request, w http.ResponseWriter, hub *ws.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Websocketga ulanishda xatolik yuzb erdi : %v\n", err)
		return
	}
	client := &ws.Client{Conn: conn, Send: make(chan *ws.Message), Room: "default"}

	go client.WritePump()
	go client.ReadPump(hub)

	hub.Register <- client
	defer func() {
		hub.Unregister <- client
		conn.Close()
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Xabar o'qishda xatolik yuz berdi: %v\n", err)
			break
		}
		fmt.Printf("Qabul qilingan xabar: %s\n", message)
		hub.Broadcast <- &ws.Message{Data: []byte("salom"), Room: "default"}
	}
}

func main() {
	s := gin.Default()
	hub := ws.NewHub()
	go hub.Run()

	s.GET("/ws", func(c *gin.Context) {
		echo(c.Request, c.Writer, hub)
	})
	s.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	s.Run()
}
