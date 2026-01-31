package config

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func NewConfig() *Config {
	return &Config{
		RedisAddr:     "127.0.0.1:6379",
		RedisPassword: "",
		RedisDB:       0,
	}
}
