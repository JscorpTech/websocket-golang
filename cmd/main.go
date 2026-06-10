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
	"github.com/JscorpTech/websocket/internal/handlers"
	"github.com/JscorpTech/websocket/internal/metrics"
	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// allowedOrigins, agar to'ldirilgan bo'lsa, WebSocket upgrade'da Origin
// header'ini cheklaydi (CSWSH himoyasi). Bo'sh bo'lsa — barcha Origin'larga
// ruxsat (auth baribir query-token orqali amalga oshadi).
var allowedOrigins []string

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if len(allowedOrigins) == 0 {
			return true
		}
		origin := r.Header.Get("Origin")
		// Origin yo'q (native/server klient) — ruxsat. Brauzer CSWSH'da
		// doim Origin yuboradi, shuning uchun bu cheklov brauzerlarga ta'sir qiladi.
		if origin == "" {
			return true
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
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
	// Send kanali BUFERLI bo'lishi shart: node_execution burst'larida (bitta
	// interaksiya = millisekundlarda 6-10 event) bufersiz kanal band bo'lib,
	// hub'dagi non-blocking send `default`ga tushar va klientni DARHOL uzib
	// yuborardi ("juda tez uzulib qayta ulanish" + yo'qolgan eventlar).
	client := &ws.Client{Conn: conn, Send: make(chan *ws.Message, 256), Room: "user_" + userIDStr}

	go client.WritePump()
	go client.ReadPump(hub)

	hub.Register <- client
	// defer func() {
	// 	hub.Unregister <- client
	// 	conn.Close()
	// }()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg := prometheus.NewRegistry()
	reg.Register(metrics.ActiveConnections)
	reg.Register(metrics.BroadcastMessages)

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode) // gin release mode

	conf := config.NewConfig()
	logger, err := zap.NewProduction()

	if err != nil {
		fmt.Println("Logger ishga tushmadi")
		return
	}
	allowedOrigins = conf.AllowedOrigins
	hub := ws.NewHub(logger)
	hub.MaxConnsPerUser = conf.MaxConnsPerUser
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
	var handler handlers.Handler
	handler = handlers.NewRedisHandler(ctx, conf, hub, rdb, logger)
	messages := handler.Watch(ctx)
	go func() {
		for msg := range messages {
			hub.Broadcast <- &ws.Message{
				Room: msg.Room,
				Data: msg.Data,
			}
			logger.Info("Received message", zap.Any("payload", msg))
		}
	}()

	router.GET("/ws/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	router.GET("/ws/events", func(c *gin.Context) {
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
