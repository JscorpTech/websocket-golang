package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JscorpTech/websocket/internal/config"
	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisHandler struct {
	Logger *zap.Logger
	ctx    context.Context
	rdb    *redis.Client
	hub    *ws.Hub
	conf   *config.Config
}

func NewRedisHandler(ctx context.Context, conf *config.Config, hub *ws.Hub, rdb *redis.Client, logger *zap.Logger) *RedisHandler {
	return &RedisHandler{
		Logger: logger,
		ctx:    ctx,
		rdb:    rdb,
		hub:    hub,
		conf:   conf,
	}
}

func (r *RedisHandler) Watch(ctx context.Context) chan Message {
	r.Logger.Info("Watching Redis events...")
	messages := make(chan Message)
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.Logger.Info("Stopping Redis watcher...")
				close(messages)
				return
			default:
				pubsub := r.rdb.Subscribe(r.ctx, r.conf.ChannelName)
				r.Logger.Info("Subscribed to Redis channel")

				for msg := range pubsub.Channel() {
					var payload Message
					if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
						r.Logger.Error("Failed to unmarshal Redis message", zap.Error(err))
						continue
					}
					messages <- payload
				}
				r.Logger.Warn("Redis pubsub channel closed. Reconnecting in 1s...")
				pubsub.Close()
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return messages
}
