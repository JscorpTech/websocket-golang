package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JscorpTech/websocket/internal/auth"
	"github.com/JscorpTech/websocket/internal/config"
	"github.com/JscorpTech/websocket/internal/watcher"
	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(r *http.Request, w http.ResponseWriter, hub *ws.Hub, logger *zap.Logger, conf *config.Config) {
	query := r.URL.Query()
	token := query.Get("token")
	if token == "" {
		logger.Info("Token topilmadi")
		http.Error(w, "Token topilmadi", http.StatusUnauthorized)
		return
	}
	claimData, err := auth.VerifyAndParseJWT(token, conf.PublicKEY)
	if err != nil {
		logger.Error("token verification error", zap.Error(err))
		http.Error(w, "Token xato", http.StatusUnauthorized)
		return
	}
	userIDRaw, ok := claimData["user_id"]
	if !ok {
		logger.Info("claim data ichida user_id mavjud emas")
		http.Error(w, "user id mavjud emas", http.StatusInternalServerError)
		return
	}
	userIDFloat, _ := userIDRaw.(float64)
	userIDStr := strconv.FormatFloat(userIDFloat, 'f', -1, 64)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Websocketga ulanishda xatolik yuzb erdi", zap.Error(err))
		return
	}
	client := &ws.Client{Conn: conn, Send: make(chan *ws.Message), Room: "user_" + userIDStr}

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
			logger.Error("Xabar o'qishda xatolik yuz berdi", zap.Error(err))
			break
		}
		hub.Broadcast <- &ws.Message{Data: message, Room: "default"}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	router := gin.Default()
	conf := config.NewConfig()
	logger, err := zap.NewProduction()

	if err != nil {
		fmt.Println("Logger ishga tushmadi")
		return
	}
	hub := ws.NewHub(logger)
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Regisga ulanishda xatolik yuz berdi", zap.Error(err))
		return
	}

	go hub.Run()
	redisWatcher := watcher.NewRedisHandler(ctx, hub, rdb, logger)
	go redisWatcher.Watch()

	router.GET("/ws", func(c *gin.Context) {
		serveWs(c.Request, c.Writer, hub, logger, conf)
	})
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		logger.Info("server ishga tushdi :8080")
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("xatolik yuz berdi", zap.Error(err))
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Serverni to'xtatishda xatolik yuz berdi")
	}
	logger.Info("Server to'xtatildi")
}
