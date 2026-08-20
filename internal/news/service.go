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

// PageResult contém as notícias da página atual e metadados de paginação.
type PageResult struct {
	Articles    []Article
	Category    *CategoryInfo
	CurrentPage int
	TotalPages  int
	TotalItems  int
	FetchedAt   time.Time
}

// NewService cria uma nova instância de Service com TTL customizado.
func NewService(ttl time.Duration) *Service {
	return &Service{
		cacheTTL: ttl,
		cache:    make(map[string]cacheEntry),
	}
}

// InvalidateCache remove a categoria do cache para forçar uma nova busca.
func (s *Service) InvalidateCache(categoryKey string) {
	catKey := NormalizeCategory(categoryKey)
	if catKey == "" {
		catKey = "br"
	}
	s.mu.Lock()
	delete(s.cache, catKey)
	s.mu.Unlock()
}

// GetPagedNews obtém as notícias paginadas com opção de forçar atualização.
func (s *Service) GetPagedNews(ctx context.Context, categoryKey string, page, pageSize int, forceRefresh bool) (*PageResult, error) {
	catKey := NormalizeCategory(categoryKey)
	if catKey == "" {
		catKey = "br"
	}

	info, ok := DefaultCategories[catKey]
	if !ok {
		return nil, fmt.Errorf("categoria desconhecida: %s", categoryKey)
	}

	if pageSize <= 0 {
		pageSize = 2
	}
	if page <= 0 {
		page = 1
	}

	if forceRefresh {
		s.InvalidateCache(catKey)
	}

	var allArticles []Article
	var fetchedAt time.Time

	s.mu.RLock()
	entry, found := s.cache[catKey]
	s.mu.RUnlock()

	if !forceRefresh && found && time.Since(entry.fetchedAt) < s.cacheTTL {
		slog.Debug("retornando notícias do cache", "categoria", catKey, "total", len(entry.articles))
		allArticles = entry.articles
		fetchedAt = entry.fetchedAt
	} else {
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
			if found && len(entry.articles) > 0 {
				slog.Warn("usando cache antigo após falha de rede", "categoria", catKey)
				allArticles = entry.articles
				fetchedAt = entry.fetchedAt
			} else {
				return nil, fmt.Errorf("não foi possível carregar notícias no momento")
			}
		} else {
			sort.Slice(allArticles, func(i, j int) bool {
				return allArticles[i].PublishedAt.After(allArticles[j].PublishedAt)
			})

			fetchedAt = time.Now()
			s.mu.Lock()
			s.cache[catKey] = cacheEntry{
				articles:  allArticles,
				fetchedAt: fetchedAt,
			}
			s.mu.Unlock()
		}
	}

	totalItems := len(allArticles)
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex > totalItems {
		startIndex = totalItems
	}
	if endIndex > totalItems {
		endIndex = totalItems
	}

	pagedArticles := allArticles[startIndex:endIndex]

	return &PageResult{
		Articles:    pagedArticles,
		Category:    &info,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		FetchedAt:   fetchedAt,
	}, nil
}

// GetNews obtém até `limit` notícias para a categoria especificada (compatibilidade).
func (s *Service) GetNews(ctx context.Context, categoryKey string, limit int) ([]Article, *CategoryInfo, error) {
	res, err := s.GetPagedNews(ctx, categoryKey, 1, limit, false)
	if err != nil {
		catKey := NormalizeCategory(categoryKey)
		info := DefaultCategories[catKey]
		return nil, &info, err
	}
	return res.Articles, res.Category, nil
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

// GetPagedNews usa a instância padrão do serviço para obter notícias paginadas.
func GetPagedNews(ctx context.Context, categoryKey string, page, pageSize int, forceRefresh bool) (*PageResult, error) {
	return defaultService.GetPagedNews(ctx, categoryKey, page, pageSize, forceRefresh)
}

// FormatRelativeTime retorna uma representação amigável do tempo decorrido.
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	diff := time.Since(t)
	if diff < 0 || diff < time.Minute {
		return "agora mesmo"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("há %d min", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "há 1h"
		}
		return fmt.Sprintf("há %dh", hours)
	}
	if diff < 48*time.Hour {
		return "ontem"
	}
	days := int(diff.Hours() / 24)
	return fmt.Sprintf("há %d dias", days)
}

