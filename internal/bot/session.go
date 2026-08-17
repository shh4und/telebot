package bot

import (
	"sync"
	"telegram-bot/internal/ai"
)

var (
	userModels   sync.Map
)

// GetUserModel retorna o modelo de IA selecionado pelo usuário.
// Se nenhum modelo tiver sido selecionado, retorna ai.DefaultModel.
func GetUserModel(userID int64) string {
	if val, ok := userModels.Load(userID); ok {
		if model, ok := val.(string); ok && model != "" {
			return model
		}
	}
	return ai.DefaultModel
}

// SetUserModel define o modelo de IA preferido para um usuário.
func SetUserModel(userID int64, model string) {
	userModels.Store(userID, model)
}
