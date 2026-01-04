package config

import (
	"os"

	"github.com/joho/godotenv"
)

// ConfigEnv holds the configuration values for the application.
type ConfigEnv struct {
	BotToken string
}

// Envs is a global variable that holds the loaded environment configuration.
var Envs = GetEnvs()

// GetEnvs loads environment variables and returns a ConfigEnv struct.
func GetEnvs() ConfigEnv {
	godotenv.Load()

	return ConfigEnv{
		BotToken: os.Getenv("BOT_TK"),
	}
}
