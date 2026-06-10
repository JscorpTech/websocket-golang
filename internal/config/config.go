package config

import (
	"crypto/rsa"
	"os"
	"strconv"
	"strings"

	"github.com/JscorpTech/websocket/internal/auth"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	PublicKEY     *rsa.PublicKey
	ChannelName   string
	// AllowedOrigins: WebSocket upgrade uchun ruxsat etilgan Origin'lar
	// (ALLOWED_ORIGINS, vergul bilan). Bo'sh bo'lsa Origin tekshiruvi
	// o'tkazib yuboriladi (auth baribir query-token orqali amalga oshadi) —
	// lekin to'ldirilsa CSWSH ga qarshi himoya yoqiladi.
	AllowedOrigins []string
	// MaxConnsPerUser: bitta foydalanuvchi (user_<id> xonasi) ochishi mumkin
	// bo'lgan ulanishlar soni. Resurs tugatish (DoS) hujumini cheklaydi.
	MaxConnsPerUser int
}

func NewConfig() *Config {
	pubKey, err := auth.ParseRSAPublicKeyFromPEM([]byte(os.Getenv("PUBLIC_KEY")))
	if err != nil {
		panic("Failed to parse RSA public key: " + err.Error())
	}
	return &Config{
		RedisAddr:       os.Getenv("REDIS_ADDR"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		RedisDB:         0,
		PublicKEY:       pubKey,
		ChannelName:     "websocket",
		AllowedOrigins:  parseOrigins(os.Getenv("ALLOWED_ORIGINS")),
		MaxConnsPerUser: parseIntDefault(os.Getenv("MAX_CONNS_PER_USER"), 10),
	}
}

func parseOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseIntDefault(raw string, def int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && v > 0 {
		return v
	}
	return def
}
