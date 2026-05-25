package config

import "task_api/internal/utils"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:     utils.GetEnv("DB_HOST", "localhost"),
		Port:     utils.GetEnv("DB_PORT", "5433"),
		User:     utils.GetEnv("DB_USER", "khangdinh1510"),
		Password: utils.GetEnv("DB_PASSWORD", "123"),
	}
}