package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ConfigEnv holds the configuration values for the application.
type ConfigEnv struct {
	BotToken  string
	UserID    int64
	ApiHost   string
	LogLevel  string
	LogFormat string
}

// Envs is a global variable that holds the loaded environment configuration.
var Envs = GetEnvs()

// GetEnvs loads environment variables and returns a ConfigEnv struct.
func GetEnvs() ConfigEnv {
	godotenv.Load()
	userID, _ := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		logFormat = "text"
	}

	return ConfigEnv{
		BotToken:  os.Getenv("BOT_TK"),
		UserID:    userID,
		ApiHost:   os.Getenv("API_HOST"),
		LogLevel:  logLevel,
		LogFormat: logFormat,
	}
}
