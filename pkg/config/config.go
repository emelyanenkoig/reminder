package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Bot      BotConfig
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
	SSLMode  string
}

type ServerConfig struct {
	Host string
	Port int
}

type BotConfig struct {
	Token string `mapstructure:"token"`
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("DATABASE_PORT"))
	if err != nil {
		return nil, err
	}

	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		return nil, err
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     os.Getenv("DATABASE_HOST"),
			User:     os.Getenv("DATABASE_USER"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			DBName:   os.Getenv("DATABASE_DBNAME"),
			Port:     port,
			SSLMode:  os.Getenv("DATABASE_SSLMODE"),
		},
		Server: ServerConfig{
			Host: os.Getenv("SERVER_HOST"),
			Port: serverPort,
		},
		Bot: BotConfig{
			Token: os.Getenv("BOT_TOKEN"),
		},
	}

	return config, nil
}
func LoadLocalConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			User:     "postgres",
			Password: "postgres",
			DBName:   "reminder",
			Port:     5432,
			SSLMode:  "disable",
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Bot: BotConfig{
			Token: "token",
		},
	}
}
