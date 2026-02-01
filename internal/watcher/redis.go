package watcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisHandler struct {
	Logger *zap.Logger
	ctx    context.Context
	rdb    *redis.Client
	hub    *ws.Hub
}

type Message struct {
	Room string          `json:"room"`
	Data json.RawMessage `json:"data"`
}

func NewRedisHandler(ctx context.Context, hub *ws.Hub, rdb *redis.Client, logger *zap.Logger) *RedisHandler {
	return &RedisHandler{
		Logger: logger,
		ctx:    ctx,
		rdb:    rdb,
		hub:    hub,
	}
}

func (r *RedisHandler) Watch() {
	r.Logger.Info("Watching Redis events...")
	for {
		pubsub := r.rdb.Subscribe(r.ctx, "websocket")
		r.Logger.Info("Subscribed to Redis channel")

		for msg := range pubsub.Channel() {
			var payload Message
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				r.Logger.Error("Failed to unmarshal Redis message", zap.Error(err))
				continue
			}
			r.hub.Broadcast <- &ws.Message{
				Room: payload.Room,
				Data: payload.Data,
			}
			r.Logger.Info("Received message from Redis", zap.Any("payload", payload))
		}
		r.Logger.Warn("Redis pubsub channel closed. Reconnecting in 1s...")
		pubsub.Close()
		time.Sleep(1 * time.Second)
	}
}
