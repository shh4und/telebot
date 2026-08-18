package bot

import (
	"testing"
	"time"
)

func TestCooldownTracker(t *testing.T) {
	cd := NewCooldownTracker(100 * time.Millisecond)
	userID := int64(12345)

	// 1ª chamada: deve ser permitida
	allowed, remaining := cd.Allow(userID)
	if !allowed || remaining > 0 {
		t.Fatalf("esperava Allow=true e remaining=0 na 1ª chamada, obteve %v e %v", allowed, remaining)
	}

	// 2ª chamada imediata: deve ser bloqueada
	allowed, remaining = cd.Allow(userID)
	if allowed || remaining <= 0 {
		t.Fatalf("esperava Allow=false e remaining > 0 na 2ª chamada imediata, obteve %v e %v", allowed, remaining)
	}

	// Outro usuário: deve ser permitido imediatamente
	otherAllowed, _ := cd.Allow(int64(99999))
	if !otherAllowed {
		t.Fatalf("esperava Allow=true para outro usuário")
	}

	// Espera o cooldown expirar
	time.Sleep(110 * time.Millisecond)

	// 3ª chamada após o tempo: deve ser permitida
	allowed, remaining = cd.Allow(userID)
	if !allowed || remaining > 0 {
		t.Fatalf("esperava Allow=true após expiração do cooldown, obteve %v e %v", allowed, remaining)
	}
}
