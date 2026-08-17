package bot

import (
	"sync"
	"telegram-bot/internal/ai"
	"testing"
)

func TestUserSessionModel(t *testing.T) {
	userID := int64(123456)

	// Inicialmente deve retornar o modelo default
	defaultModel := GetUserModel(userID)
	if defaultModel != ai.DefaultModel {
		t.Errorf("expected %s, got %s", ai.DefaultModel, defaultModel)
	}

	// Atualiza o modelo
	SetUserModel(userID, "deepseek-r1:1.5b")
	updatedModel := GetUserModel(userID)
	if updatedModel != "deepseek-r1:1.5b" {
		t.Errorf("expected deepseek-r1:1.5b, got %s", updatedModel)
	}

	// Outro usuário deve receber default se não configurado
	otherUserID := int64(987654)
	if GetUserModel(otherUserID) != ai.DefaultModel {
		t.Errorf("expected %s for other user, got %s", ai.DefaultModel, GetUserModel(otherUserID))
	}
}

func TestConcurrentSessionAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			SetUserModel(id, "qwen3.5:0.8b")
			model := GetUserModel(id)
			if model != "qwen3.5:0.8b" {
				t.Errorf("concurrent mismatch for user %d: got %s", id, model)
			}
		}(int64(i))
	}
	wg.Wait()
}
