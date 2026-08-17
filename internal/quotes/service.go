package quotes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	client   *http.Client
	cacheTTL time.Duration
	cache    *MarketSummary
	cacheAt  time.Time
	mu       sync.RWMutex
}

var defaultService = NewService(1 * time.Minute)

// NewService cria um serviço de cotações com TTL configurável.
func NewService(ttl time.Duration) *Service {
	return &Service{
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		cacheTTL: ttl,
	}
}

// AwesomeAPI response structs
type awesomeAPIResponse struct {
	USDBRL awesomeItem `json:"USDBRL"`
	EURBRL awesomeItem `json:"EURBRL"`
}

type awesomeItem struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	CreateDate string `json:"create_date"`
}

// CoinGecko response struct
type coinGeckoResponse map[string]struct {
	USD          float64 `json:"usd"`
	USD24hChange float64 `json:"usd_24h_change"`
	BRL          float64 `json:"brl"`
	BRL24hChange float64 `json:"brl_24h_change"`
}

// FetchFiatQuotes busca cotações de Dólar e Euro na AwesomeAPI.
func (s *Service) FetchFiatQuotes(ctx context.Context) (dollar CurrencyQuote, euro CurrencyQuote, err error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL,EUR-BRL"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return dollar, euro, err
	}
	req.Header.Set("User-Agent", "TelegramBot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return dollar, euro, fmt.Errorf("falha ao consultar AwesomeAPI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dollar, euro, fmt.Errorf("AwesomeAPI retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dollar, euro, err
	}

	var data awesomeAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return dollar, euro, fmt.Errorf("falha no parse JSON de moedas: %w", err)
	}

	dollar = parseAwesomeItem(data.USDBRL)
	euro = parseAwesomeItem(data.EURBRL)
	return dollar, euro, nil
}

func parseAwesomeItem(item awesomeItem) CurrencyQuote {
	bid, _ := strconv.ParseFloat(item.Bid, 64)
	ask, _ := strconv.ParseFloat(item.Ask, 64)
	high, _ := strconv.ParseFloat(item.High, 64)
	low, _ := strconv.ParseFloat(item.Low, 64)
	pct, _ := strconv.ParseFloat(item.PctChange, 64)

	t, err := time.Parse("2006-01-02 15:04:05", item.CreateDate)
	if err != nil {
		t = time.Now()
	}

	return CurrencyQuote{
		Code:      item.Code,
		Name:      item.Name,
		Bid:       bid,
		Ask:       ask,
		High:      high,
		Low:       low,
		PctChange: pct,
		UpdatedAt: t,
	}
}

// FetchCryptoQuotes busca cotações de Bitcoin, Ethereum e Solana no CoinGecko.
func (s *Service) FetchCryptoQuotes(ctx context.Context) ([]CryptoQuote, error) {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum,solana&vs_currencies=usd,brl&include_24hr_change=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TelegramBot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar CoinGecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data coinGeckoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("falha no parse JSON de criptos: %w", err)
	}

	orderedKeys := []struct {
		id     string
		symbol string
		name   string
	}{
		{id: "bitcoin", symbol: "BTC", name: "Bitcoin"},
		{id: "ethereum", symbol: "ETH", name: "Ethereum"},
		{id: "solana", symbol: "SOL", name: "Solana"},
	}

	var quotes []CryptoQuote
	for _, meta := range orderedKeys {
		if val, ok := data[meta.id]; ok {
			quotes = append(quotes, CryptoQuote{
				ID:        meta.id,
				Symbol:    meta.symbol,
				Name:      meta.name,
				PriceBRL:  val.BRL,
				PriceUSD:  val.USD,
				Change24h: val.USD24hChange,
			})
		}
	}

	return quotes, nil
}

// GetMarketSummary obtém o resumo do mercado com cache.
func (s *Service) GetMarketSummary(ctx context.Context, forceRefresh bool) (*MarketSummary, error) {
	s.mu.RLock()
	cached := s.cache
	cachedAt := s.cacheAt
	s.mu.RUnlock()

	if !forceRefresh && cached != nil && time.Since(cachedAt) < s.cacheTTL {
		return cached, nil
	}

	var summary MarketSummary
	var dollarErr, cryptoErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		var d, e CurrencyQuote
		d, e, dollarErr = s.FetchFiatQuotes(ctx)
		if dollarErr == nil {
			summary.Dollar = d
			summary.Euro = e
		}
	}()

	go func() {
		defer wg.Done()
		var cryptos []CryptoQuote
		cryptos, cryptoErr = s.FetchCryptoQuotes(ctx)
		if cryptoErr == nil {
			summary.Cryptos = cryptos
		}
	}()

	wg.Wait()

	if dollarErr != nil && cryptoErr != nil {
		if cached != nil {
			slog.Warn("usando cache antigo de cotações devido a erro nas APIs")
			return cached, nil
		}
		return nil, fmt.Errorf("erro ao obter cotações: moedas (%v), cripto (%v)", dollarErr, cryptoErr)
	}

	summary.UpdatedAt = time.Now()

	s.mu.Lock()
	s.cache = &summary
	s.cacheAt = time.Now()
	s.mu.Unlock()

	return &summary, nil
}

// GetMarketSummary chama o serviço padrão.
func GetMarketSummary(ctx context.Context, forceRefresh bool) (*MarketSummary, error) {
	return defaultService.GetMarketSummary(ctx, forceRefresh)
}

// FormatCurrency formata valores monetários em padrão amigável BRL / USD.
func FormatBRL(val float64) string {
	if val >= 1000 {
		return formatNumber(val, 2, ".", ",")
	}
	return formatNumber(val, 2, ".", ",")
}

func FormatUSD(val float64) string {
	return formatNumber(val, 2, ",", ".")
}

func formatNumber(val float64, decimals int, thousandsSep, decimalSep string) string {
	parts := strings.Split(fmt.Sprintf("%.*f", decimals, val), ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = parts[1]
	}

	var result []string
	for i := len(intPart); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		result = append([]string{intPart[start:i]}, result...)
	}

	formattedInt := strings.Join(result, thousandsSep)
	if decPart != "" {
		return formattedInt + decimalSep + decPart
	}
	return formattedInt
}

func VariationIndicator(pct float64) string {
	if pct > 0 {
		return fmt.Sprintf("🟢 +%.2f%%", pct)
	} else if pct < 0 {
		return fmt.Sprintf("🔴 %.2f%%", pct)
	}
	return "⚪ 0.00%"
}
