package news

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

type cacheEntry struct {
	articles  []Article
	fetchedAt time.Time
}

// Service gerencia a obtenção e cache de notícias.
type Service struct {
	cacheTTL time.Duration
	cache    map[string]cacheEntry
	mu       sync.RWMutex
}

var defaultService = NewService(5 * time.Minute)

// NewService cria uma nova instância de Service com TTL customizado.
func NewService(ttl time.Duration) *Service {
	return &Service{
		cacheTTL: ttl,
		cache:    make(map[string]cacheEntry),
	}
}

// GetNews obtém até `limit` notícias para a categoria especificada.
func (s *Service) GetNews(ctx context.Context, categoryKey string, limit int) ([]Article, *CategoryInfo, error) {
	catKey := NormalizeCategory(categoryKey)
	if catKey == "" {
		catKey = "br" // default para Brasil
	}

	info, ok := DefaultCategories[catKey]
	if !ok {
		return nil, nil, fmt.Errorf("categoria desconhecida: %s", categoryKey)
	}

	// Verifica cache
	s.mu.RLock()
	entry, found := s.cache[catKey]
	s.mu.RUnlock()

	if found && time.Since(entry.fetchedAt) < s.cacheTTL {
		slog.Debug("retornando notícias do cache", "categoria", catKey, "total", len(entry.articles))
		return limitArticles(entry.articles, limit), &info, nil
	}

	// Busca de todas as fontes da categoria
	var allArticles []Article
	var wg sync.WaitGroup
	var feedMu sync.Mutex

	for _, source := range info.FeedURLs {
		wg.Add(1)
		go func(src FeedSource) {
			defer wg.Done()
			articles, err := FetchFeed(ctx, src, catKey)
			if err != nil {
				slog.Warn("falha ao buscar feed RSS", "fonte", src.Source, "url", src.URL, "error", err)
				return
			}

			feedMu.Lock()
			allArticles = append(allArticles, articles...)
			feedMu.Unlock()
		}(source)
	}

	wg.Wait()

	if len(allArticles) == 0 {
		// Se falhou e tem cache antigo, usa cache como fallback
		if found && len(entry.articles) > 0 {
			slog.Warn("usando cache antigo após falha de rede", "categoria", catKey)
			return limitArticles(entry.articles, limit), &info, nil
		}
		return nil, &info, fmt.Errorf("não foi possível carregar notícias no momento")
	}

	// Ordena por data decrescente
	sort.Slice(allArticles, func(i, j int) bool {
		return allArticles[i].PublishedAt.After(allArticles[j].PublishedAt)
	})

	// Salva no cache
	s.mu.Lock()
	s.cache[catKey] = cacheEntry{
		articles:  allArticles,
		fetchedAt: time.Now(),
	}
	s.mu.Unlock()

	return limitArticles(allArticles, limit), &info, nil
}

func limitArticles(articles []Article, limit int) []Article {
	if limit <= 0 || limit > len(articles) {
		return articles
	}
	return articles[:limit]
}

// GetNews usa a instância padrão do serviço de notícias (limite padrão de 5 notícias).
func GetNews(ctx context.Context, categoryKey string) ([]Article, *CategoryInfo, error) {
	return defaultService.GetNews(ctx, categoryKey, 5)
}
