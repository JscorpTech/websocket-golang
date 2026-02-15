package handlers

import (
	"context"
	"encoding/json"
)

type Handler interface {
	Watch(context.Context) chan Message
}

type Message struct {
	Room string          `json:"room"`
	Data json.RawMessage `json:"data"`
}
