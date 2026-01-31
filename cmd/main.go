package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	router := gin.Default()
	hub := ws.NewHub()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("Regisga ulanishda xatolik yuz berdi")
		return
	}

	go hub.Run()

	router.GET("/ws", func(c *gin.Context) {
		echo(c.Request, c.Writer, hub)
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		fmt.Println("server ishga tushdi :8080")
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("xatolik yuz berdi")
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Serverni to'xtatishda xatolik yuz berdi")
	}
	fmt.Println("Server to'xtatildi")
}
