package handlers

import (
	"context"

	"github.com/JscorpTech/websocket/internal/config"
	"github.com/JscorpTech/websocket/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RabbitMQHandler struct {
	Logger *zap.Logger
	ctx    context.Context
	rdb    *redis.Client
	hub    *ws.Hub
	conf   *config.Config
}

func NewRabbitMQHandler() Handler {
	return &RabbitMQHandler{}
}

func (h *RabbitMQHandler) Watch(ctx context.Context) chan Message {
	return nil
}
