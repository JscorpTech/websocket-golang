package config

import (
	"crypto/rsa"
	"os"

	"github.com/JscorpTech/websocket/internal/auth"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	PublicKEY     *rsa.PublicKey
	ChannelName   string
}

func NewConfig() *Config {
	pubKey, err := auth.ParseRSAPublicKeyFromPEM([]byte(os.Getenv("PUBLIC_KEY")))
	if err != nil {
		panic("Failed to parse RSA public key: " + err.Error())
	}
	return &Config{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		PublicKEY:     pubKey,
		ChannelName:   "websocket",
	}
}
