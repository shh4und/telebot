package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ConfigEnv holds the configuration values for the application.
type ConfigEnv struct {
	BotToken string
	UserID   int64
	ApiHost  string
}

// Envs is a global variable that holds the loaded environment configuration.
var Envs = GetEnvs()

// GetEnvs loads environment variables and returns a ConfigEnv struct.
func GetEnvs() ConfigEnv {
	godotenv.Load()
	userID, _ := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)

	return ConfigEnv{
		BotToken: os.Getenv("BOT_TK"),
		UserID:   userID,
		ApiHost:  os.Getenv("API_HOST"),
	}
}
