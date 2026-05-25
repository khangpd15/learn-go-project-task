package config

import "task_api/internal/utils"

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     utils.GetEnv("REDIS_ADDR", "localhost:6380"),
		Username: utils.GetEnv("REDIS_USER", ""),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       utils.GetEnvAsInt("REDIS_DB", 0),
	}
}