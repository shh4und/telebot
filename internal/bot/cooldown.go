package bot

import (
	"sync"
	"time"
)

// CooldownTracker gerencia intervalos mínimos entre ações por ID de usuário em memória, sem dependências externas.
type CooldownTracker struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastCall map[int64]time.Time
}

// NewCooldownTracker cria um novo rastreador de cooldown com a duração especificada.
func NewCooldownTracker(cooldown time.Duration) *CooldownTracker {
	return &CooldownTracker{
		cooldown: cooldown,
		lastCall: make(map[int64]time.Time),
	}
}

// Allow verifica se a ação é permitida para o userID.
// Retorna:
// - allowed: true se a ação puder ser executada, false se ainda estiver em cooldown.
// - remaining: o tempo restante para o fim do cooldown (0 se permitido).
func (c *CooldownTracker) Allow(userID int64) (allowed bool, remaining time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Limpeza esparsa para evitar crescimento indefinido do mapa em processos longos
	if len(c.lastCall) > 2000 {
		for id, t := range c.lastCall {
			if now.Sub(t) > c.cooldown*2 {
				delete(c.lastCall, id)
			}
		}
	}

	if last, exists := c.lastCall[userID]; exists {
		elapsed := now.Sub(last)
		if elapsed < c.cooldown {
			return false, c.cooldown - elapsed
		}
	}

	c.lastCall[userID] = now
	return true, 0
}
