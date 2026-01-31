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
}

func NewConfig() *Config {
	pubKey, err := auth.ParseRSAPublicKeyFromPEM([]byte(os.Getenv("PUBLIC_KEY")))
	if err != nil {
		panic("Failed to parse RSA public key: " + err.Error())
	}
	return &Config{
		RedisAddr:     "127.0.0.1:6379",
		RedisPassword: "",
		RedisDB:       0,
		PublicKEY:     pubKey,
	}
}
